package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"warptail/pkg/router"
	"warptail/pkg/utils"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type apiCtx string

const ROUTECTX = apiCtx("route")

type api struct {
	*router.Router
	*chi.Mux
	config utils.DashboardConfig
}

var spa = SPAHandler{
	StaticPath: "./dashboard/dist",
	IndexPath:  "index.html",
}

func NewApi(router *router.Router, config utils.DashboardConfig) *api {
	api := api{
		Router: router,
		Mux:    chi.NewRouter(),
		config: config,
	}
	// Add middlewares
	api.Mux.Use(middleware.RequestID)
	api.Mux.Use(middleware.RealIP)
	api.Mux.Use(middleware.Logger)
	api.Mux.Use(middleware.Recoverer)
	api.Mux.Use(middleware.Compress(5))

	api.Mux.Use(api.proxy)

	api.Mux.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool { return true },
		AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:  []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	api.Mux.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Handle all other requests by serving the index.html
	api.Mux.Get("/*", spa.ServeHTTP)

	api.Mux.Post("/auth/login", api.loginHandler)

	api.Mux.Group(func(r chi.Router) {
		r.Use(TokenAuthMiddleware)
		r.Route("/api/settings", func(r chi.Router) {
			r.Get("/tailscale", api.handleTailscaleSettings)
			r.Post("/tailscale", api.handleUpdateTailscaleSettings)

			r.Get("/dashboard", api.handleDashboardSettings)
			r.Post("/dashboard", api.handleUpdateDashboardSettings)
		})
		r.Route("/api/routes", func(r chi.Router) {
			r.Get("/", api.handleGetRoutes)
			r.Post("/", api.handleCreateRoute)
		})
		r.Route("/api/routes/{routeID}", func(r chi.Router) {
			r.Use(api.RouteCtx)
			r.Get("/", api.handleGetRoute)
			r.Get("/timeseries", api.handleTimeseries)
			r.Post("/stop", api.handleStopRoute)
			r.Post("/start", api.handleStartRoute)
			r.Put("/", api.handleUpdateRoute)
			r.Delete("/", api.handleDeleteRoute)
		})
	})

	return &api
}

func (api *api) proxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
			host = host[:colonIndex]
		}

		route, err := api.GetRouteByName(host)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if route.Config().Type != utils.HTTP {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		httpRoute := route.(*router.HTTPRoute)
		httpRoute.Handle(w, r)
	})
}

func (api *api) Start(addr string) {
	log.Println("Starting API on http://localhost:8080")
	log.Println(http.ListenAndServe(addr, api))
}

func (api *api) RouteCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routeID := chi.URLParam(r, "routeID")
		route, err := api.Router.Get(routeID)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), ROUTECTX, route)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api *api) handleGetRoutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.GetAll())
}

func (api *api) handleCreateRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var config utils.RouteConfig
	decoder.Decode(&config)
	route, err := api.AddRoute(config)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(route.Config())
}

func (api *api) handleGetRoute(w http.ResponseWriter, r *http.Request) {
	route, ok := r.Context().Value(ROUTECTX).(router.RouteInfo)
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(route)
}

func (api *api) handleTimeseries(w http.ResponseWriter, r *http.Request) {
	route, ok := r.Context().Value(ROUTECTX).(router.RouteInfo)
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(route.Stats)
}

func (api *api) handleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var route utils.RouteConfig
	decoder.Decode(&route)
	api.UpdateRoute(route)
	json.NewEncoder(w).Encode(api.GetRoute(route.Id).Config())
}

func (api *api) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	route, ok := r.Context().Value(ROUTECTX).(router.RouteInfo)
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	api.DeleteRoute(route.Id)
}

func (api *api) handleStartRoute(w http.ResponseWriter, r *http.Request) {
	route, ok := r.Context().Value(ROUTECTX).(router.RouteInfo)
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	api.StartRoute(route.Id)
}

func (api *api) handleStopRoute(w http.ResponseWriter, r *http.Request) {
	route, ok := r.Context().Value(ROUTECTX).(router.RouteInfo)
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	api.StopRoute(route.Id)
}

func (api *api) handleTailscaleSettings(w http.ResponseWriter, r *http.Request) {
	config := utils.LoadConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config.Tailscale)
}

func (api *api) handleUpdateTailscaleSettings(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var tsc utils.TailscaleConfig
	decoder.Decode(&tsc)

	config := utils.LoadConfig()
	config.Tailscale = tsc
	api.UpdateTailScale(tsc)
	utils.Save(config)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config.Tailscale)
}

func (api *api) handleDashboardSettings(w http.ResponseWriter, r *http.Request) {
	config := utils.LoadConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config.Dasboard)
}

func (api *api) handleUpdateDashboardSettings(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var tsc utils.DashboardConfig
	decoder.Decode(&tsc)

	config := utils.LoadConfig()
	config.Dasboard = tsc
	utils.Save(config)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config.Dasboard)
}
