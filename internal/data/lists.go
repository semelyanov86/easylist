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
	CreatedAt time.Time `jsonapi:"attr,created_at,iso8601"`
	UpdatedAt time.Time `jsonapi:"attr,updated_at,iso8601"`
	IsPublic  bool      `jsonapi:"attr,is_public,omitempty"`
	Folder    *Folder   `jsonapi:"relation,folder,omitempty"`
}

type Lists []*List

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

func (l ListModel) GetAll(folderId int64, name string, userId int64, filters Filters) (Lists, Metadata, error) {
	var joinFolder string
	var fieldsFolder string
	if Contains(filters.Includes, "folder") {
		joinFolder = "INNER JOIN folders ON lists.folder_id = folders.id"
		fieldsFolder = ", folders.id, folders.user_id, folders.name, folders.icon, folders.version, folders.order, folders.created_at, folders.updated_at"
	}
	var query = fmt.Sprintf("SELECT COUNT(*) OVER(), lists.id, lists.user_id, lists.folder_id, lists.name, lists.icon, lists.version, lists.order, lists.link, lists.created_at, lists.updated_at%s FROM lists %s WHERE lists.user_id = ? AND lists.folder_id = ? AND (MATCH(lists.name) AGAINST(? IN NATURAL LANGUAGE MODE) OR ? = '') ORDER BY lists.`%s` %s, lists.`order` ASC LIMIT ? OFFSET ?", fieldsFolder, joinFolder, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var emptyMeta Metadata

	rows, err := l.DB.QueryContext(ctx, query, userId, folderId, name, name, filters.limit(), filters.offset())
	if err != nil {
		return nil, emptyMeta, err
	}

	defer rows.Close()
	var totalRecords = 0
	var lists Lists

	for rows.Next() {
		var list List
		var folder Folder
		if Contains(filters.Includes, "folder") {
			err = rows.Scan(&totalRecords, &list.ID, &list.UserId, &list.FolderId, &list.Name, &list.Icon, &list.Version, &list.Order, &list.Link, &list.CreatedAt, &list.UpdatedAt, &folder.ID, &folder.UserId, &folder.Name, &folder.Icon, &folder.Version, &folder.Order, &folder.CreatedAt, &folder.UpdatedAt)
			list.Folder = &folder
		} else {
			err = rows.Scan(&totalRecords, &list.ID, &list.UserId, &list.FolderId, &list.Name, &list.Icon, &list.Version, &list.Order, &list.Link, &list.CreatedAt, &list.UpdatedAt)
		}
		if err != nil {
			return nil, emptyMeta, err
		}

		lists = append(lists, &list)
	}

	if err = rows.Err(); err != nil {
		return nil, emptyMeta, err
	}

	var metadata = calculateMetadata(totalRecords, filters.Page, filters.Size)

	return lists, metadata, nil
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
		"self": fmt.Sprintf("%s/api/v1/lists/%d", DomainName, list.ID),
	}
}

func (lists Lists) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		jsonapi.KeyLastPage:     "",
		jsonapi.KeyFirstPage:    "",
		jsonapi.KeyPreviousPage: "",
		jsonapi.KeyNextPage:     "",
	}
}

func (lists Lists) JSONAPIMeta() *jsonapi.Meta {
	return &jsonapi.Meta{
		"total": 0,
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

func (m MockListModel) GetAll(folderId int64, name string, userId int64, filters Filters) (Lists, Metadata, error) {
	return nil, Metadata{}, nil
}
