package database

import (
	"context"
	"errors"
	"log"
	"math"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/types"
)

func InitClient(mongoURL string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func GetInstanceInfo(collection mongo.Collection, filter bson.M) (*types.InstanceInfo, error) {
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	var instance types.InstanceInfo
	if err := result.Decode(&instance); err != nil {
		return nil, err
	}
	return &instance, nil
}

func CreateInstanceInfo(collection mongo.Collection, info types.InstanceInfo) {
	_, err := collection.InsertOne(context.TODO(), info)
	if err != nil {
		log.Fatal(err)
	}
}

func AssignShardIdentifiers(collection mongo.Collection, instanceType string, totalShardCount int) (map[string][]int, error) {
	documentCursor, err := collection.Find(context.TODO(), bson.M{
		"instance_type":   instanceType,
		"instance_status": 0,
	})
	if err != nil {
		return nil, err
	}
	if documentCursor == nil {
		return nil, errors.New("no documents found")
	}

	var floatedShardCount float32 = float32(totalShardCount)

	var results []types.InstanceInfo
	if err := documentCursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	var shardsPerDocument int = int(math.Ceil(float64(floatedShardCount / float32(len(results)))))
	var documents map[string][]int = map[string][]int{}

	for shard, document := 0, 0; float32(shard) <= floatedShardCount && document < len(results); document, shard = document+1, shard+shardsPerDocument {
		shardsToCompleteTo := shard + shardsPerDocument - 1

		info := results[document]
		var shardSlice []int
		for i := shard; i <= shardsToCompleteTo && i <= totalShardCount; i++ {
			shardSlice = append(shardSlice, i)
		}

		_, err := collection.UpdateOne(
			context.TODO(),
			bson.M{"instance_id": info.InstanceId},
			bson.M{"$set": bson.M{"shard_ids": shardSlice}},
		)
		if err != nil {
			return nil, err
		}

		documents[info.InstanceId] = shardSlice
	}
	return documents, nil
}

func DeleteAllInstances(collection mongo.Collection) error {
	_, err := collection.DeleteMany(context.TODO(), bson.D{})
	return err
}

func FetchInstanceByInstanceName(collection mongo.Collection, instanceName string) (*types.InstanceInfo, error) {
	return GetInstanceInfo(collection, bson.M{"instance_name": instanceName})
}

func FetchInstanceByGuild(collection mongo.Collection, guildID string) (*types.InstanceInfo, error) {
	return GetInstanceInfo(collection, bson.M{"guild_ids": guildID})
}

func FetchInstanceByShard(collection mongo.Collection, shardID int) (*types.InstanceInfo, error) {
	return GetInstanceInfo(collection, bson.M{"shard_ids": shardID})
}
