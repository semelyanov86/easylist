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
	ID        int64         `json:"-"`
	Name      string        `json:"name"`
	Icon      string        `json:"icon"`
	Version   int32         `json:"version"`
	Order     int32         `json:"order"`
	UserId    sql.NullInt64 `json:"-"`
	CreatedAt time.Time     `json:"-"`
	UpdatedAt time.Time     `json:"-"`
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

	lastOrder, err := f.GetLastFolderOrderForUser(folder.UserId.Int64)
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

func (f FolderModel) Get(id int64, userId int64) (*Folder, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var query = "SELECT id, user_id, name, icon, version, `order`, created_at, updated_at FROM folders WHERE id = ? AND (user_id = ? OR user_id IS NULL)"

	var folder Folder

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var err = f.DB.QueryRowContext(ctx, query, id, userId).Scan(&folder.ID, &folder.UserId, &folder.Name, &folder.Icon, &folder.Version, &folder.Order, &folder.CreatedAt, &folder.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &folder, nil
}

func (f FolderModel) Update(folder *Folder, oldOrder int32) error {
	var _, err = f.DB.Exec("START TRANSACTION")
	if err != nil {
		return err
	}
	var query = "UPDATE folders SET name = ?, icon = ?, `order` = ?, version = version + 1, updated_at = NOW() WHERE id = ? AND user_id = ? AND version = ?"
	var args = []any{
		folder.Name,
		folder.Icon,
		folder.Order,
		folder.ID,
		folder.UserId,
		folder.Version,
	}
	folder.Version++
	folder.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var _, err2 = f.DB.ExecContext(ctx, query, args...)
	if err2 != nil {
		f.DB.Exec("ROLLBACK")
		switch {
		case errors.Is(err2, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err2
		}
	}

	if oldOrder != folder.Order {
		var query2 = "UPDATE folders SET `order` = folders.order+1 WHERE folders.order >= ? AND user_id = ? AND id != ?"
		var _, err3 = f.DB.ExecContext(ctx, query2, folder.Order, folder.UserId, folder.ID)
		if err3 != nil {
			f.DB.Exec("ROLLBACK")
			return err3
		}
	}

	_, err = f.DB.Exec("COMMIT")
	return err
}

func (f FolderModel) Delete(id int64, userId int64) error {
	if id < 1 || userId < 1 {
		return ErrRecordNotFound
	}
	var query = "DELETE FROM folders WHERE id = ? AND user_id = ?"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := f.DB.ExecContext(ctx, query, id, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
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

func (m MockFolderModel) Get(id int64, userId int64) (*Folder, error) {
	return nil, nil
}

func (m MockFolderModel) Update(folder *Folder, oldOrder int32) error {
	return nil
}

func (m MockFolderModel) Delete(id int64, userId int64) error {
	return nil
}
