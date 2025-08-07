package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lucidstacklabs/configflow/internal/pkg/apikey"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Authenticator struct {
	jwtSigningKey    []byte
	issuer           string
	audience         string
	apiKeyCollection *mongo.Collection
}

func NewAuthenticator(jwtSigningKey string, issuer string, audience string, apiKeyCollection *mongo.Collection) *Authenticator {
	return &Authenticator{jwtSigningKey: []byte(jwtSigningKey), issuer: issuer, audience: audience, apiKeyCollection: apiKeyCollection}
}

const (
	BearerToken    = "Bearer"
	TypeAdminToken = "admin"
	ApiKey         = "ApiKey"
)

func (a *Authenticator) GenerateAdminToken(adminID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  adminID,
		"iss":  a.issuer,
		"aud":  a.audience,
		"type": TypeAdminToken,
		"iat":  int(time.Now().Unix()),
		"exp":  int(time.Now().Add(time.Hour * 24).Unix()),
	})

	tokenString, err := token.SignedString(a.jwtSigningKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *Authenticator) ValidateApiKeyContext(c *gin.Context, ctx context.Context) (*AuthenticatedApiKey, error) {
	tokenType, token, err := a.extractToken(c)

	if err != nil {
		return nil, err
	}

	switch tokenType {
	case ApiKey:
		return a.validateApiKey(ctx, token)
	default:
		return nil, fmt.Errorf("invalid api key type")
	}
}

func (a *Authenticator) validateApiKey(ctx context.Context, apiKey string) (*AuthenticatedApiKey, error) {
	result := a.apiKeyCollection.FindOne(ctx, bson.M{
		"secret": apiKey,
	})

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("invalid api key")
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	apiKeyData := &apikey.ApiKey{}

	err := result.Decode(apiKeyData)

	if err != nil {
		return nil, err
	}

	return &AuthenticatedApiKey{ID: apiKeyData.ID.Hex()}, nil
}

func (a *Authenticator) ValidateAdminContext(c *gin.Context) (*AuthenticatedAdmin, error) {
	tokenType, token, err := a.extractToken(c)

	if err != nil {
		return nil, err
	}

	switch tokenType {
	case BearerToken:
		return a.validateAdminToken(token)
	default:
		return nil, fmt.Errorf("invalid token type")
	}
}

func (a *Authenticator) validateAdminToken(tokenString string) (*AuthenticatedAdmin, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSigningKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, ok := claims["type"]
		if !ok {
			return nil, fmt.Errorf("invalid token type")
		}

		tokenTypeString, ok := tokenType.(string)
		if !ok {
			return nil, fmt.Errorf("invalid token type")
		}

		if tokenTypeString != TypeAdminToken {
			return nil, fmt.Errorf("invalid token type")
		}

		id, ok := claims["sub"]

		if !ok {
			return nil, fmt.Errorf("invalid access token")
		}

		adminIDString, ok := id.(string)

		if !ok {
			return nil, fmt.Errorf("invalid access token")
		}

		return &AuthenticatedAdmin{
			ID: adminIDString,
		}, nil
	} else {
		return nil, fmt.Errorf("invalid access token")
	}
}

func (a *Authenticator) extractToken(c *gin.Context) (string, string, error) {
	authorizationHeader := c.GetHeader("Authorization")

	if len(authorizationHeader) == 0 {
		return "", "", fmt.Errorf("authorization header is not set")
	}

	components := strings.Split(authorizationHeader, " ")

	if len(components) != 2 {
		return "", "", fmt.Errorf("invalid access token")
	}

	return components[0], components[1], nil
}
