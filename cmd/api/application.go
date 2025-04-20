package main

import (
	"context"
	"database/sql"
	"log/slog"
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
}

type rateLimiter struct {
	cfg     rateLimiterCfg
	mu      sync.Mutex
	clients map[string]*client
}

type rateLimiterCfg struct {
	rateLimitEnabled bool
	ratePerSecond    float64
	rateBurst        int
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(cfg rateLimiterCfg) *rateLimiter {
	rl := &rateLimiter{
		cfg:     cfg,
		clients: make(map[string]*client),
	}
	go rl.inactiveClientsCleanup()
	return rl
}

func (r *rateLimiter) inactiveClientsCleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		r.mu.Lock()
		for ip, c := range r.clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(r.clients, ip)
			}
		}
		r.mu.Unlock()
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

func (r *rateLimiter) getRPS() rate.Limit {
	return rate.Limit(r.cfg.ratePerSecond)
}

func (r *rateLimiter) getBurst() int {
	return r.cfg.rateBurst
}

func (r *rateLimiter) IsEnabled() bool {
	return r.cfg.rateLimitEnabled
}
