package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/notebase/book"
	"github.com/sikozonpc/notebase/config"
	"github.com/sikozonpc/notebase/highlight"
	"github.com/sikozonpc/notebase/medium"
	"github.com/sikozonpc/notebase/storage"
	"github.com/sikozonpc/notebase/user"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIServer struct {
	addr string
	db   *mongo.Client
}

func NewAPIServer(addr string, db *mongo.Client) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	ctx := context.Background()

	gcpStorage, err := storage.NewGCPStorage(ctx)
	if err != nil {
		log.Fatal(err)
	}

	mailer := medium.NewMailer(config.Envs.SendGridAPIKey, config.Envs.SendGridFromEmail)

	bookStore := book.NewStore(s.db)

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)

	highlightStore := highlight.NewStore(s.db)
	highlightHandler := highlight.NewHandler(highlightStore, userStore, gcpStorage, bookStore, mailer)
	highlightHandler.RegisterRoutes(subrouter)

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	log.Println("Listening on", s.addr)
	log.Println("Process PID", os.Getpid())

	env := config.Envs.Env
	if env == "development" {
		v := reflect.ValueOf(config.Envs)

		for i := 0; i < v.NumField(); i++ {
			log.Println(v.Type().Field(i).Name, "=", v.Field(i).Interface())
		}
	}

	return http.ListenAndServe(s.addr, router)
}
