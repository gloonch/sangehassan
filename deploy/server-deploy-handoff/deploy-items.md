# Server Deploy Items (Collected)

This folder summarizes what is needed for Docker-based server deploy.

## Required for production deploy
- `deploy/docker-compose-prod-com.yml` for `sangehassan.com` / `91.239.211.60`
- `deploy/docker-compose-prod-ir.yml` for `sangehassan.ir` / `185.53.141.157`
- `deploy/.prod-com.env` copied from `deploy/.prod-com.env.example`
- `deploy/.prod-ir.env` copied from `deploy/.prod-ir.env.example`
- `deploy/nginx/prod-com.conf`
- `deploy/nginx/prod-ir.conf`
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
- `pg_data_prod` -> Postgres data (`/var/lib/postgresql/data`) on each VPS
- `images_data_prod` -> shared images between backend (`/app/storage/images`) and nginx (`/var/www/images`) on each VPS

### Bind mounts
- `./postgres/init:/docker-entrypoint-initdb.d`
- `./nginx/prod-com.conf:/etc/nginx/nginx.conf:ro` on prod-com
- `./nginx/prod-ir.conf:/etc/nginx/nginx.conf:ro` on prod-ir
- `../back/storage/images:/seed/images:ro`

## Important notes
- SQL files in `deploy/postgres/init/` run only when Postgres volume is created first time.
- Update secrets in `.prod-com.env` or `.prod-ir.env` before running on server.
- Use `docker compose --env-file <target-env> -f <target-compose> up -d --build` so Vite build args are also loaded from the target env file.
