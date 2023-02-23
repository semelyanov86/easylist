package data

import (
	"context"
	"database/sql"
	"easylist/internal/validator"
	"errors"
	"fmt"
	"github.com/google/jsonapi"
	"os"
	"time"
)

type Item struct {
	ID           int64     `jsonapi:"primary,items"`
	UserId       int64     `json:"-"`
	ListId       int64     `jsonapi:"attr,list_id"`
	Name         string    `jsonapi:"attr,name"`
	Description  string    `jsonapi:"attr,description"`
	Quantity     int32     `jsonapi:"attr,quantity"`
	QuantityType string    `jsonapi:"attr,quantity_type"`
	Price        float32   `jsonapi:"attr,price"`
	IsStarred    bool      `jsonapi:"attr,is_starred"`
	IsDone       bool      `jsonapi:"attr,is_done"`
	File         string    `jsonapi:"attr,file"`
	Order        int32     `jsonapi:"attr,order"`
	Version      int32     `json:"-"`
	CreatedAt    time.Time `jsonapi:"attr,created_at,iso8601" json:"created_at" time_format:"sql_datetime"`
	UpdatedAt    time.Time `jsonapi:"attr,updated_at,iso8601" json:"updated_at" time_format:"sql_datetime"`
	List         *List     `jsonapi:"relation,list,omitempty"`
}

const ItemsType = "items"

type Items []*Item

type ItemModel struct {
	DB *sql.DB
}

