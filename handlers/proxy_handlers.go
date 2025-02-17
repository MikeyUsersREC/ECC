package handlers

import (
	"context"
	"main/api"
	"main/database"
	"main/types"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProxyHandler struct {
	collection   *mongo.Collection
	proxyService *api.ProxyService
}

func NewProxyHandler(collection *mongo.Collection) *ProxyHandler {
	return &ProxyHandler{
		collection:   collection,
		proxyService: api.NewProxyService(),
	}
}

func (h *ProxyHandler) APIProxy() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		if strings.Contains(path, "get_mutual_guilds") || strings.Contains(path, "get_staff_guilds") {
			return h.handleMutualGuildsRequest(c, path)
		}

		return h.handleProxyRequest(c, path)
	}
}

func (h *ProxyHandler) handleMutualGuildsRequest(c *fiber.Ctx, path string) error {
	var instances []types.InstanceInfo
	cursor, err := h.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    500,
				"message": "Internal Server Error",
			},
		})
	}

	if err = cursor.All(context.TODO(), &instances); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    500,
				"message": "Internal Server Error",
			},
		})
	}

	guilds, err := h.proxyService.GatherResponses(c, instances, path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    500,
				"message": err.Error(),
			},
		})
	}

	return c.JSON(fiber.Map{"guilds": guilds})
}

func (h *ProxyHandler) handleProxyRequest(c *fiber.Ctx, path string) error {
	var reqBody struct {
		Guild string `json:"guild"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(422).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    422,
				"message": "Unprocessable Entity",
			},
		})
	}

	if reqBody.Guild == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    400,
				"message": "Guild ID is required",
			},
		})
	}

	instance, err := database.FetchInstanceByGuild(*h.collection, reqBody.Guild)
	if err != nil || instance == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    404,
				"message": "Instance not found",
			},
		})
	}

	return h.proxyService.ForwardRequest(c, instance, path)
}
