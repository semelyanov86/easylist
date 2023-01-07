package data

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type Permissions []string

func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *sql.DB
}

func (p PermissionModel) GetAllForUser(userId int64) (Permissions, error) {
	var query = `
	SELECT permissions.code from permissions 
    INNER JOIN users_permissions up on permissions.id = up.permission_id
	INNER JOIN users u on up.user_id = u.id
	WHERE u.id = ?
    `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions
	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (p PermissionModel) AddForUser(userId int64, codes ...string) error {
	var permissionMarks []string
	for _, _ = range codes {
		permissionMarks = append(permissionMarks, "?")
	}

	var query = `
INSERT INTO users_permissions (user_id, permission_id)
SELECT ?, permissions.id FROM permissions WHERE permissions.code IN (
` + strings.Join(permissionMarks, ",") + ")"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var args = []any{userId}
	for _, mark := range codes {
		args = append(args, mark)
	}

	_, err := p.DB.ExecContext(ctx, query, args...)
	return err
}

type MockPermissionModel struct {
}

func (p MockPermissionModel) GetAllForUser(userId int64) (Permissions, error) {
	return nil, nil
}

func (p MockPermissionModel) AddForUser(userId int64, codes ...string) error {
	return nil
}
