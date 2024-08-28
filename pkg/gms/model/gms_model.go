package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateEmailRecord struct {
	ID         uuid.UUID
	EmailID    string
	ExpiryDate time.Time
	CreatedOn  time.Time
	IsDeleted  bool
}
