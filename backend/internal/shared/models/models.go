package models

import (
	"encoding/json"
	"time"
)

type LoginDTO struct {
	Email    string `json:"Email" validate:"required,email"`
	Password string `json:"Password" validate:"required,min=8,max=64"`
}

type SignupDTO struct {
	Email    string `json:"Email" validate:"required,email"`
	Password string `json:"Password" validate:"required,min=8,max=64"`
}

type Track struct {
	ExternalID string      `json:"SpotifyID"`
	Name       string      `json:"Name"`
	ImageURL   string      `json:"ImageURL"`
	ArtistRefs []ArtistRef `json:"ArtistRefs"`
}

// stores data fetched from foreigh provider
type ArtistRef struct {
	ExternalID string `json:"SpotifyID"`
	Name       string `json:"Name"`
}

type Artist struct {
	ID            int    `json:"ID"`
	Name          string `json:"Name"`
	SpotifyID     string `json:"SpotifyID"`
	DescriptionUA string `json:"DescriptionUA"`
	DescriptionEN string `json:"DescriptionEN"`
	Source        string `json:"Source"`
	SourceURL     string `json:"SourceURL"`
	Country       string `json:"Country"`
	Confirmed     bool   `json:"Confirmed"`
}

type Playlist struct {
	SpotifyID string  `json:"SpotifyID"`
	Tracks    []Track `json:"Tracks"`
}

type Album struct {
	SpotifyID string  `json:"SpotifyID"`
	Tracks    []Track `json:"Tracks"`
}

// carries deduplicated tracks and artistRefs from all tracks
type Content struct {
	Tracks  []Track
	Artists []ArtistRef
}

type RuContent struct {
	Tracks  []Track  `json:"Tracks"`
	Artists []Artist `json:"Artists"`
}

type User struct {
	ID           int    `json:"ID"`
	Email        string `json:"Email"`
	Role         Role   `json:"Role"`
	PasswordHash string `json:"-"`
}

type ArtistInsertSuggestion struct {
	ID            int    `json:"ID"`
	CreatorID     int    `json:"CreatorID"`
	ArtistName    string `json:"ArtistName"`
	Description   string `json:"Description"`
	State         string `json:"State"`
	DeclineReason string `json:"DeclineReason,omitempty"`
	CreatedAt     string `json:"CreatedAt"`
	UpdatedAt     string `json:"UpdatedAt"`
}

type ArtistDeleteSuggestion struct {
	ID            int    `json:"ID"`
	CreatorID     int    `json:"CreatorID"`
	ArtistName    string `json:"ArtistName"`
	Description   string `json:"Description"`
	State         string `json:"State"`
	DeclineReason string `json:"DeclineReason,omitempty"`
	CreatedAt     string `json:"CreatedAt"`
	UpdatedAt     string `json:"UpdatedAt"`
}

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDone       JobStatus = "done"
	JobStatusFailed     JobStatus = "failed"
)

type Job struct {
	JobID     string          `json:"jobId"`
	Status    JobStatus       `json:"status"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}
