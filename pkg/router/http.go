package router

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"warptail/pkg/utils"

	"tailscale.com/tsnet"
)

type HTTPRoute struct {
	config utils.RouteConfig
	status RouterStatus
	data   *utils.TimeSeries
	*http.Client
}

func NewHTTPRoute(config utils.RouteConfig, server *tsnet.Server) *HTTPRoute {
	return &HTTPRoute{
		config: config,
		data:   utils.NewTimeSeries(time.Second, 1000),
		status: STOPPED,
		Client: server.HTTPClient(),
	}
}

func (route *HTTPRoute) Update(config utils.RouteConfig) error {
	route.config = config
	return nil
}
func (route *HTTPRoute) Start() error {
	route.status = RUNNING
	return nil
}
func (route *HTTPRoute) Stop() error {
	route.status = STOPPED
	return nil
}

func (route *HTTPRoute) Status() RouterStatus {
	return route.status
}

func (route *HTTPRoute) Config() utils.RouteConfig {
	return route.config
}

func (route *HTTPRoute) Stats() utils.TimeSeriesData {
	return route.data.Data
}

func (route *HTTPRoute) copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func parseRequestSize(r *http.Request) (int64, error) {
	contentLength := r.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, nil
	}
	return strconv.ParseInt(contentLength, 10, 64)
}

type customWriter struct {
	w     io.Writer
	total int64 // Keeps track of the total number of bytes written
}

// Write implements the io.Writer interface.
func (cw *customWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.total += int64(n) // Add the number of bytes written to the total
	return n, err
}

func (route *HTTPRoute) Handle(w http.ResponseWriter, r *http.Request) {
	if route.status != RUNNING {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if size, err := parseRequestSize(r); err == nil {
		route.data.LogRecived(uint64(size))
	}

	r.URL.Host = fmt.Sprintf("%s:%d", route.config.Machine.Address, route.config.Machine.Port)
	r.URL.Scheme = "http"
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = route.config.Machine.Address
	r.RequestURI = ""

	route.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := route.Do(r)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	cw := &customWriter{w: w}

	route.copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(cw, resp.Body)
	route.data.LogSent(uint64(cw.total))
}
