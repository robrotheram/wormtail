package main

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

type RouterStatus string

const (
	STARTING = RouterStatus("Starting")
	RUNNING  = RouterStatus("Running")
	STOPPED  = RouterStatus("Stopped")
)

type Route struct {
	Name   string `yaml:"name"`
	Listen uint16 `yaml:"listen_port"`
	Host   string `yaml:"host"`
	Port   uint16 `yaml:"port"`
	Status RouterStatus
	Stats  ProxyStats
	data   *TimeSeries
	signal chan any
}

type Router struct {
	routes map[string]*Route
	ts     tsnet.Server
	client *tailscale.LocalClient
	wg     sync.WaitGroup
}

func NewRouter(config Config) (*Router, error) {
	router := &Router{
		routes: make(map[string]*Route),
		wg:     sync.WaitGroup{},
	}
	s := new(tsnet.Server)
	s.Hostname = config.Tailscale.Hostnmae

	c, err := s.LocalClient()
	if err != nil {
		s.Close()
		return nil, err
	}
	router.client = c
	for _, route := range config.Routes {
		router.AddRoute(route)
	}
	return router, nil
}

func (r *Router) Close() {
	r.ts.Close()
}

func (r *Router) AddRoute(route Route) {
	route.data = NewTimeSeries(time.Second, 1000)
	route.signal = make(chan any)
	route.Status = STOPPED
	r.routes[route.Name] = &route

	r.StartRoute(route.Name)
}

func (r *Router) UpdateRoute(route Route) {
	r.routes[route.Name] = &route
	r.StopRoute(route.Name)
	r.StartRoute(route.Name)
}

func (r *Router) DeleteRoute(name string) {
	r.StopRoute(name)
	delete(r.routes, name)
}

func (r *Router) Get(name string) (*Route, error) {
	route, ok := r.routes[name]
	if !ok {
		return nil, fmt.Errorf("route %s not found", name)
	}
	route.Stats = route.data.Total()
	return route, nil
}

func (r *Router) GetAll() []*Route {
	routes := []*Route{}
	for key, _ := range r.routes {
		r, _ := r.Get(key)
		routes = append(routes, r)
	}
	return routes
}

func (r *Router) StartRoute(name string) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.routes[name].Start(r.client)
	}()
}

func (r *Router) StopRoute(name string) {
	r.routes[name].Stop()
}

func (r *Router) Start() {
	for _, route := range r.routes {
		r.StartRoute(route.Name)
	}
}

func (r *Router) Wait() {
	r.wg.Wait()
}

func (r *Route) Stop() {
	r.signal <- true
}

func (r *Route) Start(clinet *tailscale.LocalClient) error {
	r.Status = STARTING
	slog.Info(fmt.Sprintf("listening on tcp://localhost:%d", r.Listen))
	laddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", r.Listen))
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}

	r.Status = RUNNING
	for {
		select {
		case <-r.signal:
			r.Status = STOPPED
			listener.Close()
			return nil
		default:
			conn, _ := listener.AcceptTCP()
			remoteAddr := conn.RemoteAddr().String()
			fmt.Println("Remote address:", remoteAddr)

			p := &Proxy{
				ts:     clinet,
				lconn:  conn,
				erred:  false,
				errsig: make(chan bool),
				log:    r.data,
			}
			go p.Start(r.Host, r.Port)
		}
	}
}
