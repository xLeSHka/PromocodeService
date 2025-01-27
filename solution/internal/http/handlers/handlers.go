package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"solution/internal/models"
	"solution/internal/service"
	"solution/internal/utils"
	"solution/pkg/logger"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Service interface {
	CompanySignUp(ctx context.Context, company models.Company) error
	CompanySignIn(ctx context.Context, company models.Company) (*models.Company, error)
	UpdateToken(ctx context.Context, id, token string) error
	GetToken(ctx context.Context, id string) (string, error)
	CreatePromo(ctx context.Context, promo *models.Promo) error
	GetPromos(ctx context.Context, sortRules *models.CompanySort) ([]models.GetPromoResponse, int, error)
	GetPromo(ctx context.Context, promo models.Promo) (*models.GetPromoResponse, error)
	GetPromoStat(ctx context.Context, promo models.GetPromoStatRequest) (*models.GetPromoStatResponse, error)
	EditPromo(ctx context.Context, promo *models.Promo) (*models.GetPromoResponse, error)
	UserSignUp(ctx context.Context, user models.User) error
	UserSignIn(ctx context.Context, user models.User) (*models.User, error)
	GetUser(ctx context.Context, user models.User) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
	FeedUser(ctx context.Context, sortRules *models.UserSort) ([]models.FeedUserResponse, int, error)
	UserGetPromo(ctx context.Context, promo models.UserPromoRequest) (*models.FeedUserResponse, error)
	UserLikePromo(ctx context.Context, promo models.UserLikedPromo) error
	UserDeleteLike(ctx context.Context, promo models.UserLikedPromo) error
	UserCreateComment(ctx context.Context, comment models.UserCommentCreateRequest) (*models.Comment, error)
	UserGetComments(ctx context.Context, sortRules *models.CommentSort) ([]models.Comment, int, error)
	UserGetComment(ctx context.Context, comment models.UserGetComment) (*models.Comment, error)
	UserEditComment(ctx context.Context, comment models.UserEditCommentRequest) (*models.Comment, error)
	UserDeleteComment(ctx context.Context, comment models.UserDeleteCommentRequest) error
	CheckCache(ctx context.Context, userID string) (bool, error)
	Cache(ctx context.Context, userID, until string, value bool) error
	UserActivatePromo(ctx context.Context, promo models.ActivateRequest) (string, error)
	GetUserHistory(ctx context.Context, sortRules *models.HistorySort) ([]models.FeedUserResponse, int, error)
}
type Handlers struct {
	service          Service
	SigningKey       string
	AntifraudAddress string
	CryptoKey        []byte
	validate         *validator.Validate
	logger.Logger
}

func New(srv Service, SigningKey, AntifraudAddress string, CryptoKey []byte, validate *validator.Validate, l logger.Logger) *Handlers {
	return &Handlers{srv, SigningKey, AntifraudAddress, CryptoKey, validate, l}
}
func (h *Handlers) Ping(c echo.Context) error {
	return c.JSON(200, echo.Map{"status": "PROOOOOOOOOOOOOOOOOD"})
}
func (h *Handlers) BussinessAuthJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		bearerToken := c.Request().Header.Get("Authorization")
		splitToken := strings.Split(bearerToken, " ")
		if len(splitToken) != 2 {
			h.Error(c.Request().Context(), "authorization header not valid")
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}

		company, err := utils.VerifyToken(splitToken[1], h.SigningKey)

		if err != nil {
			h.Error(c.Request().Context(), "", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}
		token, err := h.service.GetToken(c.Request().Context(), company.ID)
		if err != nil {
			h.Error(c.Request().Context(), "", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}
		if token != splitToken[1] {
			h.Error(c.Request().Context(), "token not valid", zap.String("expected", token), zap.String("recieved", splitToken[1]))
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}
		if company.Role != "company" {
			h.Error(c.Request().Context(), "access deny, forbidden")
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}
		c.Set("user", company)
		return next(c)
	}
}
func (h *Handlers) UserAuthJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		bearerToken := c.Request().Header.Get("Authorization")
		splitToken := strings.Split(bearerToken, " ")
		if len(splitToken) != 2 {
			h.Error(c.Request().Context(), "authorization header not valid")
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}

		user, err := utils.VerifyToken(splitToken[1], h.SigningKey)
		if err != nil {
			h.Error(c.Request().Context(), "", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}
		token, err := h.service.GetToken(c.Request().Context(), user.ID)
		if err != nil {
			h.Error(c.Request().Context(), "", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}
		if token != splitToken[1] {
			h.Error(c.Request().Context(), "token not valid", zap.String("expected", token), zap.String("recieved", splitToken[1]))
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}
		if user.Role != "user" {
			h.Error(c.Request().Context(), "access deny, forbidden")
			return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
				"status":  "error",
				"message": "Пользователь не авторизован.",
			})
		}
		c.Set("user", user)
		return next(c)
	}
}
func (h *Handlers) RequestTimeMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		t := time.Now()
		next(c)

		after := time.Since(t).Milliseconds()
		h.Info(c.Request().Context(), "request time", zap.Int64("milliseconds", after))
		return nil
	}
}

