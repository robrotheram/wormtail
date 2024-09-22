package utils

import (
	"sync"
	"time"
)

type DataPoint struct {
	Timestamp time.Time
	Value     ProxyStats
}

type ProxyStats struct {
	Sent     uint64
	Received uint64
}

type TimeSeriesData struct {
	Points  []DataPoint
	Total   ProxyStats
	bucket  time.Duration
	maxSize int
}

type TimeSeries struct {
	mu   sync.Mutex
	Data TimeSeriesData
}

func NewTimeSeries(bucket time.Duration, maxSize int) *TimeSeries {
	return &TimeSeries{
		Data: TimeSeriesData{
			Points:  make([]DataPoint, 0, maxSize),
			Total:   ProxyStats{},
			maxSize: maxSize,
			bucket:  bucket,
		},
	}
}

func (tsd *TimeSeriesData) Add(point ProxyStats) {
	now := time.Now().Truncate(tsd.bucket)
	if len(tsd.Points) > 0 && tsd.Points[len(tsd.Points)-1].Timestamp == now {
		current := tsd.Points[len(tsd.Points)-1].Value
		current.Sent += point.Sent
		current.Received += point.Received
		tsd.Points[len(tsd.Points)-1].Value = current
	} else {
		tsd.Points = append(tsd.Points, DataPoint{Timestamp: now, Value: point})
	}
	if len(tsd.Points) > tsd.maxSize {
		tsd.Points = tsd.Points[1:]
	}
	ps := ProxyStats{Sent: 0, Received: 0}
	for _, stat := range tsd.Points {
		ps.Received += stat.Value.Received
		ps.Sent += stat.Value.Sent
	}
	tsd.Total = ps
}

func (ts *TimeSeries) LogSent(value uint64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.Data.Add(ProxyStats{Sent: value})
}

func (ts *TimeSeries) LogRecived(value uint64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.Data.Add(ProxyStats{Received: value})
}
