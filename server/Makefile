
set :
	go mod tidy

run :
	go mod tidy
	go run ./cmd/main.go

migrate :
	@if [ -z "$(DATABASE_DSN)" ]; then \
		export $$(grep -v '^#' .env | xargs) && psql "$$DATABASE_DSN" -f migrations/001_init.sql; \
	else \
		psql "$(DATABASE_DSN)" -f migrations/001_init.sql; \
	fi

build :
	go build -o bin/server ./cmd/main.go

test :
	go test ./tests/...
