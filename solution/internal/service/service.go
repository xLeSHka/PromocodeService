package service

import (
	"context"
	"database/sql"
	"solution/internal/models"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

type PostgresRepo interface {
	TestCompanyRegistration(ctx context.Context, company models.Company) (bool, error)
	GetCompanyByEmail(ctx context.Context, company models.Company) (*models.Company, error)
	GetCompanyById(ctx context.Context, company models.Company) (*models.Company, error)
	AddCompany(ctx context.Context, company models.Company) error
	CreatePromo(ctx context.Context, promo *models.Promo) error
	GetPromos(ctx context.Context, sortRules *models.CompanySort) ([]models.GetPromoResponse, int, error)
	GetPromo(ctx context.Context, promo models.Promo) (*models.GetPromoResponse, error)
	GetPromoById(ctx context.Context, promo models.Promo) (*models.Promo, error)
	GetPromoStat(ctx context.Context, promo models.GetPromoStatRequest) (*models.GetPromoStatResponse, error)
	EditPromo(ctx context.Context, promo *models.Promo) (*models.GetPromoResponse, error)
	TestUserRegistration(ctx context.Context, user models.User) (bool, error)
	AddUser(ctx context.Context, user models.User) error
	GetUserByEmail(ctx context.Context, User models.User) (*models.User, error)
	GetUserById(ctx context.Context, User models.User) (*models.User, error)
	FeedUser(ctx context.Context, sortRules *models.UserSort) ([]models.FeedUserResponse, int, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	UserGetPromo(ctx context.Context, promo models.UserPromoRequest) (*models.FeedUserResponse, error)
	UserLikePromo(ctx context.Context, promo models.UserLikedPromo) error
	UserUpdateLike(ctx context.Context, promo models.UserLikedPromo) error
	UserCreateComment(ctx context.Context, comment models.UserCommentCreateRequest) (*models.Comment, error)
	UserGetComments(ctx context.Context, sortRules *models.CommentSort) ([]models.Comment, int, error)
	UserGetComment(ctx context.Context, comment models.UserGetComment) (*models.Comment, error)
	UserEditComment(ctx context.Context, comment models.UserEditCommentRequest) (*models.Comment, error)
	UserDeleteComment(ctx context.Context, comment models.UserDeleteCommentRequest) error
	UserActivatePromo(ctx context.Context, promo models.ActivateRequest) (string, error)
	CheckISLiked(ctx context.Context, promo models.UserPromoRequest) (bool, error)
	CheckComment(ctx context.Context, comment models.UserCheckComments) (bool, error)
	GetUserHistory(ctx context.Context, sortRules *models.HistorySort) ([]models.FeedUserResponse, int, error)
}
type RedisRepo interface {
	HGetAll(ctx context.Context, key string) (interface{}, error)
	HGetString(ctx context.Context, key, field string) (string, error)
	HGetInt(ctx context.Context, key, field string) (int, error)
	HGetBool(ctx context.Context, key, field string) (bool, error)
	HIncrementInt(ctx context.Context, key, field string) error
	HSet(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}) error
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int, error)
	AddCompany(ctx context.Context, company models.Company) error
	GetCompanyByEmail(ctx context.Context, company models.Company) (*models.Company, error)
	GetCompanyById(ctx context.Context, company models.Company) (*models.Company, error)
	GetPromoStat(ctx context.Context, promo models.Promo) (*models.GetPromoStatResponse, error)
	AddUser(ctx context.Context, user *models.RedisUser) error
	GetUserByEmail(ctx context.Context, User models.User) (*models.User, error)
	GetUserById(ctx context.Context, User models.User) (*models.User, error)
	CacheFraud(ctx context.Context, userID, until string, value bool) error
	CheckFraud(ctx context.Context, userID string) (bool, error)
}
type Service struct {
	redisRepo    RedisRepo
	postgresRepo PostgresRepo
}

