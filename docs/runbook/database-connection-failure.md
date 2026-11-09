# Database Connection Failure

## How to Identify It
In most cases, a database connection failure does not have a predefined location where it occurs. It can happen in either the main application or a background worker. However, there is a clear indicator: in most cases, the error message will state `connection refused`. This serves as the smoking gun that the database is currently unreachable.

## How to Find the Cause
### Based on the Location of the Error
As noted, the exact place where the error occurs varies, so you will need to check the logs of both your main application and background workers. Here are examples of logs that clearly point to the issue:

- **Occurred in the main application:**
```go
panic: error making initial connection to database:
failed to connect to `user=postgres database=kimishinu`:
172.18.0.3:5432 (db):
dial error: dial tcp 172.18.0.3:5432:
connect: connection refused

goroutine 1 [running]:
main.main()
    /app/cmd/sheena/main.go:160 +0xd1c
```
- **Explanation:** We can easily tell that the failure occurred during the main application's execution due to the reference to `/app/cmd/sheena/main.go:160` and the `panic:` keyword. The primary causes for this error are:
  - The database is not running or reachable (`172.18.0.3:5432` means the database is expected at that IP address and port within a Docker network, but the container may be down).
  - A startup race condition where the database is not yet ready to accept connections. Usual solutions involve waiting a few seconds before retrying or restarting the application containers (leaving the database container running).
  - Incorrect connection details (username, password, database name, host, or port). This can be verified by checking the application's configuration file or environment variables.

- **Occurred outside the main application (background worker):**
```text
time=2026-11-09T07:26:14.177Z
level=ERROR
msg="producer: Error fetching initial queue settings"
err="ERROR: relation \"river_queue\" does not exist (SQLSTATE 42P01)"
```
- **Explanation:** This error is slightly more subtle; while not a direct `connection refused` socket error, it indicates a related database availability or state issue. The main causes include:
  - The database is not running or reachable (similar to the case above).
  - The database is reachable, but the application cannot connect properly due to incorrect credentials.
  - The database is reachable and credentials are correct, but the database schema is outdated. As the error message specifies, the relation `river_queue` does not exist, meaning the necessary migration has not been applied yet.

### Via Commands
- **Check if the database is running and accepting connections:**
```bash
docker compose exec <service> pg_isready -U postgres -d <database>
```
  - **HEALTHY:** `/var/run/postgresql:5432 - accepting connections`
  - **UNHEALTHY:** `/var/run/postgresql:5432 - no response` (or any status other than healthy)

- **Check database network connectivity:**
```bash
docker compose exec <service> nc -zv <service> 5432
```
  - **HEALTHY:** `db (172.18.0.2:5432) open`
  - **UNHEALTHY:** `nc: connect to db port 5432 (tcp) failed: Connection refused` (or any status other than healthy)