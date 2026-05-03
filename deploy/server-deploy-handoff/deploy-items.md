# Server Deploy Items (Collected)

This folder summarizes what is needed for Docker-based server deploy.

## Required for production deploy
- `deploy/docker-compose-prod.yml`
- `deploy/.prod.env`
- `deploy/nginx/nginx.conf`
- `deploy/postgres/init/001_init.sql`
- `deploy/postgres/init/002_auth.sql`
- `deploy/postgres/init/003_stone_ads.sql`
- `deploy/postgres/init/004_projects.sql`
- `deploy/postgres/init/005_products_metadata.sql`

## Build contexts required by compose
`docker-compose-prod.yml` builds images from local source, so these directories must exist on server too:
- `back/`
- `front/`

## Optional but useful
- `deploy/sangehassan.dump` (database dump)
- `deploy/scripts/migrate_external_content_section_images.sh`
- `deploy/scripts/seed_projects_demo.sql`

## Production volumes (from compose)
### Named volumes
- `pg_data_prod` -> Postgres data (`/var/lib/postgresql/data`)
- `images_data_prod` -> shared images between backend (`/app/storage/images`) and nginx (`/var/www/images`)

### Bind mounts
- `./postgres/init:/docker-entrypoint-initdb.d`
- `./nginx/nginx.conf:/etc/nginx/nginx.conf:ro`
- `../back/storage/images:/seed/images:ro`

## Important notes
- SQL files in `deploy/postgres/init/` run only when Postgres volume is created first time.
- Update secrets in `.prod.env` before running on server.
