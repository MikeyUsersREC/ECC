package handlers

import (
	"context"
	"main/database"
	"main/services"
	"main/types"
	"os"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type InstanceHandler struct {
	collection       *mongo.Collection
	instanceService  *services.InstanceService
	easyPanelService *services.EasypanelService
}

func NewInstanceHandler(collection *mongo.Collection, baseProject services.Project) *InstanceHandler {
	return &InstanceHandler{
		collection:      collection,
		instanceService: services.NewInstanceService(collection),
		easyPanelService: services.NewEasypanelService(
			os.Getenv("EASYPANEL_API_URL"),
			os.Getenv("EASYPANEL_PROJECT_NAME"),
			os.Getenv("EASYPANEL_AUTH_KEY"),
			baseProject,
		),
	}
}

func (h *InstanceHandler) RegisterInstance() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody types.PartialInstanceInfo
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err})
		}

		instance, err := database.FetchInstanceByInstanceName(*h.collection, requestBody.InstanceName)
		if instance != nil {
			return c.Status(400).JSON(fiber.Map{"error": "instance with provided name exists"})
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err})
		}

		newInstance := h.instanceService.CreateNewInstance(requestBody)
		privateKey, err := h.instanceService.SetupInstanceEncryption(&newInstance, requestBody.Token)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err})
		}

		database.CreateInstanceInfo(*h.collection, newInstance)

		h.easyPanelService.CreateApp(
			newInstance.InstanceName,
			newInstance.InstanceId,
			privateKey,
		)

		return c.JSON(fiber.Map{
			"instance":   newInstance,
			"privateKey": privateKey,
		})
	}
}

func (h *InstanceHandler) GetInstance() fiber.Handler {
	return func(c *fiber.Ctx) error {
		instance, err := database.GetInstanceInfo(*h.collection, bson.M{"instance_id": c.Params("instanceID")})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err})
		}
		return c.JSON(instance)
	}
}

func (h *InstanceHandler) FetchAllInstances() fiber.Handler {
	return func(c *fiber.Ctx) error {
		results, err := h.collection.Find(context.TODO(), bson.M{})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		var instances []*types.InstanceInfo
		if err = results.All(context.TODO(), &instances); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"instances": instances})
	}
}
