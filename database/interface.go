package database

import (
	"main/types"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type InstanceStore interface {
	GetInstance(id string) (*types.InstanceInfo, error)
	CreateInstance(info types.InstanceInfo) error
	FindByName(name string) (*types.InstanceInfo, error)
	FindByGuild(guildID string) (*types.InstanceInfo, error)
	FindByShard(shardID int) (*types.InstanceInfo, error)
	DeleteAll() error
	AssignShards(instanceType string, totalShardCount int) (map[string][]int, error)
}

type MongoStore struct {
	collection *mongo.Collection
}

func NewMongoStore(collection *mongo.Collection) InstanceStore {
	return &MongoStore{collection: collection}
}

func (s *MongoStore) GetInstance(id string) (*types.InstanceInfo, error) {
	return GetInstanceInfo(*s.collection, bson.M{"instance_id": id})
}

func (s *MongoStore) CreateInstance(info types.InstanceInfo) error {
	CreateInstanceInfo(*s.collection, info)
	return nil
}

func (s *MongoStore) FindByName(name string) (*types.InstanceInfo, error) {
	return FetchInstanceByInstanceName(*s.collection, name)
}

func (s *MongoStore) FindByGuild(guildID string) (*types.InstanceInfo, error) {
	return FetchInstanceByGuild(*s.collection, guildID)
}

func (s *MongoStore) FindByShard(shardID int) (*types.InstanceInfo, error) {
	return FetchInstanceByShard(*s.collection, shardID)
}

func (s *MongoStore) DeleteAll() error {
	return DeleteAllInstances(*s.collection)
}

func (s *MongoStore) AssignShards(instanceType string, totalShardCount int) (map[string][]int, error) {
	return AssignShardIdentifiers(*s.collection, instanceType, totalShardCount)
}
