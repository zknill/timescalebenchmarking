package timeseries

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const query = `
SELECT time_bucket('1 minute', ts) AS minute_bucket,
       host,
       MIN(usage) AS min_cpu_usage,
       MAX(usage) AS max_cpu_usage
FROM cpu_usage
WHERE host = ?
  AND ts >= ? AND ts <= ?
GROUP BY minute_bucket, host
ORDER BY minute_bucket;
`

type CPUStats struct {
	Host string
	Min  float64
	Max  float64
}

type DB struct {
	driver *sql.DB
}

func NewDB(driver *sql.DB) DB {
	return DB{driver: driver}
}

func (db *DB) CPUStats(ctx context.Context, host string, start, end time.Time) ([]CPUStats, error) {
	rows, err := db.driver.QueryContext(ctx, query, host, start, end)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	stats := []CPUStats{}

	for rows.Next() {
		s := CPUStats{}

		if err := rows.Scan(&(s.Host), &(s.Min), &(s.Max)); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		stats = append(stats, s)

	}

	return stats, nil
}
