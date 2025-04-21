package main

import (
	"context"
	"database/sql"
	"log/slog"
	"maps"
	"sync"
	"time"

	"github.com/shortykevich/greenlight/internal/data"
	"golang.org/x/time/rate"
)

type application struct {
	config  config
	logger  *slog.Logger
	models  data.Models
	limiter RateLimiter
}

func newApplication() *application {
	return &application{}
}

func (app *application) setConfig(cfg config) *application {
	app.config = cfg
	return app
}

func (app *application) setLogger(log *slog.Logger) *application {
	app.logger = log
	return app
}

func (app *application) setModels(models data.Models) *application {
	app.models = models
	return app
}

func (app *application) setRateLimiter(limiter RateLimiter) *application {
	app.limiter = limiter
	return app
}

func (app *application) openPostgresDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

type RateLimiter interface {
	Allow(ip string) bool
	IsEnabled() bool
	RunCleanup(ctx context.Context)
}

type rateLimiter struct {
	cfg          rateLimiterCfg
	mu           sync.Mutex
	clients      map[string]*client
	rebuilded_at time.Time
}

type rateLimiterCfg struct {
	enable bool
	rps    float64
	burst  int
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(cfg rateLimiterCfg) RateLimiter {
	rl := &rateLimiter{
		cfg:     cfg,
		clients: make(map[string]*client),
	}
	rl.rebuilded_at = time.Now()
	return rl
}

func (r *rateLimiter) RunCleanup(ctx context.Context) {
	clientsTicker := time.NewTicker(3 * time.Minute)
	rebuildTicker := time.NewTicker(6 * time.Hour)
	defer clientsTicker.Stop()
	defer rebuildTicker.Stop()

	for {
		select {
		case <-clientsTicker.C:
			r.cleanupInactive()
		case <-rebuildTicker.C:
			r.rebuildMap()
		case <-ctx.Done():
			return
		}
	}
}

func (r *rateLimiter) rebuildMap() {
	r.mu.Lock()
	defer r.mu.Unlock()

	newMap := make(map[string]*client, len(r.clients))
	maps.Copy(newMap, r.clients)
	r.clients = newMap
}

func (r *rateLimiter) cleanupInactive() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for ip, c := range r.clients {
		if time.Since(c.lastSeen) > 3*time.Minute {
			delete(r.clients, ip)
		}
	}
}

func (r *rateLimiter) Allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.clients[ip]
	if !ok {
		c = &client{
			limiter:  rate.NewLimiter(r.getRPS(), r.getBurst()),
			lastSeen: time.Now(),
		}
		r.clients[ip] = c
	}

	c.lastSeen = time.Now()

	return c.limiter.Allow()
}

func (r *rateLimiter) IsEnabled() bool {
	return r.cfg.enable
}

func (r *rateLimiter) ShutdownCtx(ctx context.Context) {

}

func (r *rateLimiter) getRPS() rate.Limit {
	return rate.Limit(r.cfg.rps)
}

func (r *rateLimiter) getBurst() int {
	return r.cfg.burst
}

type MockLimiter struct{}

func NewMockLimiter() RateLimiter {
	return &MockLimiter{}
}

func (m *MockLimiter) Allow(ip string) bool {
	return true
}

func (m *MockLimiter) IsEnabled() bool {
	return false
}

func (m *MockLimiter) RunCleanup(ctx context.Context) {}
