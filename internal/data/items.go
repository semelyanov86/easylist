package data

import "time"

type Item struct {
	ID           int64     `json:"-"`
	UserId       int64     `json:"-"`
	ListId       int64     `json:"list_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Quantity     int32     `json:"quantity"`
	QuantityType string    `json:"quantity_type"`
	Price        float32   `json:"price"`
	IsStarred    bool      `json:"is_starred"`
	File         string    `json:"file"`
	Order        int32     `json:"order"`
	Version      int32     `json:"version"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}
