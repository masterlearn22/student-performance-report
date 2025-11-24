package models

import (
	"time"
	"github.com/google/uuid"
)

const (
	StatusDraft     = "draft"
	StatusSubmitted = "submitted"
	StatusVerified  = "verified"
	StatusRejected  = "rejected"
)

type AchievementReference struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	StudentID          uuid.UUID  `json:"studentId" db:"student_id"`
	MongoAchievementID string     `json:"mongoAchievementId" db:"mongo_achievement_id"`
	Status             string     `json:"status" db:"status"` 
	SubmittedAt        *time.Time `json:"submittedAt" db:"submitted_at"`
	VerifiedAt         *time.Time `json:"verifiedAt" db:"verified_at"`
	VerifiedBy         *uuid.UUID `json:"verifiedBy" db:"verified_by"`
	RejectionNote      *string    `json:"rejectionNote" db:"rejection_note"`
	CreatedAt          time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" db:"updated_at"`
}