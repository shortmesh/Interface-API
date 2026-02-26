package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"interface-api/internal/database"
)

type Server struct {
	port int
	host string
	db   database.Service
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	NewServer := &Server{
		port: port,
		host: host,
		db:   database.New(),
	}

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", NewServer.host, NewServer.port),
		Handler:           NewServer.RegisterRoutes(),
		IdleTimeout:       time.Minute,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	return server
}
