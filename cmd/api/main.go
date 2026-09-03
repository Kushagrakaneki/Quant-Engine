package main

import (
	
	"log"
	"net/http"
	"time"

	"github.com/Kushagrakaneki/Quant-Engine/internal/config"
	
	"github.com/Kushagrakaneki/Quant-Engine/pkg/security"
	"github.com/Kushagrakaneki/Quant-Engine/internal/api"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

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

	api.SetUpRoutes(r,jwtManager)

	srv:=&http.Server{
		Addr:":"+cfg.Port,
		Handler:r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("Quant-Lite Trading Gateway online on port %s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}


}