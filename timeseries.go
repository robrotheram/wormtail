package main

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

type TimeSeries struct {
	mu      sync.Mutex
	Data    []DataPoint
	bucket  time.Duration
	maxSize int
}

func NewTimeSeries(bucket time.Duration, maxSize int) *TimeSeries {
	return &TimeSeries{
		Data:    make([]DataPoint, 0, maxSize),
		maxSize: maxSize,
		bucket:  bucket,
	}
}

func (ts *TimeSeries) Total() ProxyStats {
	ps := ProxyStats{Sent: 0, Received: 0}
	for _, stat := range ts.Data {
		ps.Received += stat.Value.Received
		ps.Sent += stat.Value.Sent
	}
	return ps
}

func (ts *TimeSeries) LogSent(value uint64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	now := time.Now().Truncate(ts.bucket)
	if len(ts.Data) > 0 && ts.Data[len(ts.Data)-1].Timestamp == now {
		ts.Data[len(ts.Data)-1].Value.Sent += value
	} else {
		ts.Data = append(ts.Data, DataPoint{Timestamp: now, Value: ProxyStats{Sent: value, Received: 0}})
	}
	if len(ts.Data) > ts.maxSize {
		ts.Data = ts.Data[1:]
	}
}

func (ts *TimeSeries) LogRecived(value uint64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	now := time.Now().Truncate(ts.bucket)
	if len(ts.Data) > 0 && ts.Data[len(ts.Data)-1].Timestamp == now {
		ts.Data[len(ts.Data)-1].Value.Received += value
	} else {
		ts.Data = append(ts.Data, DataPoint{Timestamp: now, Value: ProxyStats{Sent: 0, Received: value}})
	}
	if len(ts.Data) > ts.maxSize {
		ts.Data = ts.Data[1:]
	}
}
