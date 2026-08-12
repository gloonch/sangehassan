# Production release, rollback and smoke checklist

## Release gate

- [ ] A current database and private-file backup exists and can be read.
- [ ] Checksums of migrations 014–018 match the deployed copies.
- [ ] The 014→019 chain passes on a new PostgreSQL 16 database and 019 passes on a restored application database.
- [ ] `go test ./...`, critical race tests, `go vet ./...`, panel tests/build, and website tests/build/prerender pass.
- [ ] Customer isolation and the nine-role RBAC matrix pass.
- [ ] Minimal order, internal order, export with deposit, and installation/acceptance/close paths pass.
- [ ] All four Compose files validate and production secrets are supplied through environment variables.
- [ ] `COOKIE_SECURE=true`, JWT is at least 32 characters, origins are exact HTTPS origins, and `SMS_PROVIDER=disabled`.
- [ ] Private file storage is writable by API/worker and is not mounted by nginx.
- [ ] Chrome automated checks and the manual Safari/Mobile Safari 360px checklist pass.

Do not label a release production-ready when migration, customer isolation, RBAC, or any critical path is unverified.

Run `./scripts/verify-operations-migrations.sh` before applying migration 019.
With Docker running, execute `./scripts/migration-smoke-test.sh` for the disposable PostgreSQL 16 chain and idempotency check.

## Deployment smoke test

- [ ] `/health`, `/ready`, and `/api/v1/version` return success without secrets.
- [ ] Internal login works only through `/panel/`; forced password change and logout work.
- [ ] Dashboard order is My Actions, Alerts, Quick Actions, then metrics.
- [ ] Search never returns an entity outside the current role.
- [ ] Order/workflow form saves a non-file draft, warns on unsaved exit, submits once, and clears the draft.
- [ ] Pagination, Access Denied, Not Found, Settings and Admin Tools render correctly.
- [ ] Customer A cannot access Customer B’s order, workflow, timeline, payment, document, shipment, installation, notification, or file.
- [ ] Disabled user and revoked-session access tokens are rejected.
- [ ] SMS-disabled and notification-template failures do not rollback a core order mutation.

## Rollback

1. Stop new mutations at the reverse proxy or place the application in maintenance mode.
2. Record release version, commit, build time, migration version, Request IDs, and failure time.
3. Roll back application containers to the previously verified images. Do not reverse schema with ad-hoc SQL.
4. If data restoration is required, restore the matching PostgreSQL and private-file snapshots into an isolated target first.
5. Run readiness, smoke, customer-isolation and document-download checks against the restored target.
6. Switch traffic only after sign-off; retain failed-state logs and backups for root-cause analysis.

Migration 019 is additive. Its tables/indexes can remain while application containers are rolled back, but that choice must be validated against the exact previous application version.
