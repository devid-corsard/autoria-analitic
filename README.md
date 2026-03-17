# autoria

Personal tool to fetch newest cars from the [Autoria](https://auto.ria.com) API and store them in PostgreSQL. Use Superset in Docker to visualize the data.

## Setup

1. **Run Postgres** and create a database:

   ```bash
   createdb autoria_db   # or create via psql
   ```

2. **Create `.env`** with Postgres credentials and Autoria API key:

   ```env
   api_key=your_autoria_api_key

   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=your_postgres_user
   DB_PASSWORD=your_postgres_password
   DB_NAME=autoria_db
   ```

   Copy from `.env.example` and fill in your values.

3. **Run migrations**:

   ```bash
   make migrate-up
   # or: go run ./cmd/migrate
   ```

## Usage

**Fetch new car IDs and fill details:**

```bash
go run . -fetch-new
```

Fetches up to 1000 newest car IDs from the API, saves them to the DB, then fetches and stores details for any records missing them.

**Fill details only (when rate limited):**

If you hit the API rate limit (429), run without `-fetch-new` to only fill details for records that already have IDs but no details:

```bash
go run .
```

This avoids the search API and only uses the GetByID endpoint, which helps when you're rate limited.

## Superset (visualization)

1. Ensure Postgres is running and the database exists.
2. Create the Docker network: `docker network create frogfort-network`
3. Start Superset: `make superset` or `docker compose up -d`
4. Open http://localhost:8088 (login: `admin` / `admin`)
5. Add Postgres in **Data → Databases**: host=`postgres`, port=5432, and your DB_USER, DB_PASSWORD, DB_NAME from `.env`

> When running Superset in Docker, use host `postgres` (the container name) so it can reach the Postgres container on the shared network.
