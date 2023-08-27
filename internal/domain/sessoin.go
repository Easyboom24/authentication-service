package domain

import "time"

type Session struct {
	FingerPrint string `json:"fingerPrint" bson:"fingerprint"`
	RefreshToken string `json:"refreshToken" bson:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt" bson:"expiresAt"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}