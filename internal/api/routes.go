package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/Kushagrakaneki/Quant-Engine/pkg/security"
	"net/http"
)

func SetUpRoutes(r *chi.Mux,jwtManager *security.JWTManager){

	r.Get("/health",func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"Gateway OK"}`))
	})

	r.Group(func(protected chi.Router) {
		protected.Use(JWTMiddleware(jwtManager))
		protected.Post("/orders",PlaceOrderHandler)
	})

}



