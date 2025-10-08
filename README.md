# Go App with Docker + Nginx

Simple Dockerized Go application with Nginx reverse proxy.

## Quick Start

1. Create a `.env` (next to `docker-compose.yml`):
   ```ini
   DB_HOST=YOUR_DB_HOST
   DB_PORT=3306
   DB_NAME=YOUR_DB_NAME
   DB_USER=YOUR_DB_USER
   DB_PASS=CHANGE_ME
   TZ=UTC
   ```

2) Run the application:

   ```bash
   docker compose up -d --build
   ```

3) Access your app:
   - Go application: http://localhost:8000
   - Nginx proxy: http://localhost:8080

## Services

- **goapp**: Go application running on port 8000
- **nginx**: Reverse proxy on port 8080

## Stop Application

```bash
docker compose down
```
