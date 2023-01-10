package data

import (
	"context"
	"database/sql"
	"easylist/internal/validator"
	"errors"
	"os"
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

	var query = "SELECT id, user_id, list_id, name, description, quantity, quantity_type, price, is_starred, file, version, `order`, created_at, updated_at FROM items WHERE id = ? AND user_id = ?"

	var item Item

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var err = i.DB.QueryRowContext(ctx, query, id, userId).Scan(&item.ID, &item.UserId, &item.ListId, &item.Name, &item.Description, &item.Quantity, &item.QuantityType, &item.Price, &item.IsStarred, &item.File, &item.Version, &item.Order, &item.CreatedAt, &item.UpdatedAt)
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
	var query = "UPDATE items SET list_id = ?, name = ?, description = ?, quantity = ?, quantity_type = ?, price = ?, is_starred = ?, file = ?, `order` = ?, version = version + 1, updated_at = NOW() WHERE id = ? AND user_id = ? AND version = ?"
	var args = []any{
		item.ListId,
		item.Name,
		item.Description,
		item.Quantity,
		item.QuantityType,
		item.Price,
		item.IsStarred,
		item.File,
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
