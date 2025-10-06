# Dockerized Go + Nginx (Simple)

Minimal setup: environment variables are set inside `docker-compose.yml`.

## Provided DB values (already placed in docker-compose.yml)

- DB_HOST: migrate.cluster-c3o6qc26ias1.eu-west-1.rds.amazonaws.com
- DB_PORT: 3306
- DB_NAME: ocpp_CSGO
- DB_USER: lina
- DB_PASS: 123456

> Note: DB password field is left empty in the compose file — set it before running if the DB requires a password.

## Run

1. Edit `docker-compose.yml` → set `DB_PASS` if required.
2. Build & run:

```bash
docker compose build
docker compose up

```
