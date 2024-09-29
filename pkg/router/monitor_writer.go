package router

import (
	"net"
	"sync"
)

type ConnMonitor struct {
	rw           net.Conn
	bytesRead    int64
	bytesWritten int64
	mu           sync.Mutex
}

func (crw *ConnMonitor) Close() error {
	return crw.rw.Close()
}

// Read method that counts the number of bytes read.
func (crw *ConnMonitor) Read(p []byte) (int, error) {
	n, err := crw.rw.Read(p)
	crw.mu.Lock()
	crw.bytesRead += int64(n)
	crw.mu.Unlock()
	return n, err
}

// Write method that counts the number of bytes written.
func (crw *ConnMonitor) Write(p []byte) (int, error) {
	n, err := crw.rw.Write(p)
	crw.mu.Lock()
	crw.bytesWritten += int64(n)
	crw.mu.Unlock()
	return n, err
}

// BytesRead returns the total number of bytes read so far.
func (crw *ConnMonitor) BytesRead() int64 {
	crw.mu.Lock()
	defer crw.mu.Unlock()
	return crw.bytesRead
}

// BytesWritten returns the total number of bytes written so far.
func (crw *ConnMonitor) BytesWritten() int64 {
	crw.mu.Lock()
	defer crw.mu.Unlock()
	return crw.bytesWritten
}
