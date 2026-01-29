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
```sh
cd deploy
docker compose -f docker-compose-prod.yml up --build
```

Services:
- Website: http://localhost:3000
- Admin panel: http://localhost:3001
- API: http://localhost:8080
- Images: http://localhost:8081/images/

## Env files
Update secrets in:
- `deploy/.dev.env`
- `deploy/.prod.env`

Image hosting base URL:
- `VITE_IMAGE_BASE_URL` (defaults to `http://localhost:8081` if unset)

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
