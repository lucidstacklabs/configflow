package configflow

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lucidstacklabs/configflow/internal/app/configflow/admin"
	"github.com/lucidstacklabs/configflow/internal/app/configflow/apikey"
	"github.com/lucidstacklabs/configflow/internal/app/configflow/health"
	"github.com/lucidstacklabs/configflow/internal/pkg/auth"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	config *ServerConfig
}

func NewServer(config *ServerConfig) *Server {
	return &Server{config: config}
}

type ServerConfig struct {
	Host          string
	Port          string
	MongoEndpoint string
	MongoDatabase string
	JwtSigningKey string
	JwtIssuer     string
	JwtAudience   string
}

func (s *Server) Start() {
	// Database setup

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(s.config.MongoEndpoint))

	if err != nil {
		log.Fatal("error while connecting to mongo database: ", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal("error while pinging mongo database: ", err)
	}

	mongoDatabase := client.Database(s.config.MongoDatabase)

	// Services setup

	apiKeysCollection := mongoDatabase.Collection("api_keys")
	authenticator := auth.NewAuthenticator(s.config.JwtSigningKey, s.config.JwtIssuer, s.config.JwtAudience, apiKeysCollection)
	adminService := admin.NewService(mongoDatabase.Collection("admins"), authenticator)
	apiKeyService := apikey.NewService(apiKeysCollection)

	// API server setup

	router := gin.Default()

	health.NewHandler(router).Register()
	admin.NewHandler(router, authenticator, adminService).Register()
	apikey.NewHandler(router, authenticator, apiKeyService).Register()

	log.Printf("starting api server on %s:%s", s.config.Host, s.config.Port)

	err = router.Run(fmt.Sprintf("%s:%s", s.config.Host, s.config.Port))

	if err != nil {
		log.Fatal("error starting api server: ", err)
	}
}
