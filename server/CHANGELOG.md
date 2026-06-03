# Changelog

---

## 04-06-2026

### internal/lib/database.go
feat: migrate database driver from MySQL to PostgreSQL with pgvector support

### internal/config/config.go
feat: update DATABASE_DSN default format to PostgreSQL connection string

### migrations/001_init.sql
feat: add initial PostgreSQL migration with full schema and pgvector extension

### Makefile
feat: add migrate, build, and test commands
