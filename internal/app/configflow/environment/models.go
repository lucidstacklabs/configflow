package environment

import (
	"time"

	"github.com/lucidstacklabs/configflow/internal/pkg/actor"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Environment struct {
	ID          primitive.ObjectID `bson:"_id," json:"id"`
	Name        string             `bson:"name" json:"name"`
	CreatorType actor.Type         `bson:"creator_type" json:"creator_type"`
	CreatorID   string             `bson:"creator_id" json:"creator_id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
