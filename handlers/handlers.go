package handlers

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Handlers struct {
	Instance *InstanceHandler
	Proxy    *ProxyHandler
}

func NewHandlers(collection *mongo.Collection) *Handlers {
	return &Handlers{
		Instance: NewInstanceHandler(collection),
		Proxy:    NewProxyHandler(collection),
	}
}
