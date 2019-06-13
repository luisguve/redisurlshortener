package main

import(
	"os"

	"github.com/villegasl/urlshortener.redis/models"
	"github.com/villegasl/urlshortener.redis/web"
)

func main() {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}

	// redis service
	DB_Handler := models.Start(redisAddr)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// api
	web.Start(DB_Handler, port)
}