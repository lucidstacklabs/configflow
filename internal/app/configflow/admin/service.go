package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lucidstacklabs/configflow/internal/pkg/auth"
	"github.com/lucidstacklabs/configflow/internal/pkg/secret"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	mongo         *mongo.Collection
	authenticator *auth.Authenticator
}

func NewService(mongo *mongo.Collection, authenticator *auth.Authenticator) *Service {
	return &Service{mongo: mongo, authenticator: authenticator}
}

func (s *Service) Init(ctx context.Context, request *InitRequest) (*Admin, error) {
	count, err := s.mongo.CountDocuments(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	if count != 0 {
		return nil, fmt.Errorf("you are not allowed to perform this action")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	admin := &Admin{
		ID:        primitive.NewObjectID(),
		Username:  request.Username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = s.mongo.InsertOne(ctx, admin)

	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *Service) GetToken(ctx context.Context, request *TokenRequest) (*TokenResponse, error) {
	result := s.mongo.FindOne(ctx, bson.M{"username": request.Username})

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("invalid username and password combination")
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	var admin Admin

	if err := result.Decode(&admin); err != nil {
		return nil, err
	}

	err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(request.Password))

	if err != nil {
		return nil, fmt.Errorf("invalid username and password combination")
	}

	token, err := s.authenticator.GenerateAdminToken(admin.ID.Hex())

	if err != nil {
		return nil, err
	}

	return &TokenResponse{Token: token}, nil
}

func (s *Service) Get(ctx context.Context, adminID string) (*Admin, error) {
	id, err := primitive.ObjectIDFromHex(adminID)

	if err != nil {
		return nil, err
	}

	result := s.mongo.FindOne(ctx, bson.M{"_id": id})

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("admin not found")
	}

	admin := &Admin{}

	if err := result.Decode(&admin); err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *Service) ChangePassword(ctx context.Context, adminID string, request *PasswordChangeRequest) (*Admin, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	fields := bson.M{
		"$set": bson.M{
			"password":   string(hashedPassword),
			"updated_at": time.Now(),
		},
	}

	id, err := primitive.ObjectIDFromHex(adminID)

	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": id,
	}

	result := s.mongo.FindOneAndUpdate(ctx, filter, fields, options.FindOneAndUpdate().SetReturnDocument(options.After))

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("admin not found")
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	admin := &Admin{}

	if err := result.Decode(&admin); err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *Service) Add(ctx context.Context, request *AdditionRequest, creatorID string) (*PasswordResponse, error) {
	usernameExists, err := s.usernameExists(ctx, request.Username)

	if err != nil {
		return nil, err
	}

	if usernameExists {
		return nil, fmt.Errorf("username %s is already taken", request.Username)
	}

	password, err := secret.Generate(16)

	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	admin := &Admin{
		ID:        primitive.NewObjectID(),
		Username:  request.Username,
		Password:  string(hashedPassword),
		CreatorID: creatorID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = s.mongo.InsertOne(ctx, admin)

	if err != nil {
		return nil, err
	}

	return &PasswordResponse{Admin: admin, Password: password}, nil
}

func (s *Service) Delete(ctx context.Context, adminID string) (*Admin, error) {
	id, err := primitive.ObjectIDFromHex(adminID)

	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": id}

	result := s.mongo.FindOneAndDelete(ctx, filter)

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("admin not found")
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	admin := &Admin{}

	if err := result.Decode(&admin); err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *Service) List(ctx context.Context, page int64, size int64) ([]*Admin, error) {
	result, err := s.mongo.Find(ctx, bson.M{}, options.Find().SetSkip(page*size).SetLimit(size))

	if err != nil {
		return nil, err
	}

	admins := make([]*Admin, 0)

	err = result.All(ctx, &admins)

	if err != nil {
		return nil, err
	}

	return admins, nil
}

func (s *Service) ResetPassword(ctx context.Context, adminID string) (*PasswordResponse, error) {
	password, err := secret.Generate(16)

	if err != nil {
		return nil, err
	}

	admin, err := s.ChangePassword(ctx, adminID, &PasswordChangeRequest{Password: password})

	if err != nil {
		return nil, err
	}

	return &PasswordResponse{Admin: admin, Password: password}, nil
}

func (s *Service) usernameExists(ctx context.Context, username string) (bool, error) {
	count, err := s.mongo.CountDocuments(ctx, bson.M{"username": username})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
