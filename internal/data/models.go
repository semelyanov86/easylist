package data

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users interface {
		Insert(user *User) error
		GetByEmail(email string) (*User, error)
		Update(user *User) error
		GetForToken(tokenScope, tokenPlaintext string) (*User, error)
	}
	Tokens interface {
		New(userId int64, ttl time.Duration, scope string) (*Token, error)
		Insert(token *Token) error
		DeleteAllForUser(scope string, userId int64) error
	}
	Permissions interface {
		GetAllForUser(userId int64) (Permissions, error)
		AddForUser(userId int64, codes ...string) error
	}
	Folders interface {
		Insert(folder *Folder) error
		Get(id int64, userId int64) (*Folder, error)
		Update(folder *Folder, oldOrder int32) error
		Delete(id int64, userId int64) error
	}
	Lists interface {
		Insert(list *List) error
		Get(id int64, userId int64) (*List, error)
		Update(list *List, oldOrder int32) error
		Delete(id int64, userId int64) error
	}
	Items interface {
		Insert(item *Item) error
		Get(id int64, userId int64) (*Item, error)
		Update(item *Item, oldOrder int32) error
		Delete(id int64, userId int64) error
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Folders:     FolderModel{DB: db},
		Lists:       ListModel{DB: db},
		Items:       ItemModel{DB: db},
	}
}

func NewMockModels() Models {
	return Models{
		Users:       MockUserModel{},
		Tokens:      MockTokenModel{},
		Permissions: MockPermissionModel{},
		Folders:     MockFolderModel{},
		Lists:       MockListModel{},
	}
}
