# Sange Hassan - Clean Architecture Rebuild

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

Admin credentials (seeded in DB):
- Username: `admin`
- Password: `Admin123!`

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
```

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
