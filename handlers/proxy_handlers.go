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

		return h.handleGuildRequest(c, path)
	}
}

func (h *ProxyHandler) handleMutualGuildsRequest(c *fiber.Ctx, path string) error {
	var instances []types.InstanceInfo
	cursor, err := h.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if err = cursor.All(context.TODO(), &instances); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	guilds, err := h.proxyService.GatherResponses(c, instances, path)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"guilds": guilds})
}

func (h *ProxyHandler) handleGuildRequest(c *fiber.Ctx, path string) error {
	var guildObj types.GuildFetch
	if err := c.BodyParser(&guildObj); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if guildObj.Guild == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Guild ID is required"})
	}

	instance, err := database.FetchInstanceByGuild(*h.collection, guildObj.Guild)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	if instance == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Instance not found"})
	}

	return h.proxyService.ForwardRequest(c, instance, path)
}
