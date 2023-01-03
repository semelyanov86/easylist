package data

import (
	"easylist/internal/validator"
	"time"
)

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
	Version      int32     `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func ValidateItem(v *validator.Validator, item *Item) {
	v.Check(item.Name != "", "data.attributes.name", "must be provided")
	v.Check(len(item.Name) <= 190, "data.attributes.name", "must be no more than 190 characters")
	v.Check(len(item.Description) <= 500, "data.attributes.description", "must be no more than 500 characters")
	v.Check(item.Order > 0, "data.attributes.order", "order should be greater then zero")
	v.Check(item.Quantity > 0, "data.attributes.quantity", "should be greater then zero")
	v.Check(item.ListId > 0, "data.attributes.list_id", "should be greater then zero")
}
