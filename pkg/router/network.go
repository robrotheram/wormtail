package router

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
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
	quit     chan bool
	exited   chan bool
}

func NewNetworkRoute(config utils.RouteConfig, client *tailscale.LocalClient) *NetworkRoute {
	return &NetworkRoute{
		config: config,
		data:   utils.NewTimeSeries(time.Second, 1000),
		status: STOPPED,
		client: client,
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
	route.status = STOPPING
	close(route.quit)
	<-route.exited
	fmt.Println("Stopped successfully")
	route.status = STOPPED
	return nil
}

func (route *NetworkRoute) Start() error {
	if route.status == RUNNING {
		route.Stop()
	}
	route.status = STARTING
	laddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", route.config.Port))
	if err != nil {
		return err
	}
	route.quit = make(chan bool)
	route.exited = make(chan bool)

	route.listener, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}

	go route.serve()
	route.status = RUNNING
	return nil
}

func (route *NetworkRoute) serve() {
	var handlers sync.WaitGroup
	for {
		select {
		case <-route.quit:
			fmt.Println("Shutting down...")
			route.listener.Close()
			handlers.Wait()
			close(route.exited)
			return
		default:
			//fmt.Println("Listening for clients")
			route.listener.SetDeadline(time.Now().Add(1e9))
			conn, err := route.listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				fmt.Println("Failed to accept connection:", err.Error())
			}
			handlers.Add(1)
			go func() {
				route.handleConnection(conn)
				handlers.Done()
			}()
		}
	}
}

func (route *NetworkRoute) handleConnection(conn net.Conn) {
	proxy, err := route.client.UserDial(context.Background(), string(route.config.Type), route.config.Machine.Address, route.config.Machine.Port)
	if err != nil {
		log.Printf("remote connection failed: %v", err)
		return
	}

	sendWriter := &ConnMonitor{rw: proxy}
	reciveWriter := &ConnMonitor{rw: conn}

	wg := &sync.WaitGroup{}
	wg.Add(3)
	go route.copy(reciveWriter, sendWriter, wg)
	go route.copy(sendWriter, reciveWriter, wg)
	go route.monitor(sendWriter, reciveWriter, wg)
	wg.Wait()
}
func (route *NetworkRoute) monitor(to, from *ConnMonitor, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		select {
		case <-route.quit:
			to.Close()
			from.Close()
			return
		default:
			route.data.LogRecived(uint64(to.bytesRead))
			route.data.LogSent(uint64(to.bytesWritten))
		}
	}
}

func (route *NetworkRoute) copy(from, to io.ReadWriter, wg *sync.WaitGroup) {
	defer wg.Done()
	select {
	case <-route.quit:
		return
	default:
		if _, err := io.Copy(to, from); err != nil {
			return
		}
	}
}

// func (route *NetworkRoute) handle(src io.Reader, dst io.Writer) {
// 	var err error
// 	proxy, err := route.client.UserDial(context.Background(), string(route.config.Type), route.config.Machine.Address, route.config.Machine.Port)
// 	if err != nil {
// 		log.Printf("remote connection failed: %v", err)
// 		return
// 	}
// 	defer proxy.Close()
// 	go route.send(src, proxy)
// 	go route.receive(proxy, dst)
// 	<-route.errsig
// }

// func (route *NetworkRoute) send(src io.Reader, dst io.Writer) {
// 	buff := make([]byte, 0xffff)
// 	for {
// 		data := route.pipe(src, dst, buff)
// 		route.data.LogSent(uint64(data))
// 	}
// }

// func (route *NetworkRoute) receive(src io.Reader, dst io.Writer) {
// 	buff := make([]byte, 0xffff)
// 	for {
// 		data := route.pipe(src, dst, buff)
// 		route.data.LogRecived(uint64(data))
// 	}
// }

// func (route *NetworkRoute) pipe(src io.Reader, dst io.Writer, buff []byte) uint64 {
// 	n, err := src.Read(buff)
// 	if err != nil {
// 		route.err("Read failed '%s'\n", err)
// 		return 0
// 	}
// 	b := buff[:n]
// 	n, err = dst.Write(b)
// 	if err != nil {
// 		route.err("Write failed '%s'\n", err)
// 		return 0
// 	}
// 	return uint64(n)
// }

// func (route *NetworkRoute) err(s string, err error) {
// 	log.Println(s, err)
// 	if route.erred {
// 		return
// 	}
// 	route.errsig <- true
// 	route.erred = true
// }
