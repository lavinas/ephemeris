# Planner (5433)
PGPASSWORD=root psql -h localhost -p 5433 -U root -d planner -c "DROP SCHEMA IF EXISTS planner CASCADE;" && \
PGPASSWORD=root pg_dump -h 192.168.1.138 -p 5433 -U root -d planner | PGPASSWORD=root psql -h localhost -p 5433 -U root -d planner && \
# Billing (5432)
PGPASSWORD=root psql -h localhost -p 5432 -U root -d billing -c "DROP SCHEMA IF EXISTS billing CASCADE;" && \
PGPASSWORD=root pg_dump -h 192.168.1.138 -p 5432 -U root -d billing | PGPASSWORD=root psql -h localhost -p 5432 -U root -d billing && \
# Signup (5434)
PGPASSWORD=root psql -h localhost -p 5434 -U root -d signup -c "DROP SCHEMA IF EXISTS signup CASCADE;" && \
PGPASSWORD=root pg_dump -h 192.168.1.138 -p 5434 -U root -d signup | PGPASSWORD=root psql -h localhost -p 5434 -U root -d signup
