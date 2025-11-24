package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementDetails struct {
	CompetitionName  string `bson:"competitionName,omitempty" json:"competitionName,omitempty"`
	CompetitionLevel string `bson:"competitionLevel,omitempty" json:"competitionLevel,omitempty"`
	Rank             int    `bson:"rank,omitempty" json:"rank,omitempty"`
	MedalType        string `bson:"medalType,omitempty" json:"medalType,omitempty"` 

	// Publication
	PublicationType  string   `bson:"publicationType,omitempty" json:"publicationType,omitempty"`
	PublicationTitle string   `bson:"publicationTitle,omitempty" json:"publicationTitle,omitempty"`
	Authors          []string `bson:"authors,omitempty" json:"authors,omitempty"`
	Publisher        string   `bson:"publisher,omitempty" json:"publisher,omitempty"`
	ISSN             string   `bson:"issn,omitempty" json:"issn,omitempty"`

	// Organization 
	OrganizationName string    `bson:"organizationName,omitempty" json:"organizationName,omitempty"`
	Position         string    `bson:"position,omitempty" json:"position,omitempty"`
	StartDate        time.Time `bson:"startDate,omitempty" json:"startDate,omitempty"`
	EndDate          time.Time `bson:"endDate,omitempty" json:"endDate,omitempty"`

	// Certification 
	CertificationName   string    `bson:"certificationName,omitempty" json:"certificationName,omitempty"`
	IssuedBy            string    `bson:"issuedBy,omitempty" json:"issuedBy,omitempty"`
	CertificationNumber string    `bson:"certificationNumber,omitempty" json:"certificationNumber,omitempty"`
	ValidUntil          time.Time `bson:"validUntil,omitempty" json:"validUntil,omitempty"`

	// Other
	EventDate time.Time `bson:"eventDate,omitempty" json:"eventDate,omitempty"`
	Location  string    `bson:"location,omitempty" json:"location,omitempty"`
	Organizer string    `bson:"organizer,omitempty" json:"organizer,omitempty"`
	Score     float64   `bson:"score,omitempty" json:"score,omitempty"`
}

type Attachment struct {
	FileName   string    `bson:"fileName" json:"fileName"`
	FileURL    string    `bson:"fileUrl" json:"fileUrl"`
	FileType   string    `bson:"fileType" json:"fileType"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploadedAt"`
}

type Achievement struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentID       string             `bson:"studentId" json:"studentId"` 
	AchievementType string             `bson:"achievementType" json:"achievementType"` 
	Title           string             `bson:"title" json:"title"`
	Description     string             `bson:"description" json:"description"`
	Details         AchievementDetails `bson:"details" json:"details"`
	CustomFields    map[string]interface{} `bson:"customFields,omitempty" json:"customFields,omitempty"`
	Attachments     []Attachment       `bson:"attachments" json:"attachments"`
	Tags            []string           `bson:"tags" json:"tags"`
	Points          int                `bson:"points" json:"points"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}