# Sange Hassan operations platform

This repository contains the public marketplace, the internal operations panel, a Go API, an operations worker, PostgreSQL migrations, and private document storage. The API is the authorization boundary; menu visibility in React is only a usability aid.

## Structure
- `back/` Go + Gin backend (clean architecture, JWT cookie auth)
- `front/website/` Public website (React + Tailwind)
- `front/panel/` Admin panel (React + Tailwind)
- `front/shared/` Shared i18n JSON + assets
- `deploy/` Docker compose, env files, nginx config, postgres init

## Dev (Docker)
```sh
cd deploy
docker compose -f docker-compose-dev.yml up --build
```

Services:
- Website: http://localhost:5173
- Admin panel: http://localhost:5174
- API: http://localhost:8080
- Images: http://localhost:8081/images/

The first internal administrator is created only from the `BOOTSTRAP_SUPER_ADMIN_*` environment variables. No fixed production password is seeded or documented.

## Prod (Docker)
Each production target has its own compose, env, and nginx config.

### prod-com (`sangehassan.com`, `91.239.211.60`)
```sh
cd deploy
cp .prod-com.env.example .prod-com.env
# edit DB_PASSWORD and JWT_SECRET before first start
docker compose --env-file .prod-com.env -f docker-compose-prod-com.yml up -d --build
```

### prod-ir (`sangehassan.ir`, `185.53.141.157`)
This target pulls Docker base images through `docker.abrha.net` and downloads Go modules through Liara's Go proxy.

```sh
cd deploy
cp .prod-ir.env.example .prod-ir.env
# edit DB_PASSWORD and JWT_SECRET before first start
docker compose --env-file .prod-ir.env -f docker-compose-prod-ir.yml up -d --build
```

Production routes on both targets:
- Website: `/`
- Admin panel: `/panel/`
- API: `/api/`
- Images: `/images/`

## Env files
Update secrets in the env file used by the target:
- Dev: `deploy/.dev.env`
- Legacy production compose: `deploy/.prod.env`
- prod-com: `deploy/.prod-com.env`
- prod-ir: `deploy/.prod-ir.env`

Image hosting base URL:
- In production this is intentionally blank by default so the frontend uses same-origin `/api` and `/images`.
- Set `VITE_API_BASE_URL` or `VITE_IMAGE_BASE_URL` only if API or images are served from a different origin.
- Set `VITE_SITE_URL` to the canonical public origin, for example `https://sangehassan.com`, so canonical and Open Graph URLs do not follow `www` or IP hosts.
- `npm run build` prerenders public static routes and writes `sitemap.xml` plus `robots.txt`. Set `VITE_PRERENDER_API_BASE_URL` only when a reachable API is available at build time and you also want product/block/project/ad detail pages prerendered.
- `VITE_SITEMAP_LASTMOD` is optional. Leave it blank to use the build date, or set a `YYYY-MM-DD` value for deterministic sitemap `lastmod` entries.
- `CATALOG_MIN_PRODUCTS` controls the minimum number of products required for a single-facet page to be indexable and included in the sitemap. The default is `2`.
- Keep `COOKIE_SECURE=true` for HTTPS production domains. For temporary direct HTTP/IP testing, cookies require `COOKIE_SECURE=false`.

## DB schema updates (existing volumes)
Postgres init SQL in `deploy/postgres/init/` only runs the first time a new volume is created. If you pull new code that adds tables (for example `content_sections` or `blocks`) and your existing DB volume was created before that, admin pages like `/dashboard/content` or `/dashboard/blocks` can fail with 500 errors.

To apply the latest schema to an existing DB volume without dropping data:
```sh
docker exec sangehassan-db psql -U sangehassan -d sangehassan -f /docker-entrypoint-initdb.d/001_init.sql
docker exec sangehassan-db psql -U sangehassan -d sangehassan -f /docker-entrypoint-initdb.d/002_auth.sql
docker exec sangehassan-db psql -U sangehassan -d sangehassan -f /docker-entrypoint-initdb.d/014_operations_phase1.sql
docker exec sangehassan-db psql -U sangehassan -d sangehassan -f /docker-entrypoint-initdb.d/015_operations_phase2.sql
docker exec sangehassan-db psql -U sangehassan -d sangehassan -f /docker-entrypoint-initdb.d/016_operations_phase3.sql
docker exec sangehassan-db psql -U sangehassan -d sangehassan -f /docker-entrypoint-initdb.d/017_finance_notifications_documents_reporting.sql
docker exec sangehassan-db psql -U sangehassan -d sangehassan -f /docker-entrypoint-initdb.d/018_supplier_purchase_quality_installation.sql
docker exec sangehassan-db psql -U sangehassan -d sangehassan -f /docker-entrypoint-initdb.d/019_application_settings_diagnostics_indexes.sql
docker exec sangehassan-db psql -U sangehassan -d sangehassan -f /docker-entrypoint-initdb.d/020_product_display_order.sql
```