func (h *Handlers) BusinessSignUp(c echo.Context) error {
	var body models.CompanySignUpRequest
	if c.Request().Header.Get("Content-Type") != "application/json" {
		h.Error(c.Request().Context(), "content-type now allowed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := c.Bind(&body); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	err := h.validate.Struct(&body)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	hashedPassword, err := utils.Encrypt([]byte(body.Password), h.CryptoKey)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}

	company := models.Company{
		CompanyID: uuid.NewString(),
		Name:      body.Name,
		Email:     body.Email,
		Password:  hashedPassword,
	}
	err = h.service.CompanySignUp(c.Request().Context(), company)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrEmailRegistrated {
			return echo.NewHTTPError(409, echo.Map{
				"status":  "error",
				"message": "Такой email уже зарегистрирован.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	token, err := utils.CreateToken(company.CompanyID, "company", h.SigningKey)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	err = h.service.UpdateToken(c.Request().Context(), company.CompanyID, token)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, models.CompanySignUpResponse{
		Token:     token,
		CompanyID: company.CompanyID,
	})
}
func (h *Handlers) BusinessSignIn(c echo.Context) error {
	var body models.CompanySignInRequest
	if c.Request().Header.Get("Content-Type") != "application/json" {
		h.Error(c.Request().Context(), "content-type now allowed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := c.Bind(&body); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	err := h.validate.Struct(&body)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	company := models.Company{
		Email: body.Email,
	}
	cmp, err := h.service.CompanySignIn(c.Request().Context(), company)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
			"status":  "error",
			"message": "Неверный email или пароль.",
		})
	}
	decryptedPassword, err := utils.Decrypt(cmp.Password, h.CryptoKey)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
			"status":  "error",
			"message": "Неверный email или пароль.",
		})
	}
	if !bytes.Equal(decryptedPassword, []byte(body.Password)) {
		h.Error(c.Request().Context(), "password not match", zap.Binary("decrypted", decryptedPassword), zap.Binary("body password", []byte(body.Password)))
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
			"status":  "error",
			"message": "Неверный email или пароль.",
		})
	}
	token, err := utils.CreateToken(cmp.CompanyID, "company", h.SigningKey)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	err = h.service.UpdateToken(c.Request().Context(), cmp.CompanyID, token)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, models.CompanySignInResponse{
		Token: token,
	})
}

