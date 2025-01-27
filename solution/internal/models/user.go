package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type RedisUser struct {
	ID        *string  `json:"id" db:"company_id" redis:"id"`
	Name      *string  `json:"name" db:"name" redis:"name" validate:"required,gte=1,lte=100"`
	SurName   *string  `json:"surname" db:"surname"  redis:"surname" validate:"required,gte=1,lte=120"`
	Email     *string  `json:"email" db:"email"  redis:"email" validate:"required,email,gte=8,lte=120"`
	AvatarUrl *string `json:"avatar_url,omitempty" db:"avatar_url,omitempty"  redis:"avatar_url,omitempty" validate:"omitempty,url,lte=350"`
	Age       *int     `json:"age" db:"age" redis:"age" validate:"required,gte=0,lte=100"`
	Country   *string  `json:"country" db:"country" redis:"country" validate:"required,country_validation"`
	Password  []byte  `json:"password" db:"password"  redis:"password" validate:"required,gte=8,lte=60,password"`
}
type User struct {
	ID        *string  `json:"id" db:"company_id" redis:"id"`
	Name      *string  `json:"name" db:"name" redis:"name" validate:"required,gte=1,lte=100"`
	SurName   *string  `json:"surname" db:"surname"  redis:"surname" validate:"required,gte=1,lte=120"`
	Email     *string  `json:"email" db:"email"  redis:"email" validate:"required,email,gte=8,lte=120"`
	AvatarUrl *string `json:"avatar_url,omitempty" db:"avatar_url,omitempty"  redis:"avatar_url,omitempty" validate:"omitempty,url,lte=350"`
	Other     *Other  `json:"other" db:"other"  redis:"other" validate:"required"`
	Password  []byte  `json:"password" db:"password"  redis:"password" validate:"required,gte=8,lte=60,password"`
}
type Other struct {
	Age     *int    `json:"age" db:"age" redis:"age" validate:"required,gte=0,lte=100"`
	Country *string `json:"country" db:"country" redis:"country" validate:"required,country_validation"`
}

func (t Other) Value() (driver.Value, error) {
	return json.Marshal(t)
}
func (t *Other) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &t)
}
