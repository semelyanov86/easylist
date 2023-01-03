package data

import (
	"easylist/internal/validator"
	"strings"
	"time"
)

type Folder struct {
	ID        int64     `json:"-"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	Version   int32     `json:"version"`
	Order     int32     `json:"order"`
	UserId    int64     `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func ValidateFolder(v *validator.Validator, folder *Folder) {
	v.Check(folder.Name != "", "data.attributes.name", "must be provided")
	v.Check(len(folder.Name) <= 190, "data.attributes.name", "must be no more than 190 characters")
	v.Check(folder.Icon == "" || strings.HasPrefix(folder.Icon, "fa-"), "data.attributes.icon", "icon must starts with fa- prefix")
	v.Check(folder.Order > 0, "data.attributes.order", "order should be greater then zero")
}