func New(redisRepo RedisRepo, postgresRepo PostgresRepo) *Service {
	return &Service{redisRepo: redisRepo, postgresRepo: postgresRepo}
}
func (s *Service) CompanySignUp(ctx context.Context, company models.Company) error {
	id, redisErr := s.redisRepo.GetString(ctx, company.Email)
	if redisErr != nil && redisErr != redis.Nil {
		return redisErr
	}
	if id != "" {
		return ErrEmailRegistrated
	}
	registrated2, err := s.postgresRepo.TestCompanyRegistration(ctx, company)
	if registrated2 {
		if id == "" {
			err = s.redisRepo.AddCompany(ctx, company)
			if err != nil {
				return err
			}
		}
		return ErrEmailRegistrated
	}
	if err != nil {
		return err
	}
	err = s.postgresRepo.AddCompany(ctx, company)
	if err != nil {
		return err
	}
	return s.redisRepo.AddCompany(ctx, company)

}
func (s *Service) CompanySignIn(ctx context.Context, company models.Company) (*models.Company, error) {
	cmp, redisErr := s.redisRepo.GetCompanyByEmail(ctx, company)

	if redisErr != nil && redisErr != redis.Nil {
		return nil, redisErr
	}
	if cmp != nil {
		return cmp, nil
	}
	cmp, err := s.postgresRepo.GetCompanyByEmail(ctx, company)
	if cmp != nil {
		if redisErr == redis.Nil {
			err = s.redisRepo.AddCompany(ctx, company)
			if err != nil {
				return nil, err
			}
		}
		return cmp, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	return nil, ErrEmailNotRegistrated
}
func (s *Service) UpdateToken(ctx context.Context, id, token string) error {
	return s.redisRepo.HSet(ctx, id, map[string]string{"token": token})
}

func (s *Service) GetToken(ctx context.Context, id string) (string, error) {
	return s.redisRepo.HGetString(ctx, id, "token")
}
func (s *Service) CreatePromo(ctx context.Context, promo *models.Promo) error {
	var err error
	cmp, err := s.redisRepo.GetCompanyById(ctx, models.Company{CompanyID: *promo.CompanyId})
	if err != nil && err != redis.Nil {
		return err
	}
	if cmp == nil {
		cmp, err = s.postgresRepo.GetCompanyById(ctx, models.Company{CompanyID: *promo.CompanyId})
		if err != nil {
			return err
		}
		err = s.redisRepo.AddCompany(ctx, *cmp)
		if err != nil {
			return err
		}
	}
	promo.CompanyName = &cmp.Name
	err = s.postgresRepo.CreatePromo(ctx, promo)
	if err != nil {
		return err
	}
	err = s.redisRepo.HSet(ctx, *promo.PromoId, map[string]interface{}{"likes": *promo.LikeCount, "used": *promo.UsedCount, "active": *promo.Active, "company_id": *promo.CompanyId})
	if err != nil {
		return err
	}
	return nil
}
func (s *Service) GetPromos(ctx context.Context, sortRules *models.CompanySort) ([]models.GetPromoResponse, int, error) {
	return s.postgresRepo.GetPromos(ctx, sortRules)
}
func (s *Service) GetPromo(ctx context.Context, promo models.Promo) (*models.GetPromoResponse, error) {
	companyID, redisErr := s.redisRepo.HGetString(ctx, *promo.PromoId, "company_id")
	if redisErr != nil && redisErr != redis.Nil {
		return nil, redisErr
	}
	if companyID != "" {
		if companyID != *promo.CompanyId {
			return nil, ErrNoPermission
		}
	}
	getted, err := s.postgresRepo.GetPromo(ctx, promo)
	if err != nil {
		return nil, ErrPromoNotFound
	}
	if *getted.CompanyId != *promo.CompanyId {
		return nil, ErrNoPermission
	}
	if redisErr == redis.Nil {
		err = s.redisRepo.HSet(ctx, *getted.PromoId, map[string]interface{}{"likes": *getted.LikeCount, "used": *getted.UsedCount, "active": *getted.Active, "company_id": *getted.CompanyId})
		if err != nil {
			return nil, err
		}
	}
	return getted, nil
}
func (s *Service) EditPromo(ctx context.Context, promo *models.Promo) (*models.GetPromoResponse, error) {
	companyID, redisErr := s.redisRepo.HGetString(ctx, *promo.PromoId, "company_id")
	if redisErr != nil && redisErr != redis.Nil {
		return nil, redisErr
	}
	if companyID != "" {
		if companyID != *promo.CompanyId {
			return nil, ErrNoPermission
		}
	}
	getted, err := s.postgresRepo.GetPromoById(ctx, *promo)
	if err != nil {
		return nil, ErrPromoNotFound
	}
	if *getted.CompanyId != *promo.CompanyId {
		return nil, ErrNoPermission
	}
	var count, tFrom, tUntil bool = *getted.Active, *getted.Active, *getted.Active
	if *getted.Mode == "UNIQUE" {
		if promo.MaxCount != nil {
			if *promo.MaxCount != 1 {
				return nil, ErrInvalidMaxCount
			}
		}
		count = len(getted.PromoUnique) > len(getted.UsedPromoUnique)
	} else {
		if promo.MaxCount != nil {
			count = *promo.MaxCount > *getted.UsedCount
		}
	}

	if promo.ActiveUntil != nil {
		tUntil = *promo.ActiveUntil >= time.Now().UTC().Add(3*time.Hour).Unix()
	}
	if promo.ActiveFrom != nil {
		tFrom = *promo.ActiveFrom <= time.Now().UTC().Add(3*time.Hour).Unix()
	}
	active := count && tFrom && tUntil
	promo.Active = &active
	edited, err := s.postgresRepo.EditPromo(ctx, promo)
	if err != nil {
		return nil, err
	}
	if redisErr == redis.Nil {
		err = s.redisRepo.HSet(ctx, *edited.PromoId, map[string]interface{}{"likes": *edited.LikeCount, "used": *edited.UsedCount, "active": *edited.Active, "company_id": *edited.CompanyId})
		if err != nil {
			return nil, err
		}
	}
	return edited, nil
}
func (s *Service) GetPromoStat(ctx context.Context, promo models.GetPromoStatRequest) (*models.GetPromoStatResponse, error) {
	company, err := s.redisRepo.GetCompanyById(ctx, models.Company{CompanyID: *promo.CompanyID})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoNotFound
		}
		return nil, err
	}
	if company.CompanyID != *promo.CompanyID {
		return nil, ErrNoPermission
	}
	promoStat, err := s.postgresRepo.GetPromoStat(ctx, promo)
	if err != nil {
		return nil, err
	}

	sort.Sort(promoStat.Countries)
	return promoStat, nil
}
func (s *Service) UserSignUp(ctx context.Context, user models.User) error {
	id, redisErr := s.redisRepo.GetString(ctx, *user.Email)
	if redisErr != nil && redisErr != redis.Nil {
		return redisErr
	}
	if id != "" {
		return ErrEmailRegistrated
	}
	registrated2, err := s.postgresRepo.TestUserRegistration(ctx, user)
	if registrated2 {
		if id == "" {
			redisusr := models.RedisUser{ID: user.ID, Name: user.Name,
				SurName: user.SurName, Email: user.Email, AvatarUrl: user.AvatarUrl, Password: user.Password}
			if user.Other != nil {
				redisusr.Age = user.Other.Age
				redisusr.Country = user.Other.Country
			}
			err = s.redisRepo.AddUser(ctx, &redisusr)
			if err != nil {
				return err
			}
		}
		return ErrEmailRegistrated
	}
	if err != nil {
		return err
	}
	err = s.postgresRepo.AddUser(ctx, user)
	if err != nil {
		return err
	}
	redisusr := models.RedisUser{ID: user.ID, Name: user.Name,
		SurName: user.SurName, Email: user.Email, AvatarUrl: user.AvatarUrl, Password: user.Password}
	if user.Other != nil {
		redisusr.Age = user.Other.Age
		redisusr.Country = user.Other.Country
	}
	return s.redisRepo.AddUser(ctx, &redisusr)
}
func (s *Service) UserSignIn(ctx context.Context, user models.User) (*models.User, error) {
	cmp, redisErr := s.redisRepo.GetUserByEmail(ctx, user)

	if redisErr != nil && redisErr != redis.Nil {
		return nil, redisErr
	}
	if cmp != nil {
		return cmp, nil
	}
	cmp, err := s.postgresRepo.GetUserByEmail(ctx, user)
	if cmp != nil {
		if redisErr == redis.Nil {
			redisusr := models.RedisUser{ID: cmp.ID, Name: cmp.Name,
				SurName: cmp.SurName, Email: user.Email, AvatarUrl: cmp.AvatarUrl, Password: cmp.Password}
			if cmp.Other != nil {
				redisusr.Age = cmp.Other.Age
				redisusr.Country = cmp.Other.Country
			}
			err = s.redisRepo.AddUser(ctx, &redisusr)
			if err != nil {
				return nil, err
			}
		}
		return cmp, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	return nil, ErrEmailNotRegistrated
}

func (s *Service) GetUser(ctx context.Context, user models.User) (*models.User, error) {
	cmp, redisErr := s.redisRepo.GetUserById(ctx, user)
	if cmp != nil {
		return cmp, nil
	}
	if redisErr != nil && redisErr != redis.Nil {
		return nil, redisErr
	}
	cmp, err := s.postgresRepo.GetUserById(ctx, user)
	if cmp != nil {
		if redisErr == redis.Nil {
			redisusr := models.RedisUser{ID: user.ID, Name: cmp.Name,
				SurName: cmp.SurName, Email: cmp.Email, AvatarUrl: cmp.AvatarUrl, Password: cmp.Password}
			if cmp.Other != nil {
				redisusr.Age = cmp.Other.Age
				redisusr.Country = cmp.Other.Country
			}
			err = s.redisRepo.AddUser(ctx, &redisusr)
			if err != nil {
				return nil, err
			}
		}
		return cmp, nil
	}
	return nil, err
}
func (s *Service) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	edited, err := s.postgresRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	redisusr := models.RedisUser{ID: user.ID, Name: edited.Name,
		SurName: edited.SurName, Email: edited.Email, AvatarUrl: edited.AvatarUrl, Password: edited.Password}
	if edited.Other != nil {
		redisusr.Age = edited.Other.Age
		redisusr.Country = edited.Other.Country
	}
	err = s.redisRepo.AddUser(ctx, &redisusr)
	if err != nil {
		return nil, err
	}

	return edited, nil
}
func (s *Service) FeedUser(ctx context.Context, sortRules *models.UserSort) ([]models.FeedUserResponse, int, error) {
	var user *models.User
	var redisErr, err error

	user, redisErr = s.redisRepo.GetUserById(ctx, models.User{ID: &sortRules.Id})
	if redisErr != nil && redisErr != redis.Nil {
		return nil, 0, redisErr
	}
	if user == nil {
		user, err = s.postgresRepo.GetUserById(ctx, models.User{ID: &sortRules.Id})
		if err != nil {
			return nil, 0, err
		}

		redisusr := models.RedisUser{ID: &sortRules.Id, Name: user.Name,
			SurName: user.SurName, Email: user.Email, AvatarUrl: user.AvatarUrl, Password: user.Password}
		if user.Other != nil {
			redisusr.Age = user.Other.Age
			redisusr.Country = user.Other.Country
		}
		err = s.redisRepo.AddUser(ctx, &redisusr)
		if err != nil {
			return nil, 0, err
		}
	}
	sortRules.Country = *user.Other.Country
	sortRules.Age = *user.Other.Age
	return s.postgresRepo.FeedUser(ctx, sortRules)

}
func (s *Service) UserGetPromo(ctx context.Context, promo models.UserPromoRequest) (*models.FeedUserResponse, error) {

	promocode, newprod := s.postgresRepo.UserGetPromo(ctx, promo)
	if promocode != nil {
		return promocode, nil
	}
	if newprod != nil {
		if newprod == sql.ErrNoRows {
			return nil, ErrPromoNotFound
		}
		return nil, newprod
	}
	return promocode, nil
}
func (s *Service) UserLikePromo(ctx context.Context, promo models.UserLikedPromo) error {
	promocode, err := s.postgresRepo.GetPromoById(ctx, models.Promo{PromoId: promo.PromoId})
	if err != nil {
		if err == ErrPromoNotFound {
			return ErrPromoNotFound
		}
		return err
	}
	tru := true
	likeCount := *promocode.LikeCount + 1
	promo.LikeCount = &likeCount
	promo.IsLiked = &tru
	check, err := s.postgresRepo.CheckISLiked(ctx, models.UserPromoRequest{PromoId: promo.PromoId, ID: promo.UserID})
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		if check {
			return nil
		}
		return s.postgresRepo.UserUpdateLike(ctx, promo)
	}
	return s.postgresRepo.UserLikePromo(ctx, promo)
}
func (s *Service) UserDeleteLike(ctx context.Context, promo models.UserLikedPromo) error {
	promocode, err := s.postgresRepo.GetPromoById(ctx, models.Promo{PromoId: promo.PromoId})
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPromoNotFound
		}
		return err
	}
	fls := false
	updateCount := *promocode.LikeCount - 1
	createCount := *promocode.LikeCount
	promo.IsLiked = &fls
	check, err := s.postgresRepo.CheckISLiked(ctx, models.UserPromoRequest{PromoId: promo.PromoId, ID: promo.UserID})
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		if !check {
			return nil
		}
		promo.LikeCount = &updateCount
		return s.postgresRepo.UserUpdateLike(ctx, promo)
	}
	promo.LikeCount = &createCount
	return s.postgresRepo.UserLikePromo(ctx, promo)
}
func (s *Service) UserCreateComment(ctx context.Context, comment models.UserCommentCreateRequest) (*models.Comment, error) {
	promo, err := s.postgresRepo.GetPromoById(ctx, models.Promo{PromoId: comment.PromoID})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoNotFound
		}
		return nil, err
	}
	user, err := s.GetUser(ctx, models.User{ID: comment.UserID})
	if err != nil {
		return nil, err
	}

	commentCount := promo.CommentCount + 1
	comment.CommentCount = &commentCount
	author := models.Author{
		Name:      user.Name,
		SurName:   user.SurName,
		AvatarUrl: user.AvatarUrl,
	}
	comment.Author = &author
	return s.postgresRepo.UserCreateComment(ctx, comment)
}
func (s *Service) UserGetComments(ctx context.Context, sortRules *models.CommentSort) ([]models.Comment, int, error) {
	_, err := s.postgresRepo.GetPromoById(ctx, models.Promo{PromoId: &sortRules.PromoId})
	if err != nil {
		return nil, 0, ErrPromoNotFound
	}
	return s.postgresRepo.UserGetComments(ctx, sortRules)
}
func (s *Service) UserGetComment(ctx context.Context, comment models.UserGetComment) (*models.Comment, error) {
	_, err := s.postgresRepo.GetPromoById(ctx, models.Promo{PromoId: comment.PromoID})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoNotFound
		}
		return nil, err
	}
	comm, err := s.postgresRepo.UserGetComment(ctx, comment)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoNotFound
		}
		return nil, err
	}
	return comm, nil
}
func (s *Service) UserEditComment(ctx context.Context, comment models.UserEditCommentRequest) (*models.Comment, error) {
	_, err := s.postgresRepo.GetPromoById(ctx, models.Promo{PromoId: comment.PromoID})
	if err != nil {
		return nil, ErrPromoNotFound
	}
	ok, err := s.postgresRepo.CheckComment(ctx, models.UserCheckComments{PromoID: comment.PromoID, UserID: comment.UserID, CommentId: comment.CommentId})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoNotFound
		}
		return nil, err
	}
	if !ok {
		return nil, ErrNoPermission
	}
	comm, err := s.postgresRepo.UserEditComment(ctx, comment)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPromoNotFound
		}
		return nil, err
	}
	return comm, nil
}
func (s *Service) UserDeleteComment(ctx context.Context, comment models.UserDeleteCommentRequest) error {
	promo, err := s.postgresRepo.GetPromoById(ctx, models.Promo{PromoId: comment.PromoID})
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPromoNotFound
		}
		return err
	}
	ok, err := s.postgresRepo.CheckComment(ctx, models.UserCheckComments{PromoID: comment.PromoID, UserID: comment.UserID, CommentId: comment.CommentId})
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPromoNotFound
		}
		return err
	}
	if !ok {
		return ErrNoPermission
	}
	comment.CommentCount = promo.CommentCount - 1
	err = s.postgresRepo.UserDeleteComment(ctx, comment)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrPromoNotFound
		}
		return err
	}
	return nil
}
func (s *Service) CheckCache(ctx context.Context, userID string) (bool, error) {
	return s.redisRepo.CheckFraud(ctx, userID)
}
func (s *Service) Cache(ctx context.Context, userID, until string, value bool) error {
	return s.redisRepo.CacheFraud(ctx, userID, until, value)
}
func (s *Service) UserActivatePromo(ctx context.Context, promo models.ActivateRequest) (string, error) {
	_, err := s.postgresRepo.GetPromoById(ctx, models.Promo{PromoId: promo.PromoID})
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrPromoNotFound
		}
		return "", err
	}
	user, err := s.GetUser(ctx, models.User{ID: promo.UserID})
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNoPermission
		}
		return "", err
	}
	promo.Age = user.Other.Age
	promo.Country = user.Other.Country
	code, err := s.postgresRepo.UserActivatePromo(ctx, promo)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNoPermission
		}
		return "", err
	}
	return code, nil
}
func (s *Service) GetUserHistory(ctx context.Context, sortRules *models.HistorySort) ([]models.FeedUserResponse, int, error) {
	return s.postgresRepo.GetUserHistory(ctx, sortRules)
}
