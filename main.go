package main

import (
	"log"

	"os"

	"github.com/ayo-ajayi/ecommerce/internal/constants"
	router "github.com/ayo-ajayi/ecommerce/internal/routes"
	"github.com/ayo-ajayi/ecommerce/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading .env file", err)
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = constants.ServerPort
	}
	handler := router.NewRouter()
	server.NewServer(":"+port, handler).Start()
}
