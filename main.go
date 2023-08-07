package main

import (
	"database/sql"
	"flag"
	"log"
	"math"
	"sort"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/zknill/timescalett/pkg/querystats"
	"github.com/zknill/timescalett/pkg/timeseries"
)

var (
	nWorkers = flag.Int("n", 10, "the number of workers to start")
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatal("bad args", flag.Args())
	}

	// database pool
	db, err := sql.Open("pgx", "host=timescaledb user=postgres database=homework password=password")
	if err != nil {
		log.Fatal("open db", err)
	}

	tsDB := timeseries.NewDB(db)

	pool := querystats.NewWorkerPool(tsDB, *nWorkers)

	records, err := querystats.ParseCSVFile(flag.Arg(0))
	if err != nil {
		log.Fatal("read records", err)
	}

	go pool.Submit(records)

	var results []time.Duration

	for d := range pool.Results() {
		results = append(results, d)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Nanoseconds() < results[j].Nanoseconds()
	})

	min := results[0]
	max := results[len(results)-1]
	med := median(results)
	sum := sum(results)
	mean := mean(sum, len(results))
	p50 := percentile(results, 0.5)
	p90 := percentile(results, 0.9)
	p95 := percentile(results, 0.95)

	log.Printf("%10s: %s", "total", sum)
	log.Printf("%10s: %s", "max", max)
	log.Printf("%10s: %s", "min", min)
	log.Printf("%10s: %s", "median", med)
	log.Printf("%10s: %s", "mean", mean)
	log.Printf("%10s: %s/%s/%s", "P50/90/95", p50, p90, p95)
}

func median(durations []time.Duration) time.Duration {
	length := len(durations)
	if length%2 == 1 {
		return durations[length/2]
	}

	middle1 := durations[length/2]
	middle2 := durations[length/2-1]
	return (middle1 + middle2) / 2
}

func sum(durations []time.Duration) time.Duration {
	var sum time.Duration

	for _, duration := range durations {
		sum += duration
	}

	return sum
}

func mean(sum time.Duration, count int) time.Duration {
	if count == 0 {
		return 0
	}

	return sum / time.Duration(count)
}

func percentile(data []time.Duration, percentile float64) time.Duration {
	if len(data) == 0 {
		return 0
	}

	index := int(math.Ceil(float64(len(data))*percentile/100.0)) - 1

	return data[index]
}
