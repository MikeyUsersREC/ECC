package services

import (
	"bytes"
	"github.com/bytedance/sonic"
	"github.com/charmbracelet/log"
	"io"
	"net/http"
	"os"
	"strings"
)

type EasypanelService struct {
	APIUrl      string
	Project     string
	AuthKey     string
	MemAlloc    int
	BaseProject Project
}

type Project struct {
	Type string `json:"type"`
	Data Data   `json:"data"`
}

type Data struct {
	ProjectName string    `json:"projectName"`
	ServiceName string    `json:"serviceName"`
	Source      Source    `json:"source"`
	Build       Build     `json:"build"`
	Env         string    `json:"env"`
	Deploy      Deploy    `json:"deploy"`
	Domains     []Domains `json:"domains"`
}

type Build struct {
	Type            string `json:"type"`
	NixpacksVersion string `json:"nixpacksVersion"`
}

type Deploy struct {
	Replicas     int    `json:"replicas"`
	Command      string `json:"command"`
	ZeroDowntime bool   `json:"zeroDowntime"`
}

type Domains struct {
	Host             string `json:"host"`
	Https            bool   `json:"https"`
	Port             int    `json:"port"`
	Path             string `json:"path"`
	Wildcard         bool   `json:"wildcard"`
	InternalProtocol string `json:"internalProtocol"`
}

type Source struct {
	Type       string `json:"type"`
	Owner      string `json:"owner"`
	Repo       string `json:"repo"`
	Ref        string `json:"ref"`
	Path       string `json:"path"`
	Autodeploy bool   `json:"autoDeploy"`
}

type PublishingInterface struct {
	Json Data `json:"json"`
}

type DeployBody struct {
	ProjectName string `json:"projectName"`
	ServiceName string `json:"serviceName"`
}

type DeployInterface struct {
	Json DeployBody `json:"json"`
}

func NewEasypanelService(APIUrl string, Project string, AuthKey string, BaseProject Project) *EasypanelService {
	return &EasypanelService{
		APIUrl:      APIUrl,
		Project:     Project,
		AuthKey:     AuthKey,
		BaseProject: BaseProject,
	}
}

func (service *EasypanelService) CreateApp(serviceName string, instanceId string, privateKeyString string) {
	var newProject Project = service.BaseProject

	serviceName = strings.ToLower(serviceName)
	newProject.Data.Domains[0].Host = strings.Replace(newProject.Data.Domains[0].Host, service.Project, serviceName, 1)
	newProject.Data.ServiceName = serviceName
	newProject.Data.ProjectName = service.Project

	var envLines []string = strings.Split(newProject.Data.Env, "\n")
	var newEnvLines []string
	for _, line := range envLines {
		var newLineParts = strings.SplitN(line, "=", 2)

		if strings.HasPrefix(line, os.Getenv("ECC_TOKEN_ENV_VAR")) {
			newLineParts[1] = instanceId
		} else if strings.HasPrefix(line, os.Getenv("RSA_CERT_ENV_VAR")) {
			newLineParts[1] = strings.Replace(privateKeyString, "\n", "\r", -1)
		}
		newEnvLines = append(newEnvLines, strings.Join(newLineParts, "="))
	}
	newEnv := strings.Join(newEnvLines, "\n")
	newEnv = strings.Replace(newEnv, "\r", "\\n", -1)

	newProject.Data.Env = newEnv
	client := http.Client{}

	var data PublishingInterface
	data.Json = newProject.Data

	raw_data, _ := sonic.Marshal(&data)
	reader := bytes.NewReader(raw_data)

	request, err := http.NewRequest(http.MethodPost, service.APIUrl+"/api/trpc/services.app.createService", reader)
	request.Header.Add("Authorization", service.AuthKey)
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if resp.StatusCode != 200 {
		bytedata, _ := io.ReadAll(resp.Body)
		log.Fatal(string(bytedata))
	}
	service.DeployApp(serviceName)

	if err != nil {
		log.Fatal(err)
	}
}

func (service *EasypanelService) DeployApp(serviceName string) {
	var data DeployInterface
	data.Json = DeployBody{
		ProjectName: service.Project,
		ServiceName: serviceName,
	}

	raw_data, _ := sonic.Marshal(&data)
	client := http.Client{}
	reader := bytes.NewReader(raw_data)
	request, err := http.NewRequest(http.MethodPost, service.APIUrl+"/api/trpc/services.app.deployService", reader)
	request.Header.Add("Authorization", service.AuthKey)
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)
	if resp.StatusCode != 200 {
		bytedata, _ := io.ReadAll(resp.Body)
		log.Fatal(string(bytedata))
	}

	if err != nil {
		log.Fatal(err)
	}
}
