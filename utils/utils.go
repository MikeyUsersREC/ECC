package utils

import (
	"fmt"

	"main/types"

	"github.com/bytedance/sonic"
)

func UnmarshalHandler(data []byte) (interface{}, error) {
	var getMutualGuilds types.GetMutualGuilds
	err := sonic.Unmarshal(data, &getMutualGuilds)

	if err == nil && len(getMutualGuilds.Guilds) > 0 {
		return getMutualGuilds, nil
	}

	var mutualGuilds []types.MutualGuild
	err = sonic.Unmarshal(data, &mutualGuilds)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data into supported types: %v", err)
	}

	return mutualGuilds, nil
}

func ShardCalculator(guild_id int64, total_shard_count int) int {
	return (int(guild_id) >> 22) % total_shard_count
}
