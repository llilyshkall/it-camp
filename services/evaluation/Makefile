.PHONY: sqlc-generate sqlc-validate sqlc-diff

# Generate Go code from SQL queries
sqlc-generate:
	sqlc generate

# Validate SQL queries
sqlc-validate:
	sqlc compile

# Show diff between current and generated code
sqlc-diff:
	sqlc diff

# Install sqlc if not present
sqlc-install:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Clean generated files
clean:
	rm -rf internal/database 