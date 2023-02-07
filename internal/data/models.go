package data

import (
	"database/sql"
	"errors"
	"github.com/liamylian/jsontime"
	"strings"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	DomainName        string
)

var json = jsontime.ConfigWithCustomTimeFormat

type ComplexModels interface {
	Folders | Items | Lists
}

type ComplexModel interface {
	*Folder | *Item | *List
}

type Models struct {
	Users interface {
		Insert(user *User) error
		GetByEmail(email string) (*User, error)
		Update(user *User) error
		GetForToken(tokenScope, tokenPlaintext string) (*User, error)
		Delete(id int64) error
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
		GetAll(name string, userId int64, filters Filters) (Folders, Metadata, error)
		DeleteByUser(userId int64) error
	}
	Lists interface {
		Insert(list *List) error
		Get(id int64, userId int64) (*List, error)
		GetAll(folderId int64, name string, userId int64, filters Filters) (Lists, Metadata, error)
		Update(list *List, oldOrder int32) error
		Delete(id int64, userId int64) error
	}
	Items interface {
		Insert(item *Item) error
		Get(id int64, userId int64) (*Item, error)
		Update(item *Item, oldOrder int32) error
		Delete(id int64, userId int64) error
		GetAll(name string, userId int64, listId int64, isStarred bool, filters Filters) (Items, Metadata, error)
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
		Items:       MockItemModel{},
	}
}

func ConvertSliceToQuestionMarks(ids []any) string {
	var result string
	for range ids {
		result = result + "?,"
	}
	return strings.Trim(result, ",")
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
