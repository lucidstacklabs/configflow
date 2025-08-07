package main

import (
	"github.com/lucidstacklabs/configflow/internal/app/configflow"
	"github.com/lucidstacklabs/configflow/internal/pkg/env"
)

func main() {
	configflow.NewServer(&configflow.ServerConfig{
		Host:          env.GetOrDefault("HOST", "0.0.0.0"),
		Port:          env.GetOrDefault("PORT", "5000"),
		MongoEndpoint: env.GetOrDefault("MONGO_ENDPOINT", "mongodb://localhost:27017"),
		MongoDatabase: env.GetOrDefault("MONGO_DATABASE", "configflow"),
		JwtSigningKey: env.GetOrDefault("JWT_SIGNING_KEY", "secret"),
		JwtIssuer:     env.GetOrDefault("JWT_ISSUER", "configflow"),
		JwtAudience:   env.GetOrDefault("JWT_AUDIENCE", "configflow"),
	}).Start()
}
