package services

import (
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
	ProjectName string  `json:"project_name"`
	ServiceName string  `json:"service_name"`
	Source      Source  `json:"source"`
	Build       Build   `json:"build"`
	Env         string  `json:"env"`
	Deploy      Deploy  `json:"deploy"`
	Domains     Domains `json:"domains"`
}

type Build struct {
	Type            string `json:"type"`
	NixpacksVersion string `json:"nixpacks_version"`
}

type Deploy struct {
	Replicas     int    `json:"replicas"`
	Command      string `json:"command"`
	ZeroDowntime bool   `json:"zero_downtime"`
}

type Domains struct {
	Host             string `json:"host"`
	Https            bool   `json:"https"`
	Port             int    `json:"port"`
	Path             string `json:"path"`
	Wildcard         bool   `json:"wildcard"`
	InternalProtocol bool   `json:"internal_protocol"`
}

type Source struct {
	Type       string `json:"type"`
	Owner      string `json:"owner"`
	Repo       string `json:"repo"`
	Ref        string `json:"ref"`
	Path       string `json:"path"`
	Autodeploy bool   `json:"autoDeploy"`
}

func NewEasypanelService(APIUrl string, Project string, AuthKey string, MemAlloc int, BaseProject Project) *EasypanelService {
	return &EasypanelService{
		APIUrl:      APIUrl,
		Project:     Project,
		AuthKey:     AuthKey,
		MemAlloc:    MemAlloc,
		BaseProject: BaseProject,
	}
}

func (service *EasypanelService) CreateApp(serviceName string) {
	var newProject Project = service.BaseProject
	newProject.Data.Domains.Host = strings.Replace(newProject.Data.Domains.Host, os.Getenv("DEFAULT_PROJECT_NAME"), serviceName, 1)
	// TODO: change env var using env name | split env by line, find starting variable, split by equals, split to \n, and then replace that value, recompile
}
