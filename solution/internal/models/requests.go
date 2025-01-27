package models

import (
	"strings"
)

type CompanySignUpRequest struct {
	Name     string `json:"name"  validate:"required,gte=5,lte=50"`
	Email    string `json:"email" validate:"required,email,gte=8,lte=120"`
	Password string `json:"password" validate:"required,password"`
}
type CompanySignUpResponse struct {
	Token     string `json:"token" validate:"lte=300"`
	CompanyID string `json:"company_id" validate:"uuid"`
}
type CompanySignInRequest struct {
	Email    string `json:"email" validate:"required,email,gte=8,lte=120"`
	Password string `json:"password" validate:"required,password"`
}
type CompanySignInResponse struct {
	Token string `json:"token" validate:"lte=300"`
}
type CreatePromoRequest struct {
	Description *string  `json:"description" db:"description" validate:"required,gte=10,lte=300"`
	ImageUrl    *string  `json:"image_url,omitempty" db:"image_url,omitempty" validate:"omitempty,lte=350,url"`
	Target      *Target  `json:"target" db:"target,required" validate:"required"`
	MaxCount    *int     `json:"max_count" db:"max_count" validate:"required,gte=0,lte=100000000"`
	ActiveFrom  *string  `json:"active_from,omitempty" db:"active_from,omitempty" validate:"omitempty,date_validation"`
	ActiveUntil *string  `json:"active_until,omitempty" db:"active_until,omitempty" validate:"omitempty,date_validation"`
	Mode        *string  `json:"mode" db:"mode" validate:"required,oneof='COMMON' 'UNIQUE'"`
	PromoCommon *string  `json:"promo_common,omitempty" validate:"required_if=Mode COMMON,omitempty,gte=5,lte=30"`
	PromoUnique []string `json:"promo_unique,omitempty" validate:"required_if=Mode UNIQUE,omitempty,gte=1,lte=5000,dive,gte=3,lte=30"`
}

type CreatePromoResponse struct {
	PromoId string `json:"id" validate:"requred,uuid"`
}
type GetPromosRequest struct {
	Limit     *int     `query:"limit" validate:"omitempty,gte=0"`
	Offset    *int     `query:"offset" validate:"omitempty,gte=0"`
	SortBy    *string  `query:"sort_by" validate:"omitempty,oneof='active_from' 'active_until' ' '"`
	Countries []string `query:"country" validate:"omitempty,dive,country_validation"`
}

type GetPromoResponse struct {
	Description *string     `json:"description" db:"description" validate:"required,gte=10,lte=300"`
	ImageUrl    *string     `json:"image_url,omitempty" db:"image_url,omitempty" validate:"omitempty,lte=350,url"`
	Target      *Target     `json:"target" db:"target" validate:"required,gte=0"`
	MaxCount    *int        `json:"max_count" db:"max_count" validate:"required" `
	ActiveFrom  *string     `json:"active_from,omitempty" db:"active_from,omitempty" validate:"omitempty,date_validation"`
	ActiveUntil *string     `json:"active_until,omitempty" db:"active_until,omitempty" validate:"omitempty,date_validation"`
	Mode        *string     `json:"mode" db:"mode" validate:"required,oneof='COMMON' 'UNIQUE'"`
	PromoCommon *string     `json:"promo_common,omitempty" db:"promo_common,omitempty" validate:"omitempty,required_if=Mode COMMON,gte=5,lte=30"`
	PromoUnique StringSlice `json:"promo_unique,omitempty" db:"promo_unique,omitempty" validate:"omitempty,required_if=Mode UNIQUE,gte=1,lte=5000,dive,gte=3,lte=30"`
	PromoId     *string     `json:"promo_id" db:"promo_id" validate:"required,uuid"`
	CompanyId   *string     `json:"company_id" db:"company_id" validate:"required,uuid"`
	CompanyName *string     `json:"company_name" db:"company_name" validate:"required,gte=5,lte=50"`
	LikeCount   *int        `json:"like_count" db:"like_count" validate:"required"`
	UsedCount   *int        `json:"used_count" db:"used_count" validate:"required"`
	Active      *bool       `json:"active" db:"active" `
}
type GetPromoRequest struct {
	ID *string `json:"promo_id" param:"id" validate:"required"`
}
type EditPromoRequest struct {
	ID          *string `json:"promo_id" param:"id" validate:"required"`
	Description *string `json:"description,omitempty" db:"description,omitempty" validate:"omitempty,gte=10,lte=300"`
	ImageUrl    *string `json:"image_url,omitempty" db:"image_url,omitempty" validate:"omitempty,lte=350,url"`
	Target      *Target `json:"target,omitempty" db:"target,omitempty" validate:"omitempty"`
	MaxCount    *int    `json:"max_count,omitempty" db:"max_count,omitempty" validate:"omitempty" `
	ActiveFrom  *string `json:"active_from,omitempty" db:"active_from,omitempty" validate:"omitempty,date_validation"`
	ActiveUntil *string `json:"active_until,omitempty" db:"active_until,omitempty" validate:"omitempty,date_validation"`
}
type GetPromoStatRequest struct {
	PromoID   *string `json:"promo_id" param:"id" validate:"required"`
	CompanyID *string `json:"user_id"  validate:"required"`
}
type GetPromoStatResponse struct {
	ActivationsCount int       `json:"activations_count" db:"activations_count" redis:"actiovations_count" validate:"gte=0"`
	Countries        Countries `json:"countries,omitempty" db:"countries,omitempty" redis:"countries,omitempty" validate:"omitempty"`
}
type Countries []Country
type Country struct {
	Country          string `json:"country" db:"country" redis:"country" validate:"country_validation"`
	ActivationsCount int    `json:"activations_count" db:"activations_count" redis:"activations_count" validate:"gte=1"`
}

