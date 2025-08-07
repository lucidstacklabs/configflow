package apikey

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiKeyCore "github.com/lucidstacklabs/configflow/internal/pkg/apikey"
	"github.com/lucidstacklabs/configflow/internal/pkg/secret"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	mongo *mongo.Collection
}

func NewService(mongo *mongo.Collection) *Service {
	return &Service{mongo: mongo}
}

func (s *Service) Create(ctx context.Context, request *CreationRequest, creatorID string) (*apiKeyCore.ApiKey, error) {
	nameExists, err := s.nameExists(ctx, request.Name)

	if err != nil {
		return nil, err
	}

	if nameExists {
		return nil, fmt.Errorf("api key %s already exists", request.Name)
	}

	apiKeySecret, err := secret.Generate(128)

	if err != nil {
		return nil, err
	}

	apiKey := &apiKeyCore.ApiKey{
		ID:        primitive.NewObjectID(),
		Name:      request.Name,
		Secret:    apiKeySecret,
		CreatorID: creatorID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = s.mongo.InsertOne(ctx, apiKey)

	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *Service) List(ctx context.Context, page int64, size int64) ([]*apiKeyCore.ApiKey, error) {
	apiKeys := make([]*apiKeyCore.ApiKey, 0)

	result, err := s.mongo.Find(ctx, bson.M{}, options.Find().SetSkip(page*size).SetLimit(size))

	if err != nil {
		return nil, err
	}

	err = result.All(ctx, &apiKeys)

	if err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (s *Service) Get(ctx context.Context, apiKeyID string) (*apiKeyCore.ApiKey, error) {
	id, err := primitive.ObjectIDFromHex(apiKeyID)

	if err != nil {
		return nil, err
	}

	result := s.mongo.FindOne(ctx, bson.M{"_id": id})

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("api key not found")
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	apiKey := &apiKeyCore.ApiKey{}

	err = result.Decode(apiKey)

	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *Service) Update(ctx context.Context, apiKeyID string, request *UpdateRequest) (*apiKeyCore.ApiKey, error) {
	id, err := primitive.ObjectIDFromHex(apiKeyID)

	if err != nil {
		return nil, err
	}

	fields := bson.M{
		"updated_at": time.Now(),
	}

	if request.Name != "" {
		count, err := s.mongo.CountDocuments(ctx, bson.M{
			"_id": bson.M{
				"$ne": id,
			},
			"name": request.Name,
		})

		if err != nil {
			return nil, err
		}

		if count > 0 {
			return nil, fmt.Errorf("api key %s already exists", request.Name)
		}

		fields["name"] = request.Name
	}

	result := s.mongo.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": fields}, options.FindOneAndUpdate().SetReturnDocument(options.After))

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("api key not found")
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	apiKey := &apiKeyCore.ApiKey{}

	err = result.Decode(apiKey)

	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *Service) Delete(ctx context.Context, apiKeyID string) (*apiKeyCore.ApiKey, error) {
	id, err := primitive.ObjectIDFromHex(apiKeyID)

	if err != nil {
		return nil, err
	}

	result := s.mongo.FindOneAndDelete(ctx, bson.M{"_id": id})

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("api key not found")
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	apiKey := &apiKeyCore.ApiKey{}

	err = result.Decode(apiKey)

	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (s *Service) GetSecret(ctx context.Context, apiKeyID string) (*SecretResponse, error) {
	apiKey, err := s.Get(ctx, apiKeyID)

	if err != nil {
		return nil, err
	}

	return &SecretResponse{Secret: apiKey.Secret}, nil
}

func (s *Service) ResetSecret(ctx context.Context, apiKeyID string) (*SecretResponse, error) {
	apiKeySecret, err := secret.Generate(128)

	if err != nil {
		return nil, err
	}

	id, err := primitive.ObjectIDFromHex(apiKeyID)

	if err != nil {
		return nil, err
	}

	fields := bson.M{
		"secret":     apiKeySecret,
		"updated_at": time.Now(),
	}

	result, err := s.mongo.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": fields})

	if err != nil {
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("api key not found")
	}

	return &SecretResponse{Secret: apiKeySecret}, nil
}

func (s *Service) Exists(ctx context.Context, apiKeyID string) (bool, error) {
	id, err := primitive.ObjectIDFromHex(apiKeyID)

	if err != nil {
		return false, err
	}

	count, err := s.mongo.CountDocuments(ctx, bson.M{"_id": id})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Service) GetByIDs(ctx context.Context, apiKeyIDs []string) ([]*apiKeyCore.ApiKey, error) {
	ids := make([]primitive.ObjectID, len(apiKeyIDs))

	for i, apiKeyID := range apiKeyIDs {
		id, err := primitive.ObjectIDFromHex(apiKeyID)

		if err != nil {
			return nil, err
		}

		ids[i] = id
	}

	result, err := s.mongo.Find(ctx, bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	})

	if err != nil {
		return nil, err
	}

	apiKeys := make([]*apiKeyCore.ApiKey, 0)

	err = result.All(ctx, &apiKeys)

	if err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (s *Service) nameExists(ctx context.Context, name string) (bool, error) {
	count, err := s.mongo.CountDocuments(ctx, bson.M{"name": name})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
