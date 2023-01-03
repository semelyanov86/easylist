package data

import (
	"easylist/internal/validator"
	"strings"
	"time"
)

type List struct {
	ID        int64     `json:"-"`
	UserId    int64     `json:"-"`
	FolderId  int64     `json:"folder_id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	Link      string    `json:"link"`
	Order     int32     `json:"order"`
	Version   int32     `json:"version"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func ValidateList(v *validator.Validator, list *List) {
	v.Check(list.Name != "", "data.attributes.name", "must be provided")
	v.Check(len(list.Name) <= 190, "data.attributes.name", "must be no more than 190 characters")
	v.Check(list.Icon == "" || strings.HasPrefix(list.Icon, "fa-"), "data.attributes.icon", "icon must starts with fa- prefix")
	v.Check(list.Order > 0, "data.attributes.order", "order should be greater then zero")
	v.Check(list.FolderId > 0, "data.attributes.folder_id", "should be greater then zero")
}
