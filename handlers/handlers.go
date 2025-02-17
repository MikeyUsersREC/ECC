package handlers

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"main/services"
)

type Handlers struct {
	Instance *InstanceHandler
	Proxy    *ProxyHandler
}

func NewHandlers(collection *mongo.Collection, baseProject services.Project) *Handlers {
	return &Handlers{
		Instance: NewInstanceHandler(collection, baseProject),
		Proxy:    NewProxyHandler(collection),
	}
}
