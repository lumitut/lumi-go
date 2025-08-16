# Database Migrations

This directory contains example database migration files. Migrations are **optional** and only needed if your service uses a database.

## Overview

Database migrations help you:
- Version control your database schema
- Apply incremental changes safely
- Roll back changes if needed
- Keep development, staging, and production databases in sync

## Migration Tools

Choose a migration tool that fits your needs:

### 1. golang-migrate
Most popular and recommended.

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration
migrate create -ext sql -dir migrations -seq create_users_table

# Run migrations
migrate -path migrations -database "postgres://user:pass@localhost:5432/db?sslmode=disable" up

# Rollback
migrate -path migrations -database "postgres://user:pass@localhost:5432/db?sslmode=disable" down 1
```

### 2. Goose
Simple and effective.

```bash
# Install
go install github.com/pressly/goose/v3/cmd/goose@latest

# Create migration
goose -dir migrations create add_users_table sql

# Run migrations
goose -dir migrations postgres "user=postgres dbname=mydb sslmode=disable" up

# Rollback
goose -dir migrations postgres "user=postgres dbname=mydb sslmode=disable" down
```

### 3. Atlas
Modern with schema-as-code support.

```bash
# Install
curl -sSf https://atlasgo.sh | sh

# Create migration
atlas migrate new create_users_table

# Run migrations
atlas migrate apply --url "postgres://localhost:5432/mydb"
```

## File Naming Convention

Use sequential numbering with descriptive names:
```
000001_init_schema.up.sql
000001_init_schema.down.sql
000002_add_users_table.up.sql
000002_add_users_table.down.sql
000003_add_indexes.up.sql
000003_add_indexes.down.sql
```

## Example Migrations

### Creating a Table (up migration)
```sql
-- 000002_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

### Dropping a Table (down migration)
```sql
-- 000002_create_users_table.down.sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

## Using with the Service

### 1. Configure Database Connection
```bash
export LUMI_CLIENTS_DATABASE_ENABLED=true
export LUMI_CLIENTS_DATABASE_URL=postgres://user:pass@localhost:5432/mydb
```

### 2. Run Migrations on Startup (Optional)
```go
// In your main.go or app initialization
import "github.com/golang-migrate/migrate/v4"
import _ "github.com/golang-migrate/migrate/v4/database/postgres"
import _ "github.com/golang-migrate/migrate/v4/source/file"

func runMigrations(databaseURL string) error {
    m, err := migrate.New(
        "file://migrations",
        databaseURL,
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

### 3. Using Docker
```dockerfile
# In your Dockerfile
COPY migrations /migrations

# Run migrations as part of startup
CMD ["sh", "-c", "migrate -path /migrations -database $DATABASE_URL up && ./server"]
```

## Best Practices

1. **Always test migrations** in development first
2. **Never modify existing migrations** - create new ones instead
3. **Keep migrations small** and focused on one change
4. **Include rollback migrations** (down files)
5. **Use transactions** when possible
6. **Document complex migrations** with comments
7. **Backup before running** migrations in production

## Migration Examples

The `.example` files in this directory show:
- Creating tables with proper types
- Adding indexes for performance
- Setting up foreign key relationships
- Creating trigger functions
- Inserting seed data

To use them:
```bash
# Copy example files
cp 000001_init_schema.up.sql.example 000001_init_schema.up.sql
cp 000001_init_schema.down.sql.example 000001_init_schema.down.sql

# Modify for your needs
vim 000001_init_schema.up.sql

# Run migrations
migrate -path . -database $DATABASE_URL up
```

## Troubleshooting

### Migration Failed
- Check the migrations table for current version
- Review error messages carefully
- Test rollback in development
- Fix and create a new migration (don't modify failed one)

### Dirty Database State
```bash
# Force version
migrate -path migrations -database $DATABASE_URL force VERSION

# Then retry
migrate -path migrations -database $DATABASE_URL up
```

### Out of Sync
- Compare schema between environments
- Use schema dump to identify differences
- Create corrective migrations

## Resources

- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [Database migration best practices](https://www.prisma.io/dataguide/types/relational/migration-strategies)
- [PostgreSQL DDL documentation](https://www.postgresql.org/docs/current/ddl.html)
