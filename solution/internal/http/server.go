package http

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Handlers interface {
	Ping(c echo.Context) error
	RequestTimeMiddleware(echo echo.HandlerFunc) echo.HandlerFunc
	BusinessSignUp(c echo.Context) error
	BusinessSignIn(c echo.Context) error
	BussinessAuthJWT(echo echo.HandlerFunc) echo.HandlerFunc
	BussinessCreatePromo(c echo.Context) error
	BussinessGetPromos(c echo.Context) error
	BussinessGetPromo(c echo.Context) error
	BussinessEditPromo(c echo.Context) error
	BussinessStatPromo(c echo.Context) error
	UserAuthJWT(echo.HandlerFunc) echo.HandlerFunc
	UserSignUp(c echo.Context) error
	UserSignIn(c echo.Context) error
	GetUser(c echo.Context) error
	UpdateUser(c echo.Context) error
	FeedUser(c echo.Context) error
	UserGetPromo(c echo.Context) error
	UserLikePromo(c echo.Context) error
	UserDeleteLike(c echo.Context) error
	UserCreateComment(c echo.Context) error
	UserGetComments(c echo.Context) error
	UserGetComment(c echo.Context) error
	UserEditComment(c echo.Context) error
	UserDeleteComment(c echo.Context) error
	UserActivate(c echo.Context) error
	UserHistory(c echo.Context) error
	UpdateuserVerdict(c echo.Context) error
}
type Server struct {
	server  *echo.Echo
	address string
}

func New(ctx context.Context, srv Handlers, SigningKey string, address string) (*Server, error) {
	e := echo.New()
	e.Use(srv.RequestTimeMiddleware)
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(
		middleware.LoggerConfig{
			Format:           `{"time":"${time_rfc3339}, "host":"${host}", "method":"${method}", "uri":"${uri}", "status":${status}, "error":"${error}}` + "\n",
			CustomTimeFormat: "2006-01-02 15:04:05",
		},
	))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowMethods: []string{"POST", "OPTIONS", "GET", "DELETE", "PATCH"},
	}))

	e.GET("/api/ping", srv.Ping)
	e.POST("/internal/update_user_verdict",srv.UpdateuserVerdict)
	e.POST("/api/business/auth/sign-up", srv.BusinessSignUp)
	e.POST("/api/business/auth/sign-in", srv.BusinessSignIn)

	e.POST("/api/business/promo", srv.BussinessCreatePromo, srv.BussinessAuthJWT) //TODO
	e.GET("/api/business/promo", srv.BussinessGetPromos, srv.BussinessAuthJWT)
	e.GET("/api/business/promo/:id", srv.BussinessGetPromo, srv.BussinessAuthJWT)
	e.PATCH("/api/business/promo/:id", srv.BussinessEditPromo, srv.BussinessAuthJWT)
	e.GET("/api/business/promo/:id/stat", srv.BussinessStatPromo, srv.BussinessAuthJWT)

	e.POST("/api/user/auth/sign-up", srv.UserSignUp)
	e.POST("/api/user/auth/sign-in", srv.UserSignIn)
	e.GET("/api/user/profile", srv.GetUser, srv.UserAuthJWT)
	e.PATCH("/api/user/profile", srv.UpdateUser, srv.UserAuthJWT)
	e.GET("/api/user/feed", srv.FeedUser, srv.UserAuthJWT)
	e.GET("/api/user/promo/:id", srv.UserGetPromo, srv.UserAuthJWT)
	e.POST("/api/user/promo/:id/like", srv.UserLikePromo, srv.UserAuthJWT)
	e.DELETE("/api/user/promo/:id/like", srv.UserDeleteLike, srv.UserAuthJWT)
	e.POST("/api/user/promo/:id/comments", srv.UserCreateComment, srv.UserAuthJWT)
	e.GET("/api/user/promo/:id/comments", srv.UserGetComments, srv.UserAuthJWT)
	e.GET("/api/user/promo/:id/comments/:comment_id", srv.UserGetComment, srv.UserAuthJWT)
	e.PUT("/api/user/promo/:id/comments/:comment_id", srv.UserEditComment, srv.UserAuthJWT)
	e.DELETE("/api/user/promo/:id/comments/:comment_id", srv.UserDeleteComment, srv.UserAuthJWT)
	e.POST("/api/user/promo/:id/activate", srv.UserActivate, srv.UserAuthJWT)
	e.GET("/api/user/promo/history", srv.UserHistory, srv.UserAuthJWT)
	server := &Server{e, address}
	return server, nil
}

func (s *Server) Start(ctx context.Context) error {
	return s.server.Start(s.address)
}
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
