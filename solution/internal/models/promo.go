package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Promo struct {
	Description     *string     `json:"description" db:"description" validate:"required,gte=10,lte=300"`
	ImageUrl        *string     `json:"image_url,omitempty" db:"image_url,omitempty" validate:"omitempty,lte=350,url"`
	Target          *Target     `json:"target" db:"target" validate:"required,gte=0"`
	MaxCount        *int        `json:"max_count" db:"max_count" validate:"required" `
	ActiveFrom      *int64      `json:"active_from,omitempty" db:"active_from,omitempty" validate:"omitempty,date_validation"`
	ActiveUntil     *int64      `json:"active_until,omitempty" db:"active_until,omitempty" validate:"omitempty,date_validation"`
	Mode            *string     `json:"mode" db:"mode" validate:"required,oneof='COMMON' 'UNIQUE'"`
	PromoCommon     *string     `json:"promo_common,omitempty" db:"promo_common,omitempty" validate:"omitempty,required_if=Mode COMMON,gte=5,lte=30"`
	PromoUnique     StringSlice `json:"promo_unique,omitempty" db:"promo_unique,omitempty" validate:"omitempty,required_if=Mode UNIQUE,gte=1,lte=5000,dive,gte=3,lte=30"`
	UsedPromoUnique StringSlice
	PromoId         *string `json:"promo_id" db:"promo_id" validate:"required,uuid"`
	CompanyId       *string `json:"company_id" db:"company_id" validate:"required,uuid"`
	CompanyName     *string `json:"company_name" db:"company_name" validate:"required,gte=5,lte=50"`
	LikeCount       *int    `json:"like_count" db:"like_count" validate:"required"`
	UsedCount       *int    `json:"used_count" db:"used_count" validate:"required"`
	CommentCount    int    `json:"comment_count" db:"comment_count" `
	Active          *bool   `json:"active" db:"active" `
}

type Target struct {
	AgeFrom    *int        `json:"age_from,omitempty" db:"age_from,omitempty" validate:"omitempty,gte=0,lte=100"`
	AgeUntil   *int        `json:"age_until,omitempty" db:"age_until,omitempty" validate:"omitempty,gte=0,lte=100"`
	Country    *string     `json:"country,omitempty" db:"country,omitempty" validate:"omitempty,country_validation"`
	Categories StringSlice `json:"categories,omitempty" db:"categories,omitempty" validate:"omitempty,lte=20,dive,gte=2,lte=20"`
}

// type PromoStat struct {
// 	ID          *string `json:"id" db:"company_id" redis:"id" validate:"required"`
// 	PromoId     *string `json:"promo_id" db:"promo_id" validate:"required,uuid"`
// 	CompanyId   *string `json:"company_id" db:"company_id" validate:"required,uuid"`
// 	IsActivated *bool   `json:"is_activated_by_user" db:"is_activated_by_user" validate:"required"`
// 	IsLiked     *bool   `json:"is_liked_by_user" db:"is_liked_by_user" validate:"required"`
// }

func (t Target) Value() (driver.Value, error) {
	return json.Marshal(t)
}
func (t *Target) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &t)
}

type StringSlice []string

func (p StringSlice) Value() (driver.Value, error) {
	var quotedStrings []string
	for _, s := range p {
		quotedStrings = append(quotedStrings, strconv.Quote(s))
	}
	value := fmt.Sprintf("{ %s }", strings.Join(quotedStrings, ","))
	return value, nil
}
func (p *StringSlice) Scan(scr interface{}) error {
	val, ok := scr.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	value := strings.TrimPrefix(string(val), "{")
	value = strings.TrimSuffix(value, "}")

	*p = strings.Split(value, ",")
	return nil
}
