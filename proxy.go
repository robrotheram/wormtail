package main

import (
	"context"
	"io"
	"log/slog"

	"tailscale.com/client/tailscale"
)

type Proxy struct {
	ts           *tailscale.LocalClient
	lconn, rconn io.ReadWriteCloser
	erred        bool
	errsig       chan bool
	log          *TimeSeries
}

func (p *Proxy) Start(host string, port uint16) {
	defer p.lconn.Close()
	var err error
	//connect to remote
	p.rconn, err = p.ts.DialTCP(context.Background(), host, port)
	if err != nil {
		slog.Warn("remote connection failed: %v", err)
		return
	}
	defer p.rconn.Close()
	go p.pipe(p.lconn, p.rconn)
	go p.pipe(p.rconn, p.lconn)
	<-p.errsig
}

func (p *Proxy) pipe(src, dst io.ReadWriter) {
	islocal := src == p.lconn
	//directional copy (64k buffer)
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			p.err("Read failed '%s'\n", err)
			return
		}
		b := buff[:n]
		n, err = dst.Write(b)
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
		if islocal {
			p.log.LogSent(uint64(n))
		} else {
			p.log.LogRecived(uint64(n))
		}
	}
}

func (p *Proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		slog.Warn(s, err)
	}
	p.errsig <- true
	p.erred = true
}
