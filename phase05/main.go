package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"traceroute/handler"
	"traceroute/helper"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: failed to load .env file")
	} else {
		fmt.Println("Successfully loaded .env file")
	}

	router := gin.Default()
	router.GET("/traceroute/:host", handler.Trace)

	// Start the server
	serverHost := helper.GetEnv("SERVER_HOST", ":8080")
	fmt.Println("Starting server")
	if err := router.Run(serverHost); err != nil {
		fmt.Println("Unable to start server")
	}
}