func (s Countries) Len() int      { return len(s) }
func (s Countries) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s Countries) Less(i, j int) bool {
	return strings.ToLower(s[i].Country) < strings.ToLower(s[j].Country)
}

type SignUpUserRequest struct {
	Name      *string `json:"name" db:"name" validate:"required,gte=1,lte=100"`
	SurName   *string `json:"surname" db:"surname" validate:"required,gte=1,lte=120"`
	Email     *string `json:"email" db:"email" validate:"required,email,gte=8,lte=120"`
	AvatarUrl *string `json:"avatar_url" db:"avatar_url" validate:"omitempty,url,lte=350"`
	Other     *Other  `json:"other" db:"other" validate:"required"`
	Password  *string `json:"password" db:"password" validate:"required,gte=8,lte=60,password"`
}
type SignUpUserResp struct {
	Token string `json:"token"`
}
type SignInUserRequest struct {
	Email    *string `json:"email" db:"email" validate:"required,email,gte=8,lte=120"`
	Password *string `json:"password" db:"password" validate:"required,gte=8,lte=60,password"`
}
type FeedUserRequest struct {
	Limit    *int    `query:"limit" validate:"omitempty,gte=0"`
	Offset   *int    `query:"offset" validate:"omitempty,gte=0"`
	Category *string `query:"category" validate:"omitempty"`
	Active   *bool   `query:"active" validate:"omitempty"`
}
type EditUserRequest struct {
	ID        *string `json:"id" db:"id" redis:"id" validate:"required"`
	Name      *string `json:"name,omitempty" db:"name,omitempty" validate:"omitempty,gte=1,lte=100"`
	SurName   *string `json:"surname,omitempty" db:"surname,omitempty" validate:"omitempty,gte=1,lte=120"`
	AvatarUrl *string `json:"avatar_url,omitempty" db:"avatar_url,omitempty" validate:"omitempty,url,lte=350"`
	Password  *string `json:"password,omitempty" db:"password,omitempty" validate:"omitempty,gte=8,lte=60,password"`
}
type EditUserResponsestruct struct {
	Name      *string `json:"name" db:"name" validate:"omitempty,gte=1,lte=100"`
	SurName   *string `json:"surname" db:"surname" validate:"omitempty,gte=1,lte=120"`
	Email     *string `json:"email" db:"email" validate:"required,email,gte=8,lte=120"`
	AvatarUrl *string `json:"avatar_url,omitempty" db:"avatar_url,omitempty" validate:"omitempty,url,lte=350"`
	Other     *Other  `json:"other" db:"other" validate:"required"`
}
type GetUserResponse struct {
	ID        *string `json:"id" db:"company_id" redis:"id"`
	Name      *string `json:"name" db:"name" redis:"name" validate:"required,gte=1,lte=100"`
	SurName   *string `json:"surname" db:"surname"  redis:"surname" validate:"required,gte=1,lte=120"`
	Email     *string `json:"email" db:"email"  redis:"email" validate:"required,email,gte=8,lte=120"`
	AvatarUrl *string `json:"avatar_url,omitempty" db:"avatar_url,omitempty"  redis:"avatar_url,omitempty" validate:"omitempty,url,lte=350"`
	Other     *Other  `json:"other" db:"other"  redis:"other" validate:"required"`
}
type FeedUserResponse struct {
	PromoId      *string `json:"promo_id" db:"promo_id" validate:"required,uuid"`
	CompanyId    *string `json:"company_id" db:"company_id" validate:"required,uuid"`
	CompanyName  *string `json:"company_name" db:"company_name" validate:"required,gte=5,lte=50"`
	Description  *string `json:"description" db:"description" validate:"required,gte=10,lte=300"`
	ImageUrl     *string `json:"image_url,omitempty" db:"image_url,omitempty" validate:"omitempty,lte=350,url"`
	Active       *bool   `json:"active" db:"active" `
	IsActivated  *bool   `json:"is_activated_by_user" db:"is_activated_by_user"`
	LikeCount    *int    `json:"like_count" db:"like_count" validate:"required"`
	IsLiked      *bool   `json:"is_liked_by_user" db:"is_liked_by_user"`
	CommentCount *int    `json:"comment_count" db:"comment_count" validate:"required"`
}
type UserPromoRequest struct {
	PromoId *string `param:"id" json:"promo_id" db:"promo_id" validate:"required,uuid"`
	ID      *string `json:"id" db:"id" redis:"id" validate:"required"`
}
type UserLikedPromo struct {
	PromoId   *string `param:"id" json:"promo_id" db:"promo_id" validate:"required,uuid"`
	UserID    *string `json:"id" db:"id" redis:"id" validate:"required"`
	LikeCount *int    `json:"like_count" db:"like_count"`
	IsLiked   *bool   `json:"is_liked_by_user" db:"is_liked_by_user" redis:"is_liked_by_user"`
}
type UserCommentCreateRequest struct {
	CommentId    *string `json:"id" db:"id" validate:"required,uuid"`
	Text         *string `json:"text" db:"text" redis:"text" validate:"required,gte=10,lte=1000"`
	Date         *string `json:"date" db:"date" redis:"date" `
	Author       *Author `json:"author" db:"author" redis:"author" `
	PromoID      *string `param:"id" json:"promo_id" db:"promo_id" validate:"required,uuid"`
	UserID       *string `json:"user_id" db:"user_id" validate:"required"`
	CommentCount *int    `json:"comment_count" db:"comment_count" `
}
type UserGetCommentResponse struct {
	CommentId *string `json:"id" db:"id" validate:"required,uuid"`
	Text      *string `json:"text" db:"text" redis:"text" validate:"required,gte=10,lte=1000"`
	Date      *string `json:"date" db:"date" redis:"date" `
	Author    *Author `json:"author" db:"author" redis:"author" `
}
type UserGetCommentsRequest struct {
	Limit   *int    `query:"limit" validate:"omitempty,gte=0"`
	Offset  *int    `query:"offset" validate:"omitempty,gte=0"`
	PromoID *string `param:"id" json:"promo_id" db:"promo_id" validate:"required,uuid"`
}
type UserGetComment struct {
	CommentId *string `param:"comment_id" json:"id" db:"id" validate:"required,uuid"`
	PromoID   *string `param:"id" json:"promo_id" db:"promo_id" validate:"required,uuid"`
}