func (i ItemModel) GetLastItemOrderForUser(userId int64, listId int64) (int, error) {
	var query = "SELECT COALESCE(MAX(`order`),0) AS 'order' FROM items WHERE items.user_id = ? AND items.list_id = ?"

	var order = 0

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := i.DB.QueryRowContext(ctx, query, userId, listId).Scan(
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

func (i ItemModel) Insert(item *Item) error {
	var query = "INSERT INTO items (user_id, list_id, name, description, quantity, quantity_type, price, is_starred, file, version, `order`, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, NOW(), NOW())"

	lastOrder, err := i.GetLastItemOrderForUser(item.UserId, item.ListId)
	if err != nil {
		return err
	}

	var args = []any{item.UserId, item.ListId, item.Name, item.Description, item.Quantity, item.QuantityType, item.Price, item.IsStarred, item.File, lastOrder}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := i.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	item.ID = id
	item.Order = int32(lastOrder)
	return nil
}

func (i ItemModel) Get(id int64, userId int64) (*Item, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var query = "SELECT id, user_id, list_id, name, description, quantity, quantity_type, price, is_starred, file, version, `order`, is_done, created_at, updated_at FROM items WHERE id = ? AND user_id = ?"

	var item Item

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var err = i.DB.QueryRowContext(ctx, query, id, userId).Scan(&item.ID, &item.UserId, &item.ListId, &item.Name, &item.Description, &item.Quantity, &item.QuantityType, &item.Price, &item.IsStarred, &item.File, &item.Version, &item.Order, &item.IsDone, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &item, nil
}

func (i ItemModel) Update(item *Item, oldOrder int32) error {
	var _, err = i.DB.Exec("START TRANSACTION")
	if err != nil {
		return err
	}
	var query = "UPDATE items SET list_id = ?, name = ?, description = ?, quantity = ?, quantity_type = ?, price = ?, is_starred = ?, file = ?, is_done = ?, `order` = ?, version = version + 1, updated_at = NOW() WHERE id = ? AND user_id = ? AND version = ?"
	var args = []any{
		item.ListId,
		item.Name,
		item.Description,
		item.Quantity,
		item.QuantityType,
		item.Price,
		item.IsStarred,
		item.File,
		item.IsDone,
		item.Order,
		item.ID,
		item.UserId,
		item.Version,
	}
	item.Version++
	item.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var _, err2 = i.DB.ExecContext(ctx, query, args...)
	if err2 != nil {
		i.DB.Exec("ROLLBACK")
		switch {
		case errors.Is(err2, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err2
		}
	}

	if oldOrder != item.Order {
		var query2 = "UPDATE items SET `order` = items.order+1 WHERE items.order >= ? AND user_id = ? AND id != ?"
		var _, err3 = i.DB.ExecContext(ctx, query2, item.Order, item.UserId, item.ID)
		if err3 != nil {
			i.DB.Exec("ROLLBACK")
			return err3
		}
	}

	_, err = i.DB.Exec("COMMIT")
	return err
}

func (i ItemModel) Delete(id int64, userId int64) error {
	if id < 1 || userId < 1 {
		return ErrRecordNotFound
	}
	item, err := i.Get(id, userId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	var query = "DELETE FROM items WHERE id = ? AND user_id = ?"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := i.DB.ExecContext(ctx, query, id, userId)
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
	if len(item.File) > 0 {
		e := os.Remove(item.File)
		if e != nil {
			return e
		}
	}
	return nil
}

func (i ItemModel) DeleteByUser(userId int64) error {
	if userId < 1 {
		return ErrRecordNotFound
	}
	var query = "DELETE FROM items WHERE user_id = ?"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := i.DB.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil
}

func (i ItemModel) GetAll(name string, userId int64, listId int64, isStarred bool, filters Filters) (Items, Metadata, error) {
	var joinList string
	var fieldsList string
	var starredFilter = ""
	if isStarred {
		starredFilter = "AND items.is_starred = 1"
	}
	if Contains(filters.Includes, "list") {
		joinList = "INNER JOIN lists ON items.list_id = lists.id"
		fieldsList = ", lists.id, lists.folder_id, lists.user_id, lists.name, lists.icon, lists.version, lists.order, lists.link, lists.created_at, lists.updated_at"
	}
	var query = fmt.Sprintf("SELECT COUNT(*) OVER(), items.id, items.user_id, items.list_id, items.name, items.description, items.quantity, items.quantity_type, items.price, items.is_starred, items.file, items.version, items.order, items.is_done, items.created_at, items.updated_at%s FROM items %s WHERE items.user_id = ? AND (items.list_id = ? OR ? = 0) %s AND (MATCH(items.name) AGAINST(? IN NATURAL LANGUAGE MODE) OR ? = '') ORDER BY items.%s %s, items.order ASC LIMIT ? OFFSET ?", fieldsList, joinList, starredFilter, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var emptyMeta Metadata

	rows, err := i.DB.QueryContext(ctx, query, userId, listId, listId, name, name, filters.limit(), filters.offset())
	if err != nil {
		return nil, emptyMeta, err
	}
	defer rows.Close()
	var totalRecords = 0
	var items Items

	for rows.Next() {
		var list List
		var item Item
		if Contains(filters.Includes, "list") {
			err = rows.Scan(&totalRecords, &item.ID, &item.UserId, &item.ListId, &item.Name, &item.Description, &item.Quantity, &item.QuantityType, &item.Price, &item.IsStarred, &item.File, &item.Version, &item.Order, &item.IsDone, &item.CreatedAt, &item.UpdatedAt, &list.ID, &list.UserId, &list.FolderId, &list.Name, &list.Icon, &list.Version, &list.Order, &list.Link, &list.CreatedAt, &list.UpdatedAt)
			item.List = &list
		} else {
			err = rows.Scan(&totalRecords, &item.ID, &item.UserId, &item.ListId, &item.Name, &item.Description, &item.Quantity, &item.QuantityType, &item.Price, &item.IsStarred, &item.File, &item.Version, &item.Order, &item.IsDone, &item.CreatedAt, &item.UpdatedAt)
		}
		if err != nil {
			return nil, emptyMeta, err
		}

		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, emptyMeta, err
	}

	var metadata = calculateMetadata(totalRecords, filters.Page, filters.Size, listId, "lists")

	return items, metadata, nil
}

func (item Item) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("%s/api/v1/items/%d", DomainName, item.ID),
	}
}

func (items Items) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		jsonapi.KeyLastPage:     "",
		jsonapi.KeyFirstPage:    "",
		jsonapi.KeyNextPage:     "",
		jsonapi.KeyPreviousPage: "",
	}
}

func (items Items) JSONAPIMeta() *jsonapi.Meta {
	return &jsonapi.Meta{
		"total": 0,
	}
}

func ValidateItem(v *validator.Validator, item *Item) {
	v.Check(item.Name != "", "data.attributes.name", "must be provided")
	v.Check(len(item.Name) <= 190, "data.attributes.name", "must be no more than 190 characters")
	v.Check(len(item.Description) <= 500, "data.attributes.description", "must be no more than 500 characters")
	v.Check(item.Order > 0, "data.attributes.order", "order should be greater then zero")
	v.Check(item.Quantity >= 0, "data.attributes.quantity", "should be greater then zero")
	v.Check(item.ListId > 0, "data.attributes.list_id", "should be greater then zero")
}

type MockItemModel struct {
}

func (i MockItemModel) GetLastItemOrderForUser(userId int64, listId int64) (int, error) {
	return 1, nil
}

func (i MockItemModel) Insert(item *Item) error {
	return nil
}

func (i MockItemModel) Get(id int64, userId int64) (*Item, error) {
	return nil, nil
}

func (i MockItemModel) Update(item *Item, oldOrder int32) error {
	return nil
}

func (i MockItemModel) Delete(id int64, userId int64) error {
	return nil
}

func (i MockItemModel) GetAll(name string, userId int64, listId int64, isStarred bool, filters Filters) (Items, Metadata, error) {
	return Items{}, Metadata{}, nil
}

func (i MockItemModel) DeleteByUser(userId int64) error {
	return nil
}
