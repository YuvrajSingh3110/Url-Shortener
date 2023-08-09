package main

import (
	"fmt"
	"log"
	"os"

	"github.com/YuvrajSingh3110/Url_Shortener/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func SetUpRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveUrl)
	app.Post("/api/v1", routes.ShortenUrl)
}

func main() {
	err := godotenv.Load() //load variables defined in .env file
	if err != nil {
		fmt.Println(err)
	}

	app := fiber.New()

	app.Use(logger.New())

	SetUpRoutes(app)

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
