package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kazakhjuice/tech-assignment-medods/internal/handler"
	repository "github.com/kazakhjuice/tech-assignment-medods/internal/repo"
	"github.com/kazakhjuice/tech-assignment-medods/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Run() {

	godotenv.Load()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Print(err)
		}
	}()

	err = client.Ping(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	db := client.Database("techassignment").Collection("tokens")
	jwtKey := os.Getenv("JWT_KEY")
	repo := repository.NewRepo(db)
	services := service.NewService(repo, jwtKey)
	handlers := handler.NewHandler(services)

	router := mux.NewRouter()

	router.HandleFunc("/login", handlers.GetTokens).Methods("GET")
	router.HandleFunc("/update", handlers.UpdateRefreshToken).Methods("PATCH")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Print(err)
		}
	}()

	log.Print("Server started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := server.Shutdown(ctx); err != nil {
		log.Print(err)

		return
	}
}
