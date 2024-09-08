// // The tshello server demonstrates how to use Tailscale as a library.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

var router *Router

func main() {
	config := LoadConfig()
	r, err := NewRouter(config)
	if err != nil {
		log.Fatalf("unable to start %v", err)
	}
	router = r
	defer router.Close()
	router.Start()
	api()
}

func api() {

	r := chi.NewRouter()

	// Add middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool { return true },
		AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:  []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	r.Group(func(r chi.Router) {
		r.Route("/api/routes", func(r chi.Router) {
			r.Get("/", router.handleGetRoutes)
			r.Post("/", router.handleCreateRoute)
		})
		r.Route("/api/routes/{routeID}", func(r chi.Router) {
			r.Use(router.RouteCtx)
			r.Get("/", router.handleGetRoute)
			r.Post("/stop", router.handleStopRoute)
			r.Post("/start", router.handleStartRoute)
			r.Put("/", router.handleUpdateRoute)
			r.Delete("/", router.handleDeleteRoute)
		})
	})

	log.Println("Starting API on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func (router *Router) RouteCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routeID := chi.URLParam(r, "routeID")
		route, err := router.Get(routeID)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "route", route)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (router *Router) handleGetRoutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(router.GetAll())
}

func (router *Router) handleCreateRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var route Route
	decoder.Decode(&route)
	router.AddRoute(route)
	json.NewEncoder(w).Encode(route)
}

func (router *Router) handleGetRoute(w http.ResponseWriter, r *http.Request) {
	route, ok := r.Context().Value("route").(*Route)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(route)
}

func (router *Router) handleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var route Route
	decoder.Decode(&route)
	router.UpdateRoute(route)
	json.NewEncoder(w).Encode(route)
}

func (router *Router) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	route, ok := r.Context().Value("route").(*Route)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	router.DeleteRoute(route.Name)
}

func (router *Router) handleStartRoute(w http.ResponseWriter, r *http.Request) {
	route, ok := r.Context().Value("route").(*Route)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	router.StartRoute(route.Name)
}

func (router *Router) handleStopRoute(w http.ResponseWriter, r *http.Request) {
	route, ok := r.Context().Value("route").(*Route)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	router.StopRoute(route.Name)
}
