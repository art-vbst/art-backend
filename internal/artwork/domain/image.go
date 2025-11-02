package domain

import (
	"time"

	"github.com/google/uuid"
)

type Image struct {
	ID          uuid.UUID `json:"id"`
	ArtworkID   uuid.UUID `json:"artwork_id"`
	IsMainImage bool      `json:"is_main_image"`
	ImageURL    string    `json:"image_url"`
	ObjectName  string    `json:"object_name"`
	ImageWidth  *int32    `json:"image_width"`
	ImageHeight *int32    `json:"image_height"`
	CreatedAt   time.Time `json:"created_at"`
}