func (h *Handlers) BussinessCreatePromo(c echo.Context) error {

	user := c.Get("user").(*utils.JWTClaims)
	if c.Request().Header.Get("Content-Type") != "application/json" {
		h.Error(c.Request().Context(), "content-type now allowed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	var body models.CreatePromoRequest
	if err := c.Bind(&body); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}

	if err := h.validate.Struct(body); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if *body.Mode == "UNIQUE" {
		if *body.MaxCount != 1 {
			h.Error(c.Request().Context(), "invalid max count")
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "Ошибка в данных запроса.",
			})
		}
		if body.PromoCommon != nil {
			h.Error(c.Request().Context(), "invalid promo mode")
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "Ошибка в данных запроса.",
			})
		}
	}
	if *body.Mode == "COMMON" {
		if body.PromoUnique != nil {
			h.Error(c.Request().Context(), "invalid promo mode")
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "Ошибка в данных запроса.",
			})
		}
	}
	if body.Target.AgeFrom != nil && body.Target.AgeUntil != nil {
		if *body.Target.AgeFrom > *body.Target.AgeUntil {
			h.Error(c.Request().Context(), "invalid from age")
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "Ошибка в данных запроса.",
			})
		}
	}
	var tFrom, tUntil int64
	if body.ActiveUntil != nil {
		now := time.Now().UTC().Add(3 * time.Hour)
		if t, _ := time.Parse(utils.TimeFormat, *body.ActiveUntil); t.Day() < now.Day() && t.Year() < now.Year() && t.Month() < now.Month() {
			h.Error(c.Request().Context(), "promocode expired")
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "Ошибка в данных запроса.",
			})
		} else {
			tUntil = t.Unix()
		}
	}
	if body.ActiveFrom != nil {
		t, _ := time.Parse(utils.TimeFormat, *body.ActiveFrom)
		tFrom = t.Unix()

	}
	if body.ActiveFrom != nil && body.ActiveUntil != nil {
		if tUntil < tFrom {
			h.Error(c.Request().Context(), "until less then from")
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "Ошибка в данных запроса.",
			})
		}
	}
	// if body.Target.Country == nil {
	// 	str := ""
	// 	body.Target.Country = &str
	// }
	promoID := uuid.NewString()
	likeCount, usedCount := 0, 0
	now := time.Now().UTC().Add(3 * time.Hour).Unix()
	until, from := true, true
	if tUntil != 0 {
		until = tUntil >= now
	}
	if tFrom != 0 {
		from = tFrom <= now
	}
	active := *body.MaxCount > usedCount && until && from
	promo := models.Promo{
		Description: body.Description,
		ImageUrl:    body.ImageUrl,
		MaxCount:    body.MaxCount,
		Target:      body.Target,
		Mode:        body.Mode,
		PromoCommon: body.PromoCommon,
		PromoUnique: body.PromoUnique,
		PromoId:     &promoID,
		CompanyId:   &user.ID,
		LikeCount:   &likeCount,
		UsedCount:   &usedCount,
		Active:      &active,
	}
	if tFrom == 0 {
		promo.ActiveFrom = nil
	} else {
		promo.ActiveFrom = &tFrom
	}
	if tUntil == 0 {
		promo.ActiveUntil = nil
	} else {
		promo.ActiveUntil = &tUntil
	}
	if len(promo.Target.Categories) == 0 {
		promo.Target.Categories = nil
	}
	err := h.service.CreatePromo(c.Request().Context(), &promo)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(201, models.CreatePromoResponse{
		PromoId: *promo.PromoId,
	})
}
func (h *Handlers) BussinessGetPromos(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	baseSort := models.CompanySort{
		CompanyId: user.ID,
		Limit:     10,
		Offset:    0,
		SortBy:    "",
		Countries: nil,
	}
	var req models.GetPromosRequest
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if req.Limit != nil {
		baseSort.Limit = *req.Limit
	}
	if req.Offset != nil {
		baseSort.Offset = *req.Offset
	}
	if req.SortBy != nil {
		baseSort.SortBy = *req.SortBy
	}
	if req.Countries != nil {
		baseSort.Countries = make([]string, 0)
		for _, contry := range req.Countries {
			baseSort.Countries = append(baseSort.Countries, strings.ToLower(contry))
		}
	}
	promos, total, err := h.service.GetPromos(c.Request().Context(), &baseSort)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	totalCount := fmt.Sprintf("%d", total)
	c.Response().Header().Add("X-Total-Count", totalCount)

	return c.JSON(200, promos)
}
func (h *Handlers) BussinessGetPromo(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	var req models.GetPromoRequest
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	promo, err := h.service.GetPromo(c.Request().Context(), models.Promo{PromoId: req.ID, CompanyId: &user.ID})
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrNoPermission {
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{
				"status":  "error",
				"message": "Промокод не принадлежит этой компании.",
			})
		}
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Промокод не найден.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	c.JSON(200, promo)
	return nil
}
func (h *Handlers) BussinessEditPromo(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	var req models.EditPromoRequest
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	h.Info(c.Request().Context(), "", zap.Any("req", req))
	var tFrom, tUntil int64
	if req.ActiveFrom != nil {
		t, _ := time.Parse(utils.TimeFormat, *req.ActiveFrom)
		tFrom = t.Unix()
	}
	if req.ActiveUntil != nil {

		t, _ := time.Parse(utils.TimeFormat, *req.ActiveUntil)
		tUntil = t.Unix()
	}
	if req.ActiveFrom != nil && req.ActiveUntil != nil {
		if tUntil < tFrom {
			h.Error(c.Request().Context(), "until less then from")
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "Ошибка в данных запроса.",
			})
		}
	}
	if req.Target != nil {

		if req.Target.AgeFrom != nil && req.Target.AgeUntil != nil {
			if *req.Target.AgeFrom > *req.Target.AgeUntil {
				h.Error(c.Request().Context(), "invalid from age")
				return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
					"status":  "error",
					"message": "Ошибка в данных запроса.",
				})
			}
		}
	}

	promo := models.Promo{
		CompanyId:   &user.ID,
		PromoId:     req.ID,
		Description: req.Description,
		ImageUrl:    req.ImageUrl,
		Target:      req.Target,
		MaxCount:    req.MaxCount,
	}
	if tFrom == 0 {
		promo.ActiveFrom = nil
	} else {
		promo.ActiveFrom = &tFrom
	}
	if tUntil == 0 {
		promo.ActiveUntil = nil
	} else {
		promo.ActiveUntil = &tUntil
	}

	edited, err := h.service.EditPromo(c.Request().Context(), &promo)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrNoPermission {
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{
				"status":  "error",
				"message": "Промокод не принадлежит этой компании.",
			})
		}
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Промокод не найден.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, edited)
}
func (h *Handlers) BussinessStatPromo(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	var req models.GetPromoStatRequest
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	req.CompanyID = &user.ID
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	stat, err := h.service.GetPromoStat(c.Request().Context(), req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrNoPermission {
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{
				"status":  "error",
				"message": "Промокод не принадлежит этой компании.",
			})
		}
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Промокод не найден.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, stat)
}

