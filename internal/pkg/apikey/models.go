package apikey

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApiKey struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Secret    string             `bson:"secret" json:"-"`
	CreatorID string             `bson:"creator_id" json:"creator_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
