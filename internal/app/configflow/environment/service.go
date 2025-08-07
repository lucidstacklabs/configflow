package environment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lucidstacklabs/configflow/internal/pkg/actor"
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

func (s *Service) Create(ctx context.Context, request *CreateRequest, creatorType actor.Type, creatorID string) (*Environment, error) {
	nameExists, err := s.nameExists(ctx, request.Name)

	if err != nil {
		return nil, err
	}

	if nameExists {
		return nil, fmt.Errorf("environment %s already exists", request.Name)
	}

	environment := &Environment{
		ID:          primitive.NewObjectID(),
		Name:        request.Name,
		CreatorType: creatorType,
		CreatorID:   creatorID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = s.mongo.InsertOne(ctx, environment)

	if err != nil {
		return nil, err
	}

	return environment, nil
}

func (s *Service) List(ctx context.Context, page int64, size int64) ([]*Environment, error) {
	cur, err := s.mongo.Find(ctx, bson.M{}, options.Find().SetSkip(page*size).SetLimit(size))

	if err != nil {
		return nil, err
	}

	environments := make([]*Environment, 0)

	err = cur.All(ctx, &environments)

	if err != nil {
		return nil, err
	}

	return environments, nil
}

func (s *Service) Get(ctx context.Context, environmentID string) (*Environment, error) {
	id, err := primitive.ObjectIDFromHex(environmentID)

	if err != nil {
		return nil, err
	}

	env := &Environment{}
	err = s.mongo.FindOne(ctx, bson.M{"_id": id}).Decode(env)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("environment not found")
	}

	if err != nil {
		return nil, err
	}

	return env, nil
}

func (s *Service) Update(ctx context.Context, environmentID string, request *UpdateRequest) (*Environment, error) {
	id, err := primitive.ObjectIDFromHex(environmentID)

	if err != nil {
		return nil, err
	}

	fields := bson.M{
		"updated_at": time.Now(),
	}

	if request.Name != "" {
		count, err := s.mongo.CountDocuments(ctx, bson.M{"_id": bson.M{"$ne": id}, "name": request.Name})

		if err != nil {
			return nil, err
		}

		if count > 0 {
			return nil, fmt.Errorf("environment %s already exists", request.Name)
		}

		fields["name"] = request.Name
	}

	environment := &Environment{}
	err = s.mongo.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": fields}).Decode(environment)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("environment not found")
	}

	if err != nil {
		return nil, err
	}

	return environment, nil
}

func (s *Service) Delete(ctx context.Context, environmentID string) (*Environment, error) {
	id, err := primitive.ObjectIDFromHex(environmentID)

	if err != nil {
		return nil, err
	}

	environment := &Environment{}

	err = s.mongo.FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(environment)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("environment not found")
	}

	if err != nil {
		return nil, err
	}

	return environment, nil
}

func (s *Service) nameExists(ctx context.Context, name string) (bool, error) {
	count, err := s.mongo.CountDocuments(ctx, bson.M{"name": name})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
