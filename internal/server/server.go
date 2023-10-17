package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	server *http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

func (s *Server) Start() {
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			log.Println("Server shutdown error: ", err)
			return
		}
		log.Println("Server stopped gracefully")
	}()
	log.Println("Server is running🎉🎉. Press Ctrl+C to stop")
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		log.Println("Server error: ", err)
		return
	}

}
