# Database Assets

This folder holds schema definitions and migration scripts for MangaHub.

## Suggested Layout

```
database/
├── migrations/          # Incremental SQL migrations
├── seeds/               # Optional seed data for local development
├── views/               # Materialized views or complex queries
└── database.sql         # Current baseline schema
```

- Keep each migration idempotent and backward compatible when possible.
- Name migration files with timestamps (e.g., `202502141200_create_tables.sql`).
- Place local-only helpers (like `.db` files) outside version control to keep the repository lean.