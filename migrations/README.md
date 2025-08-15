# Database Migrations

This directory contains database migration files for the lumi-go service.

## Migration Tool

We use [golang-migrate](https://github.com/golang-migrate/migrate) for managing database migrations.

## Installation

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Go install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## File Naming Convention

Migration files follow this naming pattern:
```
{version}_{description}.{direction}.sql
```

- `version`: 6-digit sequential number (e.g., 000001)
- `description`: Snake_case description of the migration
- `direction`: Either `up` or `down`

Example:
- `000001_init_schema.up.sql` - Creates initial schema
- `000001_init_schema.down.sql` - Rolls back initial schema

## Creating New Migrations

```bash
# Create a new migration pair
migrate create -ext sql -dir migrations -seq create_users_table

# This creates:
# - 000002_create_users_table.up.sql
# - 000002_create_users_table.down.sql
```

## Running Migrations

### Using migrate CLI

```bash
# Run all up migrations
migrate -database "postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable" \
        -path migrations up

# Run next N up migrations
migrate -database "postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable" \
        -path migrations up 2

# Rollback last migration
migrate -database "postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable" \
        -path migrations down 1

# Force version (use with caution!)
migrate -database "postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable" \
        -path migrations force 1

# Check current version
migrate -database "postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable" \
        -path migrations version
```

### Using Docker Compose

```bash
# Run migrations using the migrate service
docker-compose run --rm migrate

# Or if using profiles
docker-compose --profile migration up migrate
```

### Using Makefile

```bash
# Run up migrations
make migrate-up

# Rollback last migration
make migrate-down

# Reset database (down all, then up all)
make migrate-reset

# Create new migration
make migrate-create name=add_products_table
```

### Programmatically in Go

```go
import (
    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(dbURL string) error {
    m, err := migrate.New(
        "file://migrations",
        dbURL,
    )
    if err != nil {
        return err
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    return nil
}
```

## Best Practices

### DO:
- ✅ Always write both up and down migrations
- ✅ Test down migrations to ensure they work
- ✅ Use transactions when possible
- ✅ Add indexes for foreign keys and frequently queried columns
- ✅ Include comments for complex operations
- ✅ Use IF EXISTS/IF NOT EXISTS clauses for idempotency
- ✅ Version control all migration files
- ✅ Run migrations in CI/CD pipeline

### DON'T:
- ❌ Modify existing migration files (create new ones instead)
- ❌ Use migration files for data seeding in production
- ❌ Skip versions in the sequence
- ❌ Mix DDL and DML in the same migration without careful consideration
- ❌ Drop columns or tables without considering backward compatibility
- ❌ Use migrations for large data transformations

## Migration Examples

### Creating a table
```sql
-- up
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- down
DROP TABLE IF EXISTS products;
```

### Adding a column
```sql
-- up
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

-- down
ALTER TABLE users DROP COLUMN phone;
```

### Creating an index
```sql
-- up
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- down
DROP INDEX CONCURRENTLY IF EXISTS idx_users_email;
```

### Adding a foreign key
```sql
-- up
ALTER TABLE orders 
ADD CONSTRAINT fk_orders_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- down
ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_user_id;
```

## Troubleshooting

### Migration is stuck

Check for locks:
```sql
SELECT * FROM pg_locks WHERE NOT granted;
```

### Dirty database version

If a migration fails partway:
```bash
# Check current version
migrate -database $DATABASE_URL -path migrations version

# Force to previous version
migrate -database $DATABASE_URL -path migrations force {previous_version}

# Then try again
migrate -database $DATABASE_URL -path migrations up
```

### Connection issues

Ensure the database is accessible:
```bash
psql "postgres://lumigo:lumigo@localhost:5432/lumigo?sslmode=disable" -c "SELECT 1"
```

## Schema Documentation

Current schema includes:
- `users` - User accounts and authentication
- `sessions` - Active user sessions and refresh tokens
- `audit_logs` - Audit trail for all system actions
- `feature_flags` - Feature flag configuration
- `api_keys` - API keys for programmatic access

See `000001_init_schema.up.sql` for the complete initial schema definition.

## Environment Variables

- `DATABASE_URL` - Full PostgreSQL connection string
- `MIGRATION_DIR` - Path to migrations directory (default: `./migrations`)
- `MIGRATION_TIMEOUT` - Timeout for each migration (default: `60s`)
