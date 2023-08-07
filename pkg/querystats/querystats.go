package querystats

import (
	"context"
	"encoding/csv"
	"hash/fnv"
	"os"
	"sync"
	"time"

	"github.com/zknill/timescalett/pkg/timeseries"
)

type Record struct {
	Hostname  string
	StartTime time.Time
	EndTime   time.Time
}

func ParseCSVFile(filename string) ([]Record, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var parsedRecords []Record
	for i, row := range records {
		if i == 0 {
			// skip headers
			continue
		}

		startTime, err := time.Parse("2006-01-02 15:04:05", row[1])
		if err != nil {
			return nil, err
		}

		endTime, err := time.Parse("2006-01-02 15:04:05", row[2])
		if err != nil {
			return nil, err
		}

		record := Record{
			Hostname:  row[0],
			StartTime: startTime,
			EndTime:   endTime,
		}
		parsedRecords = append(parsedRecords, record)
	}

	return parsedRecords, nil
}

type Worker struct {
	records chan Record
	results chan time.Duration

	db timeseries.DB
}

func (w Worker) work(wg *sync.WaitGroup) {
	defer wg.Done()

	for record := range w.records {
		n := time.Now()
		w.db.CPUStats(context.Background(), record.Hostname, record.StartTime, record.EndTime)
		elapsed := time.Now().Sub(n)
		w.results <- elapsed
	}
}

type WorkerPool struct {
	workers []Worker
	results chan time.Duration
	wg      *sync.WaitGroup
}

func NewWorkerPool(db timeseries.DB, n int) WorkerPool {
	pool := WorkerPool{
		results: make(chan time.Duration),
		wg:      &sync.WaitGroup{},
	}

	pool.wg.Add(n)

	for i := 0; i < n; i++ {
		w := Worker{
			records: make(chan Record),
			results: pool.results,
			db:      db,
		}

		go w.work(pool.wg)

		pool.workers = append(pool.workers, w)
	}

	return pool
}

func (wp *WorkerPool) Submit(records []Record) {
	for i := range records {
		record := records[i]

		h := fnv.New32a()
		h.Write([]byte(record.Hostname))
		hash := h.Sum32()

		i := int(hash) % len(wp.workers)
		wp.workers[i].records <- record
	}

	for _, w := range wp.workers {
		close(w.records)
	}

	wp.wg.Wait()

	close(wp.results)
}

func (wp *WorkerPool) Results() <-chan time.Duration {
	return wp.results
}
