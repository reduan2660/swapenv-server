package models

import (
	"github.com/google/uuid"
	"time"
)

type Organization struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string
	CloudSync bool   `gorm:"default:false"`
	Type      string `gorm:"default:'personal'"`
	CreatedAt time.Time
}

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      *string
	Email     *string `gorm:"uniqueIndex"`
	GithubID  string  `gorm:"uniqueIndex"`
	OrgID     uuid.UUID
	Org       Organization `gorm:"foreignKey:OrgID"`
	Role      string       `gorm:"default:'admin'"`
	Can_share bool         `gorm:"default:false"`
	CreatedAt time.Time
}
