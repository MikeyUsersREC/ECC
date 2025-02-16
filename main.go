package main

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"main/database"
	"main/handlers"
	"main/types"
)

func loadEnvironment() *types.EccConfig {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var ecc types.EccConfig
	if err := env.Parse(&ecc); err != nil {
		log.Fatal(err)
		return nil
	}
	return &ecc
}

func main() {
	log.Info("Starting database connection ...")

	environment := loadEnvironment()
	log.Info("Loaded environment ...")

	client, err := database.InitClient(environment.MongoURL)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(environment.DatabaseName).Collection("Instances")

	handlers := handlers.NewHandlers(collection)

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "OK"})
	})

	app.Get("/instance/:instanceID", handlers.Instance.GetInstance())
	app.Get("/all", handlers.Instance.FetchAllInstances())
	app.Post("/create", handlers.Instance.RegisterInstance())
	app.Use("/api/*", handlers.Proxy.APIProxy())

	listening_host := os.Getenv("LISTEN_HOST")
	listening_port := os.Getenv("LISTEN_PORT")

	if listening_host == "" || listening_port == "" {
		listening_host = "0.0.0.0"
		listening_port = "22516"
	}

	log.Fatal(app.Listen(listening_host + ":" + listening_port))
}
