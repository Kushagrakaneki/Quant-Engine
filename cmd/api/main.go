package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Kushagrakaneki/Quant-Engine/internal/config"
	"github.com/Kushagrakaneki/Quant-Engine/pkg/security"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"quant-engine/internal/config"
	"quant-engine/pkg/security"
)

func main(){

	cfg,err:=config.LoadConfig()
	if err != nil {
		log.Fatalf("Server failed to boot: %v", err)
	}

	jwtManager:=security.NewJWTManager(
		cfg.JWTSecret,
		cfg.JWTExpirationHours,
	)

	_=jwtManager





	//new router 
	r:=chi.NewRouter()

	//global middleware
	r.Use(middleware.RequestID)//gives new id to a request
	r.Use(middleware.RealIP)//obtains ip of the requester
	r.Use(middleware.Logger)//writes about request
	r.Use(middleware.Recoverer)//handles any panic problem to vaoid crashing whole server

	r.Use(middleware.Timeout(60*time.Second))//A request shouldn't run longer than 60 seconds.

	r.Get("/health",func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"Quant-Lite Matcher is ALIVE"}`))
	})

	port:=":"+cfg.Port
	fmt.Printf("Starting Quant-Lite Trading Engine on port %s...\n", cfg.Port)

	err=http.ListenAndServe(port,r)

	if err!=nil{
		log.Fatalf("CRITICAL: Server crashed: %v\n", err)
	}


}