type UserEditCommentRequest struct {
	CommentId *string `param:"comment_id" json:"id" db:"id" validate:"required,uuid"`
	Text      *string `json:"text" db:"text" redis:"text" validate:"required,gte=10,lte=1000"`
	PromoID   *string `param:"id" json:"promo_id" db:"promo_id" validate:"required,uuid"`
	UserID    *string `json:"user_id" db:"user_id" validate:"required"`
}
type UserDeleteCommentRequest struct {
	CommentId    *string `param:"comment_id" json:"id" db:"id" validate:"required,uuid"`
	PromoID      *string `param:"id" json:"promo_id" db:"promo_id" validate:"required,uuid"`
	UserID       *string `json:"user_id" db:"user_id" validate:"required"`
	CommentCount int     `json:"comment_count" db:"comment_count"`
}
type UserCheckComments struct {
	CommentId *string `param:"comment_id" json:"id" db:"id" validate:"required,uuid"`
	PromoID   *string `param:"id" json:"promo_id" db:"promo_id" validate:"required,uuid"`
	UserID    *string `json:"user_id" db:"user_id" validate:"required"`
}
type AntifraudRequest struct {
	UserEmail string `json:"user_email"`
	PromoId   string `json:"promo_id"`
}
type AntifraudResponse struct {
	Ok         bool   `json:"ok"`
	CacheUntil string `json:"cache_until,omitempty"`
}
type ActivateRequest struct {
	PromoID *string `param:"id" json:"promo_id" db:"promo_id" validate:"required,uuid"`
	UserID  *string ` json:"used_id" db:"used_id" validate:"required,uuid"`
	Country *string
	Age     *int
}
type UserHistoryRequest struct {
	Limit  *int    `query:"limit" validate:"omitempty,gte=0"`
	Offset *int    `query:"offset" validate:"omitempty,gte=0"`
	UserID *string ` json:"user_id" db:"user_id" validate:"required,uuid"`
}
type SetUserVerdict struct {
	Email *string `json:"user_email"`
	Ok    *bool   `json:"ok"`
}
