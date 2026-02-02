.PHONY: build build-server build-toolbox run dev keygen migrate test clean

# Build all
build: build-server build-toolbox

# Build server
build-server:
	go build -o bin/server cmd/server/main.go

# Build toolbox CLI
build-toolbox:
	go build -o bin/nf-toolbox cmd/toolbox/main.go

# Run server
run: build-server
	./bin/server

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Generate ed25519 key pair
keygen:
	go run cmd/keygen/main.go

# Run database migrations
migrate:
	psql -d license_db < migrations/001_initial.sql

# Create database (run once)
db-create:
	createdb license_db

# Full setup (run once)
setup: db-create migrate keygen
	@echo "Setup complete. Now:"
	@echo "1. Add generated keys to .env"
	@echo "2. Run: make run"

# Test
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf tmp/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Deploy to Fly.io
deploy:
	fly deploy

# View Fly.io logs
logs:
	fly logs

# Install toolbox locally
install-toolbox: build-toolbox
	mkdir -p ~/.nf-tools/bin
	cp bin/nf-toolbox ~/.nf-tools/bin/
	@echo "Installed to ~/.nf-tools/bin/nf-toolbox"
	@echo "Add to PATH: export PATH=\"~/.nf-tools/bin:\$$PATH\""

# Build for all platforms
release:
	GOOS=darwin GOARCH=amd64 go build -o dist/nf-toolbox-darwin-amd64 cmd/toolbox/main.go
	GOOS=darwin GOARCH=arm64 go build -o dist/nf-toolbox-darwin-arm64 cmd/toolbox/main.go
	GOOS=linux GOARCH=amd64 go build -o dist/nf-toolbox-linux-amd64 cmd/toolbox/main.go
	GOOS=linux GOARCH=arm64 go build -o dist/nf-toolbox-linux-arm64 cmd/toolbox/main.go
	GOOS=windows GOARCH=amd64 go build -o dist/nf-toolbox-windows-amd64.exe cmd/toolbox/main.go
