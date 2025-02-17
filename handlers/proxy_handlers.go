package handlers

import (
	"context"
	"fmt"
	"main/api"
	"main/database"
	"main/types"
	"strconv"
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

func shardCalculator(guildID int64, totalShardCount int) int {
	return (int(guildID) >> 22) % totalShardCount
}

func (h *ProxyHandler) handleProxyRequest(c *fiber.Ctx, path string) error {
	var reqBody struct {
		Guild interface{} `json:"guild"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(422).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    422,
				"message": "Unprocessable Entity",
			},
		})
	}

	var guildStr string
	switch v := reqBody.Guild.(type) {
	case float64:
		guildStr = fmt.Sprintf("%.0f", v)
	case string:
		guildStr = v
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    400,
				"message": "Invalid guild ID format",
			},
		})
	}

	instance, err := database.FetchInstanceByGuild(*h.collection, guildStr)
	if err != nil || instance == nil {
		guildID, err := strconv.ParseInt(guildStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    400,
					"message": "Invalid guild ID",
				},
			})
		}

		shard := shardCalculator(guildID, 22)
		instance, err = database.FetchInstanceByShard(*h.collection, shard)
		if err != nil || instance == nil {
			return c.Status(404).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    404,
					"message": "No instance found for guild or shard",
				},
			})
		}
	}

	return h.proxyService.ForwardRequest(c, instance, path)
}
