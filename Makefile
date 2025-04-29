include .env

#=======================================================================================#
# HELPERS
#=======================================================================================#

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

#=======================================================================================#
# DEVELOPMENT
#=======================================================================================#

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-user=${GREENLIGHT_DB_USERNAME} -db-pwd=${GREENLIGHT_DB_PASSWORD} -db-host=${GREENLIGHT_DB_HOST}  -db-port=${GREENLIGHT_DB_PORT} -db-name=${GREENLIGHT_DB_NAME}

## db/psql: connect to the PostgreSQL using env variable 'GREENLIGHT_DB_DSN'
.PHONY: db/psql
db/psql:
	@psql ${GREENLIGHT_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	@migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

#=======================================================================================#
# QUALITY CONTROL
#=======================================================================================#

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@echo 'Formating files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy
	@echo 'Verifying and vendoring module dependencies...'
	go mod verify
	go mod vendor

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies...'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

#=======================================================================================#
# BUILD
#=======================================================================================#

## build/api: build the cmd/api linux/amd64 application
.PHONY: build/api/linux
build/api/linux:
	@echo 'Building cmd/api...'
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

## build/api: build the cmd/api darwin/arm64 application
.PHONY: build/api/mac
build/api/mac:
	@echo 'Building cmd/api...'
	GOOS=darwin GOARCH=arm64 go build -ldflags='-s' -o=./bin/darwin_arm64/api ./cmd/api
