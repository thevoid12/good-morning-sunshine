package model

import (
	"time"

	"github.com/google/uuid"
)

type EmailRecord struct {
	ID          uuid.UUID
	EmailID     string
	OwnerMailID string
	ExpiryDate  time.Time
	CreatedOn   time.Time
	IsDeleted   bool
}

type OwnerRecord struct {
	ID        uuid.UUID
	EmailID   string
	RateLimit int
	CreatedOn time.Time
	UpdatedOn time.Time
	IsDeleted bool
}
