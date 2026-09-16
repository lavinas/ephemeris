---
name: ephemeris-ops
description: "Use when working on the Ephemeris monorepo: billing/planner deployments, Docker Compose, NGINX, database migrations, production reverse-proxy rules, or service health checks."
---

# Ephemeris Operations

## Scope

This workspace contains a small multi-service platform with separate services for billing, planner, signup, and a shared vendor front-end/proxy layer. Use this skill when the task touches deployment, service routing, dockerized apps, database migrations, or production-facing configuration.

## Repository map

- `billing/` — billing backend and invoice logic
- `planner/` — planner backend and session management
- `signup/` — signup service
- `vendor/` — static frontend and reverse-proxy config
- `notebooks/` — operational notebooks and scripts
- `accouting/` — accounting analysis and CSV files

## Default operating assumptions

- Prefer small, targeted edits over broad refactors.
- Keep production configuration separate from development config.
- If a change impacts service routing, check both the app and the proxy configuration.
- When a change touches schemas or migrations, verify the matching SQL migration files and service startup flow.
- Favor explicit ports and host routing over implicit defaults when debugging production traffic.

## Common workflows

### 1. Inspect a service

Start with the relevant module and confirm its runtime contract:

- `billing/go.mod` and `billing/cmd/...` for API entrypoints
- `planner/go.mod` and `planner/cmd/...` for planner services
- `vendor/nginx*.conf` for reverse-proxy and IP filtering rules
- `billing/migrations/*.sql` and `planner/migration/*.sql` for schema changes

### 2. Validate Docker-based setup

For local dev and production-related changes, inspect the relevant compose file before changing environment variables or ports:

- `billing/docker-compose.dev.yml`
- `billing/docker-compose.prod.yml`
- `planner/docker-compose.dev.yml`
- `planner/docker-compose.prod.yml`
- `vendor/docker-compose.dev.yml`
- `vendor/docker-compose.prod.yml`

Typical checks:

- confirm service names and container links
- confirm exposed ports match app expectations
- confirm proxy targets use the right host and port
- confirm the app listens on the correct interface in the container

### 3. Validate NGINX routing

When working in `vendor/nginx*.conf`:

- verify `server_name` and `listen` match the deployment target
- confirm `location` blocks line up with the expected path prefixes
- check that `proxy_pass` targets the right internal service/port
- verify `proxy_set_header` values for host, real IP, and forwarded headers
- ensure restrictive IP allowlists are intentional, especially in production

If a route does not resolve as expected, confirm the upstream application is receiving requests on the correct port and path prefix.

### 4. Check database and migration impact

When a bug or feature touches persistence:

- read the migration file for the affected service
- verify whether the app expects the schema to be bootstrapped on startup or via external migration steps
- check if there are SQL scripts used for backup, inspection, or performance debugging
- validate whether the route or API contract depends on a new column or table

### 5. Verify production-safe behavior

Before finalizing a production change:

- ensure the change is reversible or narrowly scoped
- avoid broad allow-list or privileged access changes without review
- confirm the service still works behind the proxy layer
- validate the port and internal host route before committing a config change

## Useful commands

Run these from the repo root when needed:

```bash
cd billing && go test ./...
cd planner && go test ./...

docker compose -f billing/docker-compose.dev.yml up --build
docker compose -f planner/docker-compose.dev.yml up --build

nginx -t -c /path/to/nginx.conf
```

## Production notes

This project includes a vendor NGINX configuration with restricted IP access and forwarding rules into internal services. Treat this configuration as a critical security boundary and verify it carefully before editing any `allow`/`deny` rules or `proxy_pass` targets.

## Task guidance

When a request involves the project’s infrastructure, prefer this order:

1. Identify the affected service and config file.
2. Inspect the local app and proxy configuration together.
3. Check migrations or environment assumptions.
4. Make the smallest targeted fix.
5. Validate with the relevant runtime or config check.

This skill is meant to keep repo work consistent with the project’s real deployment structure and not to over-generalize toward unrelated stacks.
