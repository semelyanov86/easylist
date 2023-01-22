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

const FolderType = "folders"

type Folder struct {
	ID        int64         `jsonapi:"primary,folders"`
	Name      string        `jsonapi:"attr,name"`
	Icon      string        `jsonapi:"attr,icon"`
	Version   int32         `json:"-"`
	Order     int32         `jsonapi:"attr,order"`
	UserId    sql.NullInt64 `json:"-"`
	CreatedAt time.Time     `jsonapi:"attr,created_at,iso8601"`
	UpdatedAt time.Time     `jsonapi:"attr,updated_at,iso8601"`
	Lists     Lists         `jsonapi:"relation,lists,omitempty"`
}

type Folders []*Folder

type FolderModel struct {
	DB *sql.DB
}

func (f FolderModel) GetLastFolderOrderForUser(userId int64) (int, error) {
	var query = "SELECT COALESCE(MAX(`order`),0) AS 'order' FROM folders WHERE folders.user_id = ?"

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

func (f FolderModel) GetByIds(ids []any) (Folders, error) {
	var folders Folders
	if len(ids) < 1 {
		return folders, nil
	}
	var query = "SELECT id, user_id, name, icon, version, `order`, created_at, updated_at FROM folders WHERE folders.id IN (" + ConvertSliceToQuestionMarks(ids) + ")"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := f.DB.QueryContext(ctx, query, ids...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var totalRecords = 0

	for rows.Next() {
		var folder Folder

		err := rows.Scan(&totalRecords, &folder.ID, &folder.UserId, &folder.Name, &folder.Icon, &folder.Version, &folder.Order, &folder.CreatedAt, &folder.UpdatedAt)
		if err != nil {
			return nil, err
		}

		folders = append(folders, &folder)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return folders, nil
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

func (f FolderModel) GetAll(name string, userId int64, filters Filters) (Folders, Metadata, error) {
	var joinList string
	var fieldsList string
	var groupList string
	if Contains(filters.Includes, "lists") {
		joinList = "LEFT JOIN lists ON lists.folder_id = folders.id"
		fieldsList = ", (SELECT CONCAT('[',GROUP_CONCAT(JSON_OBJECT('id', lists.id, 'user_id', lists.user_id, 'FolderId', lists.folder_id, 'name', lists.name, 'icon', lists.icon, 'version', lists.version, 'order', lists.order, 'link', lists.link, 'created_at', lists.created_at, 'updated_at', lists.updated_at)),']')) as parsed_lists"
		groupList = "GROUP BY folders.id"
	}
	var query = fmt.Sprintf("SELECT COUNT(*) OVER(), folders.id, folders.user_id, folders.name, folders.icon, folders.version, folders.order, folders.created_at, folders.updated_at%s FROM folders %s WHERE (folders.user_id = ? OR folders.user_id IS NULL) AND (MATCH(folders.name) AGAINST(? IN NATURAL LANGUAGE MODE) OR ? = '') %s ORDER BY folders.%s %s, folders.order ASC LIMIT ? OFFSET ?", fieldsList, joinList, groupList, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var emptyMeta Metadata

	rows, err := f.DB.QueryContext(ctx, query, userId, name, name, filters.limit(), filters.offset())
	if err != nil {
		return nil, emptyMeta, err
	}

	defer rows.Close()

	var totalRecords = 0

	var folders Folders
	var currentFolder Folder

	for rows.Next() {
		var folder Folder
		var folderId, folderUserId sql.NullInt64
		var folderName, folderIcon, parsedList sql.NullString
		var folderVersion, folderOrder sql.NullInt32
		var folderCreatedAt, folderUpdatedAt sql.NullTime

		if Contains(filters.Includes, "lists") {
			err = rows.Scan(&totalRecords, &folderId, &folderUserId, &folderName, &folderIcon, &folderVersion, &folderOrder, &folderCreatedAt, &folderUpdatedAt, &parsedList)
			if err != nil {
				return nil, emptyMeta, err
			}
			if currentFolder.ID != folderId.Int64 {
				currentFolder = Folder{
					ID:        folderId.Int64,
					Name:      folderName.String,
					Icon:      folderIcon.String,
					Version:   folderVersion.Int32,
					Order:     folderOrder.Int32,
					UserId:    folderUserId,
					CreatedAt: folderCreatedAt.Time,
					UpdatedAt: folderUpdatedAt.Time,
				}
			}
			var tempTags []List

			if parsedList.Valid {
				if err := json.Unmarshal([]byte(parsedList.String), &tempTags); err != nil {
					return nil, emptyMeta, err
				}
				for _, curList := range tempTags {
					if curList.ID > 0 {
						var importList = curList
						currentFolder.Lists = append(currentFolder.Lists, &importList)
					}
				}
			}

			folder = currentFolder
			folders = append(folders, &folder)
		} else {
			err = rows.Scan(
				&totalRecords,
				&folder.ID,
				&folder.UserId,
				&folder.Name,
				&folder.Icon,
				&folder.Version,
				&folder.Order,
				&folder.CreatedAt,
				&folder.UpdatedAt,
			)
			if err != nil {
				return nil, emptyMeta, err
			}
			folders = append(folders, &folder)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, emptyMeta, err
	}

	var metadata = calculateMetadata(totalRecords, filters.Page, filters.Size)

	return folders, metadata, nil
}

func ValidateFolder(v *validator.Validator, folder *Folder) {
	v.Check(folder.Name != "", "data.attributes.name", "must be provided")
	v.Check(len(folder.Name) <= 190, "data.attributes.name", "must be no more than 190 characters")
	v.Check(folder.Icon == "" || strings.HasPrefix(folder.Icon, "fa-"), "data.attributes.icon", "icon must starts with fa- prefix")
}

func (folder Folder) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		"self": fmt.Sprintf("%s/api/v1/folders/%d", DomainName, folder.ID),
	}
}

func (folders Folders) JSONAPILinks() *jsonapi.Links {
	return &jsonapi.Links{
		jsonapi.KeyLastPage: "",
	}
}

func (folders Folders) JSONAPIMeta() *jsonapi.Meta {
	return &jsonapi.Meta{
		"total": 0,
	}
}

func (folder Folder) JSONAPIRelationshipLinks(relation string) *jsonapi.Links {
	if relation == "lists" {
		return &jsonapi.Links{
			"related": fmt.Sprintf("%s/folders/%d/lists", DomainName, folder.ID),
		}
	}
	return nil
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

func (m MockFolderModel) GetAll(name string, userId int64, filters Filters) (Folders, Metadata, error) {
	return nil, Metadata{}, nil
}
