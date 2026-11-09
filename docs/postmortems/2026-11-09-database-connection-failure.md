# Incident Postmortem - Database Connection Failure - 2026-11-09

- **Incident Date/Time:** 2026-11-09, during initial deployment. The exact timestamp of the first connection-refused error was not captured.
- **Severity:** SEV4
- **Summary:** During the initial deployment, the backend repeatedly failed to establish a connection to PostgreSQL. Investigation revealed a database startup-readiness issue and a misconfigured PostgreSQL healthcheck that attempted connections using the `root` role instead of `postgres`.

## Impact

- The backend repeatedly failed during initialization and entered a restart loop.
- Background workers could not reliably start, preventing the application from polling its configured sources and delivering updates.
- The issue affected the local development/deployment environment. No external user impact was reported.
- During troubleshooting, `docker compose down --rmi local --volumes --remove-orphans` removed the PostgreSQL named volume. Consequently, any existing data in that volume was erased.

## Timeline

| Time (UTC) | Event/Action |
|------------|--------------|
| Not recorded | Initial Docker Compose startup. The backend attempted to connect to PostgreSQL but received `connection refused`. |
| 07:26:14.177 | River reported that `river_queue` and `river_leader` relations did not exist. This was a separate database migration issue discovered during initialization. |
| 07:47:53 - 07:54:21 | PostgreSQL repeatedly logged `FATAL: role "root" does not exist`. The healthcheck had not explicitly specified a PostgreSQL username. |
| 07:57:27 | PostgreSQL reported that the database system was ready to accept connections. |
| 08:02:13 | PostgreSQL completed a fresh database initialization after the Docker volume was recreated. |
| Later | Updated the healthcheck to explicitly use the `postgres` role and verified PostgreSQL connectivity from both the database container and another container on the Compose network. |
| Later | River successfully started. Database migration issues were investigated and the application migrations were applied. |

## Root Cause

The initial connection failure was caused by a startup-readiness issue.

The backend attempted to establish a TCP connection with PostgreSQL before the database server was ready to accept connections. Starting containers through Docker Compose does not, by itself, guarantee that the database is ready to accept connections.

A separate configuration issue was also discovered in the PostgreSQL healthcheck.

The healthcheck used `pg_isready` without explicitly specifying a PostgreSQL username. Consequently, it defaulted to the operating-system user running the command inside the container, which was `root`. PostgreSQL did not have a corresponding `root` role, resulting in repeated authentication-related log messages:

`FATAL: role "root" does not exist`

The application's database URL was already configured to use the `postgres` role. Therefore, the `root` errors originated from the healthcheck rather than the application's database connection configuration.

## Resolution

- Updated the PostgreSQL healthcheck to explicitly specify the database user and database:
  `pg_isready -U postgres -d $POSTGRES_DB`
- Configured the backend to depend on PostgreSQL becoming healthy before starting.
- Verified local PostgreSQL readiness using:
  `docker compose exec db pg_isready -U postgres -d kimishinu`
- Verified Docker network connectivity using another container on the Compose network:
  `pg_isready -h db -p 5432 -U postgres -d kimishinu`
- Investigated and applied the required database migrations. River and application tables were subsequently available, allowing the application to start its workers and process updates.

## Lessons Learned

- Starting a container does not mean its service is ready to accept connections.
- Docker Compose healthchecks should explicitly specify the intended database user and database.
- The PostgreSQL healthcheck and the application's database connection are separate operations; errors from one should not automatically be attributed to the other.
- `connection refused`, authentication errors, and missing database relations represent different failure categories and require different investigations. One does not necessarily imply the other.
- Database connectivity should be verified before initializing components that depend on it, such as River.