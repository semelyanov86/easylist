package data

import "time"

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
