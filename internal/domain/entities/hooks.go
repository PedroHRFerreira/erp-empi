package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Timestamps struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func assignID(id *string) {
	if *id == "" {
		*id = uuid.NewString()
	}
}

func notFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
