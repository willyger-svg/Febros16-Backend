package models

import (
	"time"
)

// User inawakilisha jedwali la watumiaji (sasa ina vitu vyote vya wasifu)
type User struct {
	ID                string    `json:"id"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"`
	Role              string    `json:"role"`
	Bio               string    `json:"bio"`
	ProfilePictureURL string    `json:"profile_picture_url"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	HasCompletedAssessment *bool `json:"has_completed_assessment"`
	AssessmentData   *string    `json:"assessment_data"`
	GoogleID         *string    `json:"google_id"`
}
