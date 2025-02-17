package main

import (
	"errors"
	"github.com/bytedance/sonic"
	"io/ioutil"
	"main/services"
	"net/http"
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

	err := os.Mkdir("keys", 0644)
	if err != nil {
		return
	}

	environment := loadEnvironment()
	log.Info("Loaded environment ...")

	client, err := database.InitClient(environment.MongoURL)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(environment.DatabaseName).Collection("Instances")

	if _, err := os.Stat(os.Getenv("EXAMPLE_PROJECT_FILE_NAME")); errors.Is(err, os.ErrNotExist) {
		client := http.Client{}
		request, err := http.NewRequest("GET", os.Getenv("EXAMPLE_PROJECT_DOWNLOAD"), nil)
		request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36") // to satisfy websites which check..
		if err != nil {
			log.Fatal(err)
		}

		resp, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var data map[string]interface{}
		if err := sonic.Unmarshal(body, &data); err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile(os.Getenv("EXAMPLE_PROJECT_FILE_NAME"), body, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	project_data, _ := ioutil.ReadFile(os.Getenv("EXAMPLE_PROJECT_FILE_NAME"))
	var baseProject services.Project
	err = sonic.Unmarshal(project_data, &baseProject)
	if err != nil {
		log.Fatal(err)
	}

	handlers := handlers.NewHandlers(collection, baseProject)

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
