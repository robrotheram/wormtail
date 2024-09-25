package router

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"time"
	"warptail/pkg/utils"

	"tailscale.com/client/tailscale"
)

type NetworkRoute struct {
	config   utils.RouteConfig
	status   RouterStatus
	client   *tailscale.LocalClient
	data     *utils.TimeSeries
	listener *net.TCPListener
	erred    bool
	errsig   chan bool
}

func NewNetworkRoute(config utils.RouteConfig, client *tailscale.LocalClient) *NetworkRoute {
	return &NetworkRoute{
		config: config,
		data:   utils.NewTimeSeries(time.Second, 1000),
		status: STOPPED,
		client: client,
		erred:  false,
		errsig: make(chan bool),
	}
}

func (route *NetworkRoute) Status() RouterStatus {
	return route.status
}

func (route *NetworkRoute) Config() utils.RouteConfig {
	return route.config
}

func (route *NetworkRoute) Stats() utils.TimeSeriesData {
	return route.data.Data
}

func (route *NetworkRoute) Update(config utils.RouteConfig) error {
	route.Stop()
	route.config = config
	return route.Start()
}

func (route *NetworkRoute) Stop() error {
	route.listener.Close()
	route.status = STOPPED
	return nil
}

func (route *NetworkRoute) Start() error {
	route.status = STARTING
	slog.Info(fmt.Sprintf("listening on tcp://localhost:%d", route.config.Port))
	laddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", route.config.Port))
	if err != nil {
		return err
	}
	route.listener, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}

	route.status = RUNNING
	for {
		conn, err := route.listener.AcceptTCP()
		if err != nil {
			return err
		}
		defer conn.Close()
		remoteAddr := conn.RemoteAddr().String()
		fmt.Println("Remote address:", remoteAddr)
		go route.handle(conn, conn)
	}
}

func (route *NetworkRoute) handle(src io.Reader, dst io.Writer) {
	var err error
	proxy, err := route.client.UserDial(context.Background(), string(route.config.Type), route.config.Machine.Address, route.config.Machine.Port)
	if err != nil {
		log.Printf("remote connection failed: %v", err)
		return
	}
	defer proxy.Close()
	go route.send(src, proxy)
	go route.receive(proxy, dst)
	<-route.errsig
}

func (route *NetworkRoute) send(src io.Reader, dst io.Writer) {
	buff := make([]byte, 0xffff)
	for {
		data := route.pipe(src, dst, buff)
		route.data.LogSent(uint64(data))
	}
}

func (route *NetworkRoute) receive(src io.Reader, dst io.Writer) {
	buff := make([]byte, 0xffff)
	for {
		data := route.pipe(src, dst, buff)
		route.data.LogRecived(uint64(data))
	}
}

func (route *NetworkRoute) pipe(src io.Reader, dst io.Writer, buff []byte) uint64 {
	n, err := src.Read(buff)
	if err != nil {
		route.err("Read failed '%s'\n", err)
		return 0
	}
	b := buff[:n]
	n, err = dst.Write(b)
	if err != nil {
		route.err("Write failed '%s'\n", err)
		return 0
	}
	return uint64(n)
}

func (route *NetworkRoute) err(s string, err error) {
	log.Println(s, err)
	if route.erred {
		return
	}
	route.errsig <- true
	route.erred = true
}
