package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	urlShortener "url_shortner/internal/urlShortner"
	user "url_shortner/internal/user"

	"url_shortner/internal/middleware"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Logging the current request")
		next.ServeHTTP(w, r)
		fmt.Println("Request finished")
	})
}
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env found")
	}
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		fmt.Printf("Database is not reachable: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected to the database!")
	repo := urlShortener.NewRepository(db)
	handler := &urlShortener.Handler{Repo: repo}

	userRepo := user.NewRepository(db)
	userHandler := &user.Handler{
		Repo:      userRepo,
		JWTSecret: jwtSecret,
	}

	r := chi.NewRouter()
	r.Use(LoggerMiddleware)
	r.Get("/{code}", handler.GetUrl)
	r.Post("/register", userHandler.Register)
	r.Post("/login", userHandler.Login)
	r.Group(func(r chi.Router) {
		r.Use(middleware.JwtAuth(jwtSecret))

		r.Post("/", handler.AddUrl)
		r.Get("/my-urls", handler.GetMyUrls)
	})

	serv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		fmt.Println("Server Starting on  :8080")
		if err := serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %s\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Signal Recieved , Shutting Down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serv.Shutdown(ctx); err != nil {
		log.Fatal("Server Forced to shutdown : ", err)
	}
	fmt.Println("Closing Database...")
	db.Close()
	fmt.Println("Bye!")
}
