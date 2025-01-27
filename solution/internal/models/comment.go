package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Comment struct {
	CommentId *string `query:"comment_id" json:"id" db:"id" validate:"required,uuid"`
	Text      *string `json:"text" db:"text" redis:"text" validate:"required,gte=10,lte=1000"`
	Date      *string `json:"date" db:"date" redis:"date" `
	Author    *Author `json:"author" db:"author" redis:"author" `
}
type Author struct {
	Name      *string `json:"name" db:"name" redis:"name" validate:"required,gte=1,lte=100"`
	SurName   *string `json:"surname" db:"surname"  redis:"surname" validate:"required,gte=1,lte=120"`
	AvatarUrl *string `json:"avatar_url,omitempty" db:"avatar_url,omitempty"  redis:"avatar_url,omitempty" validate:"omitempty,url,lte=350"`
}

func (t Author) Value() (driver.Value, error) {
	return json.Marshal(t)
}
func (t *Author) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &t)
}
