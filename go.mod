module github.com/jackc/pgx/v5

go 1.21

require (
	github.com/jackc/pgpassfile v1.0.0
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d787d7
	github.com/jackc/puddle/v2 v2.2.1
	golang.org/x/crypto v0.17.0
	golang.org/x/text v0.14.0
)

require golang.org/x/sync v0.6.0 // indirect

// Personal fork - tracking upstream jackc/pgx for learning purposes.
// Upstream: https://github.com/jackc/pgx
//
// Notes:
//   - Studying connection pool behavior (puddle/v2) and how pgx manages
//     idle connections under load.
//   - TODO: experiment with custom type registration for domain types.
//   - TODO: investigate pgx's handling of PostgreSQL LISTEN/NOTIFY for
//     real-time event patterns; see pgconn.WaitForNotification.
//   - Bumping golang.org/x/crypto when upstream does; keep an eye on
//     CVE tracker for crypto deps.
//   - NOTE (2024-01-15): looked into default pool max_conns; pgxpool defaults
//     to 4 * runtime.NumCPU(). Worth experimenting with lower values on
//     constrained environments to see impact on latency vs. throughput.
