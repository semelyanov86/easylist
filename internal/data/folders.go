package data

import (
	"context"
	"database/sql"
	"easylist/internal/validator"
	"errors"
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

type FolderModel struct {
	DB *sql.DB
}

func (f FolderModel) GetLastFolderOrderForUser(userId int64) (int, error) {
	var query = "SELECT COALESCE(MAX(`order`),1) AS 'order' FROM folders WHERE folders.user_id = ?"

	var order = 0

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := f.DB.QueryRowContext(ctx, query, userId).Scan(
		&order,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 1, nil
		default:
			return 1, err
		}
	}
	return order + 1, nil
}

func (f FolderModel) Insert(folder *Folder) error {
	var query = "INSERT INTO folders (user_id, name, icon, version, `order`, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW())"

	lastOrder, err := f.GetLastFolderOrderForUser(folder.UserId)
	if err != nil {
		return err
	}

	var args = []any{folder.UserId, folder.Name, folder.Icon, folder.Version, lastOrder}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := f.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	folder.ID = id
	folder.Order = int32(lastOrder)
	return nil
}

func (f FolderModel) Get(id int64) (*Folder, error) {
	return nil, nil
}

func (f FolderModel) Update(folder Folder) error {
	return nil
}

func (f FolderModel) Delete(id int64) error {
	return nil
}

func ValidateFolder(v *validator.Validator, folder *Folder) {
	v.Check(folder.Name != "", "data.attributes.name", "must be provided")
	v.Check(len(folder.Name) <= 190, "data.attributes.name", "must be no more than 190 characters")
	v.Check(folder.Icon == "" || strings.HasPrefix(folder.Icon, "fa-"), "data.attributes.icon", "icon must starts with fa- prefix")
}

type MockFolderModel struct {
}

func (m MockFolderModel) Insert(folder *Folder) error {
	return nil
}

func (m MockFolderModel) Get(id int64) (*Folder, error) {
	return nil, nil
}

func (m MockFolderModel) Update(folder Folder) error {
	return nil
}

func (m MockFolderModel) Delete(id int64) error {
	return nil
}