func (h *Handlers) UserSignUp(c echo.Context) error {
	var req models.SignUpUserRequest
	if c.Request().Header.Get("Content-Type") != "application/json" {
		h.Error(c.Request().Context(), "content-type now allowed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	hashedPassword, err := utils.Encrypt([]byte(*req.Password), h.CryptoKey)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	id := uuid.NewString()
	user := models.User{
		ID:        &id,
		Name:      req.Name,
		SurName:   req.SurName,
		Email:     req.Email,
		AvatarUrl: req.AvatarUrl,
		Other:     req.Other,
		Password:  hashedPassword,
	}
	err = h.service.UserSignUp(c.Request().Context(), user)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrEmailRegistrated {
			return echo.NewHTTPError(409, echo.Map{
				"status":  "error",
				"message": "Такой email уже зарегистрирован.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	token, err := utils.CreateToken(*user.ID, "user", h.SigningKey)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	err = h.service.UpdateToken(c.Request().Context(), *user.ID, token)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, models.SignUpUserResp{
		Token: token,
	})
}
func (h *Handlers) UserSignIn(c echo.Context) error {
	var req models.SignInUserRequest
	if c.Request().Header.Get("Content-Type") != "application/json" {
		h.Error(c.Request().Context(), "content-type now allowed")
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := c.Bind(&req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	err := h.validate.Struct(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	company := models.User{
		Email: req.Email,
	}
	usr, err := h.service.UserSignIn(c.Request().Context(), company)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
			"status":  "error",
			"message": "Неверный email или пароль.",
		})
	}
	decryptedPassword, err := utils.Decrypt(usr.Password, h.CryptoKey)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
			"status":  "error",
			"message": "Неверный email или пароль.",
		})
	}
	if !bytes.Equal(decryptedPassword, []byte(*req.Password)) {
		h.Error(c.Request().Context(), "password not match", zap.Binary("decrypted", decryptedPassword), zap.Binary("body password", []byte(*req.Password)))
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
			"status":  "error",
			"message": "Неверный email или пароль.",
		})
	}
	token, err := utils.CreateToken(*usr.ID, "user", h.SigningKey)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	err = h.service.UpdateToken(c.Request().Context(), *usr.ID, token)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, models.CompanySignInResponse{
		Token: token,
	})
}
func (h *Handlers) FeedUser(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	baseSort := models.UserSort{
		Id:       user.ID,
		Limit:    10,
		Offset:   0,
		Category: nil,
		Active:   nil,
	}
	var req models.FeedUserRequest
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if req.Limit != nil {
		baseSort.Limit = *req.Limit
	}
	if req.Offset != nil {
		baseSort.Offset = *req.Offset
	}
	if req.Category != nil {
		baseSort.Category = req.Category
	}
	if req.Active != nil {
		baseSort.Active = req.Active
	}
	users, total, err := h.service.FeedUser(c.Request().Context(), &baseSort)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	totalCount := fmt.Sprintf("%d", total)
	c.Response().Header().Add("X-Total-Count", totalCount)

	return c.JSON(200, users)
}
func (h *Handlers) GetUser(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	usr, err := h.service.GetUser(c.Request().Context(), models.User{ID: &user.ID})
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
			"status":  "error",
			"message": "Пользователь не авторизован.",
		})
	}
	resp := models.GetUserResponse{
		Name:    usr.Name,
		SurName: usr.SurName,
		Email:   usr.Email,
		Other:   usr.Other,
	}
	if usr.AvatarUrl != nil {
		resp.AvatarUrl = usr.AvatarUrl
	}
	return c.JSON(200, resp)
}
func (h *Handlers) UpdateUser(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	var req models.EditUserRequest
	req.ID = &user.ID
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}

	usr := models.User{
		ID:        req.ID,
		Name:      req.Name,
		SurName:   req.SurName,
		AvatarUrl: req.AvatarUrl,
		Password:  nil,
	}
	if req.Password != nil {

		hashedPassword, err := utils.Encrypt([]byte(*req.Password), h.CryptoKey)
		if err != nil {
			h.Error(c.Request().Context(), "", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "Ошибка в данных запроса.",
			})
		}
		usr.Password = hashedPassword
	}
	edited, err := h.service.UpdateUser(c.Request().Context(), &usr)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, edited)
}
func (h *Handlers) UserGetPromo(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)

	var req models.UserPromoRequest
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	req.ID = &user.ID
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	users, err := h.service.UserGetPromo(c.Request().Context(), req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Промокод не найден.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}

	return c.JSON(200, users)
}
func (h *Handlers) UserLikePromo(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)

	var req models.UserLikedPromo
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	req.UserID = &user.ID

	err = h.service.UserLikePromo(c.Request().Context(), req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Промокод не найден.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, echo.Map{"status": "ok"})
}
func (h *Handlers) UserDeleteLike(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)

	var req models.UserLikedPromo
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	req.UserID = &user.ID
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}

	err = h.service.UserDeleteLike(c.Request().Context(), req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Промокод не найден.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, echo.Map{"status": "ok"})
}
func (h *Handlers) UserCreateComment(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)

	var req models.UserCommentCreateRequest
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	req.UserID = &user.ID
	now := time.Now().UTC().Add(3 * time.Hour).Format(time.RFC3339)
	req.Date = &now
	id := uuid.NewString()
	req.CommentId = &id
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}

	comment, err := h.service.UserCreateComment(c.Request().Context(), req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Промокод не найден.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(201, comment)
}
func (h *Handlers) UserGetComments(c echo.Context) error {
	_ = c.Get("user").(*utils.JWTClaims)
	baseSort := models.CommentSort{
		Limit:  10,
		Offset: 0,
	}
	var req models.UserGetCommentsRequest
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	baseSort.PromoId = *req.PromoID
	if req.Limit != nil {
		baseSort.Limit = *req.Limit
	}
	if req.Offset != nil {
		baseSort.Offset = *req.Offset
	}
	comments, total, err := h.service.UserGetComments(c.Request().Context(), &baseSort)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if total == 0 {
		h.Error(c.Request().Context(), "", zap.Int("total count", total))
		return echo.NewHTTPError(http.StatusNotFound, echo.Map{
			"status":  "error",
			"message": "Промокод не найден.",
		})
	}
	totalCount := fmt.Sprintf("%d", total)
	c.Response().Header().Add("X-Total-Count", totalCount)

	return c.JSON(200, comments)
}
func (h *Handlers) UserGetComment(c echo.Context) error {
	var req models.UserGetComment
	if err := c.Bind(&req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	comment, err := h.service.UserGetComment(c.Request().Context(), req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Такого промокода или комментария не существует.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})

	}
	return c.JSON(200, comment)
}
func (h *Handlers) UserEditComment(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	var req models.UserEditCommentRequest
	if err := c.Bind(&req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	req.UserID = &user.ID
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	comment, err := h.service.UserEditComment(c.Request().Context(), req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Такого промокода или комментария не существует.",
			})
		}
		if err == service.ErrNoPermission {
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{
				"status":  "error",
				"message": "Комментарий не принадлежит пользователю.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})

	}
	return c.JSON(200, comment)
}
func (h *Handlers) UserDeleteComment(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	var req models.UserDeleteCommentRequest
	if err := c.Bind(&req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	req.UserID = &user.ID
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}

	err := h.service.UserDeleteComment(c.Request().Context(), req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Такого промокода или комментария не существует.",
			})
		}
		if err == service.ErrNoPermission {
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{
				"status":  "error",
				"message": "Комментарий не принадлежит пользователю.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})

	}
	return c.JSON(200, echo.Map{"status": "ok"})
}
func (h *Handlers) UserActivate(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	var req models.ActivateRequest
	if err := c.Bind(&req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	req.UserID = &user.ID
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	value, err := h.service.CheckCache(c.Request().Context(), user.ID)
	if err != nil && err != redis.Nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	ok := false
	if err == redis.Nil {
		var fraudResp models.AntifraudResponse
		usr, err := h.service.GetUser(c.Request().Context(), models.User{ID: &user.ID})
		if err != nil {
			h.Error(c.Request().Context(), "", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "Ошибка в данных запроса.",
			})
		}
		body := models.AntifraudRequest{
			UserEmail: *usr.Email,
			PromoId:   *req.PromoID,
		}
		for range 2 {
			data, err := json.Marshal(body)
			if err != nil {
				h.Error(c.Request().Context(), "", zap.Error(err))
				return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
					"status":  "error",
					"message": "Ошибка в данных запроса.",
				})
			}
			buff := bytes.NewBuffer(data)
			fraudReq, err := http.NewRequest("POST", "http://"+h.AntifraudAddress+"/api/validate", buff)
			if err != nil {
				h.Error(c.Request().Context(), "", zap.Error(err))
				return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
					"status":  "error",
					"message": "Ошибка в данных запроса.",
				})
			}
			fraudReq.Header.Set("Content-Type", "application/json")
			client := http.Client{}
			var resp *http.Response
			resp, err = client.Do(fraudReq)
			if err != nil {
				h.Error(c.Request().Context(), "", zap.Error(err))
				return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
					"status":  "error",
					"message": "Ошибка в данных запроса.",
				})
			}
			if resp.StatusCode == 200 {
				ok = true
				respData, err := io.ReadAll(resp.Body)
				if err != nil {
					h.Error(c.Request().Context(), "", zap.Error(err))
					return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
						"status":  "error",
						"message": "Ошибка в данных запроса.",
					})
				}
				defer resp.Body.Close()

				err = json.Unmarshal(respData, &fraudResp)
				if err != nil {
					h.Error(c.Request().Context(), "", zap.Error(err))
					return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
						"status":  "error",
						"message": "Ошибка в данных запроса.",
					})
				}
				value = fraudResp.Ok
				if fraudResp.CacheUntil != "" {
					err = h.service.Cache(c.Request().Context(), user.ID, fraudResp.CacheUntil, fraudResp.Ok)
					if err != nil {
						h.Error(c.Request().Context(), "", zap.Error(err))
						return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
							"status":  "error",
							"message": "Ошибка в данных запроса.",
						})
					}
				}
				break
			}
		}
		if !ok {
			h.Error(c.Request().Context(), "antifraud не ответил кодом 200")
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{
				"status":  "error",
				"message": "Вы не можете использовать этот промокод.",
			})
		}
	}
	if !value {
		h.Error(c.Request().Context(), "antifraud отказал в доступе")
		return echo.NewHTTPError(http.StatusForbidden, echo.Map{
			"status":  "error",
			"message": "Вы не можете использовать этот промокод.",
		})
	}
	promo, promoErr := h.service.UserActivatePromo(c.Request().Context(), req)
	if promoErr != nil {
		h.Error(c.Request().Context(), "", zap.Error(promoErr))
		if err == service.ErrPromoNotFound {
			return echo.NewHTTPError(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "Промокод не найден.",
			})
		}
		if promoErr == service.ErrNoPermission {
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{
				"status":  "error",
				"message": "Вы не можете использовать этот промокод.",
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, echo.Map{"promo": promo})
}
func (h *Handlers) UserHistory(c echo.Context) error {
	user := c.Get("user").(*utils.JWTClaims)
	baseSort := models.HistorySort{
		UserID: user.ID,
		Limit:  10,
		Offset: 0,
	}
	var req models.UserHistoryRequest
	err := c.Bind(&req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	req.UserID = &user.ID
	if err := h.validate.Struct(req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}

	if req.Limit != nil {
		baseSort.Limit = *req.Limit
	}
	if req.Offset != nil {
		baseSort.Offset = *req.Offset
	}
	promos, total, err := h.service.GetUserHistory(c.Request().Context(), &baseSort)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	totalCount := fmt.Sprintf("%d", total)
	c.Response().Header().Add("X-Total-Count", totalCount)

	return c.JSON(200, promos)
}
func (h *Handlers) UpdateuserVerdict(c echo.Context) error {
	var req models.SetUserVerdict
	if err := c.Bind(&req); err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	data, err := json.Marshal(req)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	buff := bytes.NewBuffer(data)
	fraudReq, err := http.NewRequest("POST", "http://"+h.AntifraudAddress+"/internal/update_user_verdict", buff)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	fraudReq.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	var resp *http.Response
	resp, err = client.Do(fraudReq)
	if err != nil {
		h.Error(c.Request().Context(), "", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	if resp.StatusCode != 200 {
		h.Error(c.Request().Context(), "antifraud update verdict not 200 status code")
		return echo.NewHTTPError(resp.StatusCode, echo.Map{
			"status":  "error",
			"message": "Ошибка в данных запроса.",
		})
	}
	return c.JSON(200, echo.Map{"status": "ok"})
}
