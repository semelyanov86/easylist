package data

import (
	"database/sql"
	"errors"
)

var ErrInvalidLinkFormat = errors.New("Invalid link format")

type Link struct {
	sql.NullString
}

func (l Link) MarshalJSON() ([]byte, error) {
	if l.Valid {
		return json.Marshal(l.String)
	} else {
		return json.Marshal(nil)
	}
}

func (l *Link) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		l.Valid = true
		l.String = *x
	} else {
		l.Valid = false
	}
	return nil
}
