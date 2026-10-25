package entity

import "github.com/sbondCo/Watcharr/database/dbmodel"

type PushSubscription struct {
	dbmodel.GormModel
	UserID    uint   `json:"userId" gorm:"not null;index"`
	Endpoint  string `json:"endpoint" gorm:"not null;uniqueIndex"`
	P256dh    string `json:"p256dh" gorm:"not null"`
	Auth      string `json:"auth" gorm:"not null"`
	UserAgent string `json:"userAgent,omitempty"`
}

