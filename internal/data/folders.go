package data

import "time"

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