Apply migrations in numeric order and take a database backup first. PostgreSQL init scripts do not migrate an existing volume automatically. The runtime readiness endpoint requires migration 20 to be registered. Moving an existing PostgreSQL 15 data directory to the PostgreSQL 16 image requires `pg_dump`/`pg_restore` or `pg_upgrade`; never attach a version-15 data directory directly to version 16.

## Operational dashboard bootstrap

Phase 1 uses the shared `users` table and database-backed roles and permissions. Configure the first protected administrator before starting the backend:

```sh
BOOTSTRAP_SUPER_ADMIN_PHONE=+989121234567
BOOTSTRAP_SUPER_ADMIN_PASSWORD=replace_with_a_strong_temporary_password
BOOTSTRAP_SUPER_ADMIN_FIRST_NAME=مدیر
BOOTSTRAP_SUPER_ADMIN_LAST_NAME=اصلی
```

Bootstrap runs only while no active Super Admin exists and forces a password change on first login. Existing database volumes must apply `014_operations_phase1.sql` before starting the updated backend.

Phase 2 adds versioned Workflow templates, normalized per-order snapshots, lifecycle transitions, dynamic forms, handoff discrepancies, triggers, and private files. Apply `015_operations_phase2.sql` after Phase 1. Set `WORKFLOW_FILE_DIR` when the default `/app/storage/workflow-files` is not suitable; production compose files persist this directory in a backend-only volume that is never mounted by nginx.

Phase 3 adds order lines, Batch-scoped fulfillment, immutable inventory movements, private Shipment/Package files, partial delivery, operational costs, Workflow scopes and controlled branching. Apply `016_operations_phase3.sql` after Phase 2; existing Order-scoped instances are backfilled without receiving new Batch or Step records.

## Production configuration and health

Production startup fails closed when JWT is shorter than 32 characters, secure cookies are disabled, allowed origins are empty, the database pool is invalid, or the fake SMS provider is selected. Relevant controls are `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, `HTTP_*_TIMEOUT`, `MAX_UPLOAD_SIZE_MB`, `APP_VERSION`, `GIT_COMMIT`, and `BUILD_TIME`. Keep `SMS_PROVIDER=disabled` until a real provider is explicitly implemented.

- `/health` is process liveness only.
- `/ready` verifies database connectivity and migration 19.
- `/api/v1/version` returns non-sensitive build and schema metadata.
- Private workflow, payment, shipment, quality, and installation files are served only by authorized API endpoints from `WORKFLOW_FILE_DIR`; nginx must never mount that volume.

Application modules can be disabled from `/panel/dashboard/settings`. Disabling a module blocks new mutations with `MODULE_DISABLED` but preserves read access to existing records. Public marketplace login remains independent from the customer operations portal.

## RBAC and operational recovery

Users may have multiple roles and effective permissions are the union of active roles. `SUPER_ADMIN` receives current and future permissions while backend invariants protect the last Super Admin. Disabling a user, resetting a password, or revoking sessions deletes refresh tokens and invalidates issued access tokens through `auth_invalid_before`.

Known safe reconciliation and correction operations are exposed at `/panel/dashboard/admin-tools`; arbitrary SQL and generic record editors are intentionally excluded. Every mutation requires a reason, permission, idempotency key, request ID, and audit entry.

## Backup and restore

Back up PostgreSQL and the private file volume daily. Retain 14 daily, 8 weekly, and 12 monthly recovery points. Encrypt backups, restrict their credentials, and test a complete restore monthly in an isolated environment. A valid restore includes matching database and private-file snapshots, followed by `/ready`, the smoke script, customer-isolation tests, and a document download check. See [the operations guide](docs/operations-guide.md) and [the release checklist](docs/production-release-checklist.md).

## Data import (SangeHassan export)
The extracted JSON lives in `data/` and images in `data/images/products`. We added a Go importer and expanded the DB schema (attributes + product-category relations).

Run after the database is up:
```sh
# reset DB if this is an existing volume (init SQL runs only once)
cd deploy
docker compose -f docker-compose-dev.yml down -v
docker compose -f docker-compose-dev.yml up -d db

# import from host (override DB_HOST to localhost)
cd ../back
DB_HOST=localhost DB_PORT=5432 DB_USER=sangehassan DB_PASSWORD=sangehassan_dev DB_NAME=sangehassan DB_SSLMODE=disable \\
  go run ./cmd/importdata -data ../data
```

Notes:
- The importer copies product images into `back/storage/images/products` and writes `/images/products/...` URLs.
- `permalink` from the export is ignored as requested.

## Product tags import (normalized terms)
This repo supports normalized product tags (many-to-many) via `product_terms` + `product_term_links`.

After `importdata` (or anytime later), import the structured CSV exports:
```sh
cd back
DB_HOST=localhost DB_PORT=5432 DB_USER=sangehassan DB_PASSWORD=sangehassan_dev DB_NAME=sangehassan DB_SSLMODE=disable \\
  go run ./cmd/importmeta --extended ../data/products_extended.csv --descriptions ../data/products_with_descriptions.csv
```

Import behavior:
- Default is additive/merge (does not wipe existing tags).
- Use `--overwrite-terms` only if you intentionally want to replace existing product-term links.
