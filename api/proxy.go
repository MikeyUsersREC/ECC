package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"

	"main/types"
	"main/utils"
)

type ProxyService struct {
	client *http.Client
}

func NewProxyService() *ProxyService {
	return &ProxyService{
		client: &http.Client{},
	}
}

func (p *ProxyService) ForwardRequest(c *fiber.Ctx, instance *types.InstanceInfo, path string) error {
	if instance == nil {
		return c.Status(422).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    422,
				"message": "Unprocessable Entity",
			},
		})
	}

	if path == "" {
		return c.Status(422).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    422,
				"message": "Unprocessable Entity",
			},
		})
	}

	targetURL := fmt.Sprintf("%s://%s%s",
		strings.ToLower(instance.Protocol),
		instance.Hostname,
		strings.Replace(path, "api/", "", 1))

	resp, err := p.makeRequest(c, targetURL)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusBadGateway,
				"message": err.Error(),
			},
		})
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    fiber.StatusInternalServerError,
				"message": "Failed to read response body",
			},
		})
	}

	c.Status(resp.StatusCode)
	c.Set("Content-Type", resp.Header.Get("Content-Type"))
	return c.Send(body)
}

func (p *ProxyService) GatherResponses(c *fiber.Ctx, instances []types.InstanceInfo, path string) ([]types.MutualGuild, error) {
	var targets []string
	for _, instance := range instances {
		targets = append(targets, fmt.Sprintf("%s://%s%s",
			strings.ToLower(instance.Protocol),
			instance.Hostname,
			strings.Replace(path, "api/", "", 1)))
	}

	responses, err := p.sendRequestToTargets(c, targets)
	if err != nil {
		return nil, err
	}

	return p.consolidateResponses(responses), nil
}

func (p *ProxyService) sendRequestToTargets(c *fiber.Ctx, targets []string) ([]interface{}, error) {
	var bodies []interface{}

	for _, target := range targets {
		body, err := p.sendRequest(c, target)
		if err != nil {
			return nil, err
		}

		newData, err := utils.UnmarshalHandler(body)
		if err != nil {
			log.Fatal(err)
		}
		bodies = append(bodies, newData)
	}

	return bodies, nil
}

func (p *ProxyService) sendRequest(c *fiber.Ctx, target string) ([]byte, error) {
	request, err := http.NewRequest(c.Method(), target, bytes.NewBuffer(c.Body()))
	if err != nil {
		return nil, err
	}

	origHeaders := c.GetReqHeaders()
	for key, values := range origHeaders {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	resp, err := p.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (p *ProxyService) makeRequest(c *fiber.Ctx, target string) (*http.Response, error) {
	request, err := http.NewRequest(c.Method(), target, bytes.NewBuffer(c.Body()))
	if err != nil {
		return nil, err
	}

	origHeaders := c.GetReqHeaders()
	for key, values := range origHeaders {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	return p.client.Do(request)
}

func (p *ProxyService) consolidateResponses(responses []interface{}) []types.MutualGuild {
	var combined []types.MutualGuild

	for _, resp := range responses {
		switch v := resp.(type) {
		case types.GetMutualGuilds:
			for _, guild := range v.Guilds {
				if !containsGuild(combined, guild) {
					combined = append(combined, guild)
				}
			}
		case []types.MutualGuild:
			for _, guild := range v {
				if !containsGuild(combined, guild) {
					combined = append(combined, guild)
				}
			}
		}
	}

	return combined
}

func containsGuild(guilds []types.MutualGuild, guild types.MutualGuild) bool {
	for _, g := range guilds {
		if g.ID == guild.ID {
			return true
		}
	}
	return false
}
