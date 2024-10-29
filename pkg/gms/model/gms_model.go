package model

import (
	"time"

	"github.com/google/uuid"
)

type EmailRecord struct {
	ID            uuid.UUID
	EmailID       string
	OwnerMailID   string
	ExpiryDate    time.Time
	RandomNumbers string // the random number days which are used to send good morning msgs. this is stored to make sure that rand doesnt generate the same number again and again
	TimeZone      string
	CreatedOn     time.Time
	IsDeleted     bool
}

type OwnerRecord struct {
	ID        uuid.UUID
	EmailID   string
	RateLimit int
	CreatedOn time.Time
	UpdatedOn time.Time
	IsDeleted bool
}
