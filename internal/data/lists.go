package data

import (
	"context"
	"database/sql"
	"easylist/internal/validator"
	"errors"
	"fmt"
	"github.com/google/jsonapi"
	"strings"
	"time"
)

type List struct {
	ID        int64     `jsonapi:"primary,lists"`
	UserId    int64     `json:"-"`
	FolderId  int64     `jsonapi:"attr,folder_id"`
	Name      string    `jsonapi:"attr,name"`
	Icon      string    `jsonapi:"attr,icon"`
	Link      Link      `jsonapi:"attr,link"`
	Order     int32     `jsonapi:"attr,order"`
	Version   int32     `json:"-"`
	CreatedAt time.Time `jsonapi:"attr,created_at"`
	UpdatedAt time.Time `jsonapi:"attr,updated_at"`
	IsPublic  bool      `jsonapi:"attr,is_public,omitempty"`
}

type ListModel struct {
	DB *sql.DB
}

func (l *ListModel) GetLastListOrderForUser(userId int64) (int, error) {
	var query = "SELECT COALESCE(MAX(`order`),0) AS 'order' FROM lists WHERE lists.user_id = ?"

	var order = 0

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := l.DB.QueryRowContext(ctx, query, userId).Scan(
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

func (l ListModel) Insert(list *List) error {
	var query = "INSERT INTO lists (user_id, folder_id, name, icon, version, `order`, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())"

	lastOrder, err := l.GetLastListOrderForUser(list.UserId)
	if err != nil {
		return err
	}

	var folderId = list.FolderId
	if folderId == 0 {
		folderId = 1
	}

	var args = []any{list.UserId, folderId, list.Name, list.Icon, list.Version, lastOrder}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := l.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	list.ID = id
	list.Order = int32(lastOrder)

	return nil
}

func (l ListModel) Get(id int64, userId int64) (*List, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var query = "SELECT id, user_id, folder_id, name, icon, version, `order`, link, created_at, updated_at FROM lists WHERE id = ? AND user_id = ?"

	var list List

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var err = l.DB.QueryRowContext(ctx, query, id, userId).Scan(&list.ID, &list.UserId, &list.FolderId, &list.Name, &list.Icon, &list.Version, &list.Order, &list.Link, &list.CreatedAt, &list.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &list, nil
}

func (l ListModel) Update(list *List, oldOrder int32) error {
	var _, err = l.DB.Exec("START TRANSACTION")
	if err != nil {
		return err
	}
	var query = "UPDATE lists SET name = ?, icon = ?, folder_id = ?, link = ?, `order` = ?, version = version + 1, updated_at = NOW() WHERE id = ? AND user_id = ? AND version = ?"
	var args = []any{
		list.Name,
		list.Icon,
		list.FolderId,
		list.Link,
		list.Order,
		list.ID,
		list.UserId,
		list.Version,
	}
	list.Version++
	list.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var _, err2 = l.DB.ExecContext(ctx, query, args...)
	if err2 != nil {
		l.DB.Exec("ROLLBACK")
		switch {
		case errors.Is(err2, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err2
		}
	}

	if oldOrder != list.Order {
		var query2 = "UPDATE lists SET `order` = lists.order+1 WHERE lists.order >= ? AND user_id = ? AND id != ?"
		var _, err3 = l.DB.ExecContext(ctx, query2, list.Order, list.UserId, list.ID)
		if err3 != nil {
			l.DB.Exec("ROLLBACK")
			return err3
		}
	}

	_, err = l.DB.Exec("COMMIT")
	return err
}

func (l ListModel) Delete(id int64, userId int64) error {
	if id < 1 || userId < 1 {
		return ErrRecordNotFound
	}
	var query = "DELETE FROM lists WHERE id = ? AND user_id = ?"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := l.DB.ExecContext(ctx, query, id, userId)
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

func ValidateList(v *validator.Validator, list *List) {
	v.Check(list.Name != "", "data.attributes.name", "must be provided")
	v.Check(len(list.Name) <= 190, "data.attributes.name", "must be no more than 190 characters")
	v.Check(list.Icon == "" || strings.HasPrefix(list.Icon, "fa-"), "data.attributes.icon", "icon must starts with fa- prefix")
	v.Check(list.Order > 0, "data.attributes.order", "order should be greater then zero")
	v.Check(list.FolderId > 0, "data.attributes.folder_id", "should be greater then zero")
}

func (list List) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("/api/v1/lists/%d", list.ID),
	}
}

type MockListModel struct {
}

func (m MockListModel) Insert(list *List) error {
	return nil
}

func (m MockListModel) Get(id int64, userId int64) (*List, error) {
	return nil, nil
}

func (m MockListModel) Update(list *List, oldOrder int32) error {
	return nil
}

func (m MockListModel) Delete(id int64, userId int64) error {
	return nil
}
