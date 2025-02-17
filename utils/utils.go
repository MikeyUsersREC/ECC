package utils

import (
	"encoding/json"
	"main/types"
)

func UnmarshalHandler(data []byte) (interface{}, error) {
	var errorResp map[string]interface{}
	if err := json.Unmarshal(data, &errorResp); err == nil {
		if _, hasDetail := errorResp["detail"]; hasDetail {
			return nil, nil
		}
	}

	var getMutualGuilds types.GetMutualGuilds
	if err := json.Unmarshal(data, &getMutualGuilds); err == nil && len(getMutualGuilds.Guilds) > 0 {
		return getMutualGuilds, nil
	}

	var mutualGuilds []types.MutualGuild
	if err := json.Unmarshal(data, &mutualGuilds); err == nil {
		return mutualGuilds, nil
	}

	return nil, nil
}

func ShardCalculator(guild_id int64, total_shard_count int) int {
	return (int(guild_id) >> 22) % total_shard_count
}
