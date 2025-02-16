package types

type InstanceInfo struct {
	InstanceId     string   `bson:"instance_id"`
	InstanceName   string   `bson:"instance_name"`
	InstanceType   string   `bson:"instance_type"`
	InstanceStatus int      `bson:"instance_status"`
	EncryptedToken []byte   `bson:"encrypted_token"`
	SHA256Hash     []byte   `bson:"sha256_hash"`
	ShardIds       []int    `bson:"shard_ids"`
	GuildIds       []string `bson:"guild_ids"`
	Hostname       string   `bson:"hostname"`
	Protocol       string   `bson:"protocol"`
}

type PartialInstanceInfo struct {
	InstanceName string
	InstanceType string
	Token        string
	Hostname     string
	Protocol     string
}

type EccConfig struct {
	MongoURL     string `env:"MONGO_URL"`
	DatabaseName string `env:"MONGO_DB_NAME"`
}
