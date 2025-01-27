package redisrepository

import (
	"context"
	"solution/internal/models"
	"solution/internal/service"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

func New(client *redis.Client) *RedisRepo {
	return &RedisRepo{client}
}

var expiredTime = 24 * time.Hour

func (rr *RedisRepo) HGetAll(ctx context.Context, key string) (interface{}, error) {
	var value interface{}
	err := rr.client.HGetAll(ctx, key).Scan(value)
	return value, err
}
func (rr *RedisRepo) HGetString(ctx context.Context, key, field string) (string, error) {
	return rr.client.HGet(ctx, key, field).Result()
}
func (rr *RedisRepo) HGetInt(ctx context.Context, key, field string) (int, error) {
	return rr.client.HGet(ctx, key, field).Int()
}
func (rr *RedisRepo) HGetBool(ctx context.Context, key, field string) (bool, error) {
	return rr.client.HGet(ctx, key, field).Bool()
}
func (rr *RedisRepo) HIncrementInt(ctx context.Context, key, field string) error {
	v, err := rr.client.HGet(ctx, key, field).Int()
	if err != nil {
		return err
	}
	v++
	err = rr.client.HSet(ctx, key, field, v).Err()
	if err != nil {
		return err
	}
	return rr.client.HExpire(ctx, key, expiredTime, rr.client.HKeys(ctx, key).Val()...).Err()
}
func (rr *RedisRepo) HSet(ctx context.Context, key string, value interface{}) error {
	err := rr.client.HSet(ctx, key, value).Err()
	if err != nil {
		return err
	}
	return rr.client.HExpire(ctx, key, expiredTime, rr.client.HKeys(ctx, key).Val()...).Err()
}
func (rr *RedisRepo) Set(ctx context.Context, key string, value interface{}) error {
	return rr.client.Set(ctx, key, value, expiredTime).Err()
}
func (rr *RedisRepo) GetString(ctx context.Context, key string) (string, error) {
	return rr.client.Get(ctx, key).Result()
}

func (rr *RedisRepo) GetInt(ctx context.Context, key string) (int, error) {
	return rr.client.Get(ctx, key).Int()
}
func (rr *RedisRepo) AddCompany(ctx context.Context, company models.Company) error {
	err := rr.Set(ctx, company.Email, company.CompanyID)
	if err != nil {
		return err
	}
	err = rr.HSet(ctx, company.CompanyID, company)
	if err != nil {
		return err
	}
	return nil
}
func (rr *RedisRepo) GetCompanyByEmail(ctx context.Context, company models.Company) (*models.Company, error) {
	id, err := rr.GetString(ctx, company.Email)
	if err != nil {
		return nil, err
	}
	var cmp models.Company
	err = rr.client.HGetAll(ctx, id).Scan(&cmp)
	return &cmp, err
}
func (rr *RedisRepo) GetCompanyById(ctx context.Context, company models.Company) (*models.Company, error) {
	var cmp models.Company
	err := rr.client.HGetAll(ctx, company.CompanyID).Scan(&cmp)
	return &cmp, err
}

func (rr *RedisRepo) GetPromoStat(ctx context.Context, promo models.Promo) (*models.GetPromoStatResponse, error) {
	// var countries models.GetPromoStatRequest
	// err := rr.client.HGet(ctx,*promo.PromoId,"stat").Scan()
	var countries models.Countries
	keys, err := rr.client.HKeys(ctx, *promo.PromoId).Result()
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		if len(k) != 2 {
			continue
		}
		activations, err := rr.client.HGet(ctx, *promo.PromoId, k).Int()
		if err != nil {
			return nil, err
		}
		countries = append(countries, models.Country{ActivationsCount: activations, Country: k})
	}

	activationsCount, err := rr.client.HGet(ctx, *promo.PromoId, "activations").Int()
	if err != nil {
		return nil, err
	}
	companyID, err := rr.client.HGet(ctx, *promo.PromoId, "company_id").Result()
	if err != nil {
		return nil, err
	}
	if companyID != *promo.CompanyId {
		return nil, service.ErrNoPermission
	}
	return &models.GetPromoStatResponse{ActivationsCount: activationsCount, Countries: countries}, nil
}

func (rr *RedisRepo) GetUserByEmail(ctx context.Context, User models.User) (*models.User, error) {
	id, err := rr.GetString(ctx, *User.Email)
	if err != nil {
		return nil, err
	}
	var usr models.RedisUser
	err = rr.client.HGetAll(ctx, id).Scan(&usr)
	if err != nil {
		return nil, err
	}
	resp := models.User{
		ID:        usr.ID,
		Name:      usr.Name,
		SurName:   usr.SurName,
		Email:     usr.Email,
		AvatarUrl: usr.AvatarUrl,
		Other:     nil,
		Password:  usr.Password,
	}
	if *usr.Age != 0 || *usr.Country != "" {
		resp.Other = &models.Other{}
		if *usr.Age != 0 {
			resp.Other.Age = usr.Age
		}
		if *usr.Country != "" {
			resp.Other.Country = usr.Country
		}
	}

	return &resp, err
}

func (rr *RedisRepo) GetUserById(ctx context.Context, User models.User) (*models.User, error) {
	var usr models.RedisUser
	err := rr.client.HGetAll(ctx, *User.ID).Scan(&usr)
	if err != nil {
		return nil, err
	}
	resp := models.User{
		ID:        usr.ID,
		Name:      usr.Name,
		SurName:   usr.SurName,
		Email:     usr.Email,
		AvatarUrl: usr.AvatarUrl,
		Other:     nil,
		Password:  usr.Password,
	}
	if *usr.Age != 0 || *usr.Country != "" {
		resp.Other = &models.Other{}
		if *usr.Age != 0 {
			resp.Other.Age = usr.Age
		}
		if *usr.Country != "" {
			resp.Other.Country = usr.Country
		}
	}

	return &resp, err
}
func (rr *RedisRepo) AddUser(ctx context.Context, user *models.RedisUser) error {
	err := rr.Set(ctx, *user.Email, user.ID)
	if err != nil {
		return err
	}

	err = rr.HSet(ctx, *user.ID, user)
	if err != nil {
		return err
	}
	return nil
}
func (rr *RedisRepo) CacheFraud(ctx context.Context, userID, until string, value bool) error {
	t, _ := time.Parse("2006-01-02T15:04:05.000Zhh:mm", until)
	t = t.Add(3 * time.Hour)
	err := rr.client.Set(ctx, userID+"_fraud", value, 0).Err()
	if err != nil {
		return err
	}
	err = rr.client.ExpireAt(ctx, userID+"_fraud", t).Err()
	if err != nil {
		return err
	}
	return nil
}
func (rr *RedisRepo) CheckFraud(ctx context.Context, userID string) (bool, error) {
	value, err := rr.client.Get(ctx, userID+"_fraud").Bool()
	if err != nil {
		return false, err
	}
	return value, nil
}
