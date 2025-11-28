# Rockstar Web Framework Makefile

.PHONY: build build-no-plugins test test-plugins clean fmt vet lint deps check help plugin-list discover-plugins

# Discover all plugins in the plugins/ directory
PLUGINS := $(wildcard plugins/*/plugin.go)
PLUGIN_DIRS := $(dir $(PLUGINS))
PLUGIN_NAMES := $(foreach dir,$(PLUGIN_DIRS),$(notdir $(patsubst %/,%,$(dir))))

# Build the framework with all plugins
build: discover-plugins
	@echo "Building Rockstar Web Framework with plugins..."
	@if [ -n "$(PLUGIN_NAMES)" ]; then \
		echo "Plugins found: $(PLUGIN_NAMES)"; \
	else \
		echo "No plugins found in plugins/ directory"; \
	fi
	go build -o rockstar ./cmd/rockstar
	@echo "✓ Build complete: ./rockstar"

# Build without plugins (using build tag)
build-no-plugins:
	@echo "Building Rockstar Web Framework without plugins..."
	go build -tags noplugins -o rockstar ./cmd/rockstar
	@echo "✓ Build complete (no plugins): ./rockstar"

# Discover plugins and update go.mod with replacements
discover-plugins:
	@echo "Discovering plugins..."
	@for dir in $(PLUGIN_DIRS); do \
		plugin_name=$$(basename $$dir); \
		plugin_path="github.com/echterhof/rockstar-web-framework/plugins/$$plugin_name"; \
		if ! grep -q "replace $$plugin_path" go.mod; then \
			echo "Adding plugin to go.mod: $$plugin_name"; \
			go mod edit -replace $$plugin_path=./plugins/$$plugin_name; \
		fi; \
	done
	@echo "Generating plugin imports..."
	@if [ -f scripts/generate-plugin-imports.sh ]; then \
		chmod +x scripts/generate-plugin-imports.sh; \
		./scripts/generate-plugin-imports.sh; \
	elif [ -f scripts/generate-plugin-imports.ps1 ]; then \
		powershell -ExecutionPolicy Bypass -File scripts/generate-plugin-imports.ps1; \
	fi
	@go mod tidy

# List all registered plugins
plugin-list:
	@echo "Registered plugins:"
	@go run ./cmd/rockstar --list-plugins

# Run tests
test:
	go test ./...

# Run plugin tests
test-plugins:
	@echo "Running plugin tests..."
	@for dir in $(PLUGIN_DIRS); do \
		echo "Testing $$(basename $$dir)..."; \
		cd $$dir && go test ./... && cd ../..; \
	done

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	go clean ./...
	rm -f coverage.out coverage.html
	rm -f rockstar rockstar.exe
	@for dir in $(PLUGIN_DIRS); do \
		cd $$dir && go clean ./... && cd ../..; \
	done

# Format code
fmt:
	go fmt ./...
	@for dir in $(PLUGIN_DIRS); do \
		cd $$dir && go fmt ./... && cd ../..; \
	done

# Run go vet
vet:
	go vet ./...
	@for dir in $(PLUGIN_DIRS); do \
		cd $$dir && go vet ./... && cd ../..; \
	done

# Download dependencies
deps:
	go mod download
	go mod tidy
	@for dir in $(PLUGIN_DIRS); do \
		cd $$dir && go mod download && go mod tidy && cd ../..; \
	done

# Run all checks
check: fmt vet test test-plugins

# Help
help:
	@echo "Available targets:"
	@echo "  build              - Build the framework with all plugins"
	@echo "  build-no-plugins   - Build the framework without plugins"
	@echo "  test               - Run framework tests"
	@echo "  test-plugins       - Run plugin tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  clean              - Clean build artifacts"
	@echo "  fmt                - Format code (framework and plugins)"
	@echo "  vet                - Run go vet (framework and plugins)"
	@echo "  deps               - Download and tidy dependencies"
	@echo "  check              - Run fmt, vet, test, and test-plugins"
	@echo "  plugin-list        - List all registered plugins"
	@echo "  discover-plugins   - Discover plugins and update go.mod"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "Plugins found: $(PLUGIN_NAMES)"