package models

import (
	"time"

	pgvector "github.com/pgvector/pgvector-go"
)

type AudioFeature struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	Filename      string          `gorm:"size:255;not null" json:"filename"`
	Instrument    string          `gorm:"size:100;not null" json:"instrument"`
	Genre         string          `gorm:"size:100" json:"genre"`
	Technique     string          `gorm:"size:100;default:standard" json:"technique"`
	DurationSec   float64         `gorm:"default:3.0" json:"duration_sec"`
	SampleRate    int             `gorm:"default:22050" json:"sample_rate"`
	FeatureVector pgvector.Vector `gorm:"type:vector(19);not null" json:"feature_vector"`
	CreatedAt     time.Time       `json:"created_at"`
}
