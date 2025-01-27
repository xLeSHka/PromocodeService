package main

import (
	"context"
	"os"
	"os/signal"
	"reflect"
	"solution/internal/config"
	"solution/internal/http"
	"solution/internal/http/handlers"
	postgresrepository "solution/internal/repository/postgresRepository"
	redisrepository "solution/internal/repository/redisRepository"
	"solution/internal/service"
	"solution/internal/utils"
	"solution/pkg/db/cache"
	"solution/pkg/db/postgres"
	"solution/pkg/logger"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

const SigningKey = "my-secret-key"

var CryptoKey = []byte("12345678901234567890123456789012")

func main() {
	ctx := context.Background()
	mainLogger := logger.New()
	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)

	utils.Validate.RegisterValidation("password", utils.PasswordValidationFunc)

	utils.Validate.RegisterValidation("country_validation", utils.CountryValidationFunc)
	utils.Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	utils.Validate.RegisterValidation("date_validation", utils.DateValidationFunc)

	cfg, err := config.Read()
	if err != nil {
		mainLogger.Fatal(ctx, "failed read env", zap.Error(err))
	}
	db, err := postgres.New(cfg.PostgresConfig)

	if err != nil {
		mainLogger.Fatal(ctx, "failed read env", zap.Error(err))
	}
	db.Db.Exec(`CREATE TABLE if not exists companies
	(
		company_id uuid NOT NULL,
		name character varying(50) NOT NULL,
		email character varying(120) NOT NULL,
		password bytea NOT NULL,
		PRIMARY KEY (company_id, email)
	);`)
	db.Db.Exec(`CREATE TABLE if not exists users
	(
		id uuid NOT NULL,
		name character varying(120) NOT NULL,
		surname character varying(140) NOT NULL,
		email character varying(120) NOT NULL,
		avatar_url text,
		other jsonb NOT NULL,
		password bytea NOT NULL,
		PRIMARY KEY (id, email)
	);`)
	db.Db.Exec(`CREATE TABLE if not exists promos
	(
		id serial NOT NULL,
		description text NOT NULL,
		image_url text,
		target jsonb,
		max_count integer NOT NULL,
		active_from bigint,
		active_until bigint,
		mode character varying(16) NOT NULL,
		promo_common character varying(64),
		promo_unique character varying(64)[],
		used_promo_unique character varying(64)[],
		promo_id uuid NOT NULL,
		company_id uuid NOT NULL,
		company_name character varying(64) NOT NULL,
		like_count integer NOT NULL,
		used_count integer NOT NULL,
		comment_count int NOT NULL,
		active boolean NOT NULL,
		PRIMARY KEY (promo_id)
	);`)
	db.Db.Exec(`CREATE TABLE if not exists activations
	(
		seq_id serial NOT NULL,
		activate_time bigint,
		country character varying(4) NOT NULL,
		promo_id uuid NOT NULL,
		id uuid NOT NULL,
		PRIMARY KEY(seq_id)
	);`)
	db.Db.Exec(`CREATE TABLE if not exists promosstat
	(
		promo_id uuid NOT NULL,
		id uuid NOT NULL,
		is_liked_by_user boolean NOT NULL,
		PRIMARY KEY (id)
	);`)
	db.Db.Exec(`CREATE TABLE if not exists comments
	(

		serial_number serial NOT NULL,
		id uuid NOT NULL,
		user_id uuid NOT NULL,
		promo_id uuid NOT NULL,
		text text NOT NULL,
		date varchar(50) NOT NULL,
		author jsonb NOT NULL,
		PRIMARY KEY(serial_number,id)
	);`)
	postgresRepo := postgresrepository.New(db)
	client := cache.New(cfg.RedisConfig)

	redsiRepo := redisrepository.New(client)

	srv := service.New(redsiRepo, postgresRepo)

	handelrs := handlers.New(srv, SigningKey, cfg.AntifraudAddress, CryptoKey, utils.Validate, mainLogger)
	server, err := http.New(ctx, handelrs, SigningKey, cfg.ServerAddress)

	if err != nil {
		mainLogger.Error(ctx, "", zap.Error(err))
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := server.Start(ctx); err != nil {
			mainLogger.Error(ctx, "failed start server", zap.Error(err))
		}

	}()
	mainLogger.Error(ctx, "start server", zap.String("address", cfg.ServerAddress))
	<-graceCh

	if err := server.Stop(ctx); err != nil {
		mainLogger.Error(ctx, "failed graceful shutdown server", zap.Error(err))
	}

	mainLogger.Info(ctx, "server stopped")
	// hashedPassword, err := utils.HashPassword("123456789012345678901234567890123456789012345678901234567890")
	// fmt.Println(hashedPassword, err, 2)
	// cmp := &models.Company{CompanyID: "ff010a23-04a8-4456-b21f-69cbb61a867d", CompanyName: "Nikita company", Email: "123@mail.com", Password: hashedPassword}
	// res2, err2 := sq.Insert("companies").Columns("company_id", "company_name", "email", "password").RunWith(db.Db).PlaceholderFormat(sq.Dollar).Values(cmp.CompanyID, cmp.CompanyName, cmp.Email, hashedPassword).Exec()
	// fmt.Println(res2, err2, 3)
	// cmp2 := &models.Company{CompanyID: "ff010a23-04a8-4456-b21f-69cbb61a867d"}
	// err3 := sq.Select("company_name", "email", "password").From("companies").Where(sq.Eq{"company_id": cmp2.CompanyID}).PlaceholderFormat(sq.Dollar).RunWith(db.Db).Scan(&cmp2.CompanyName, &cmp2.Email, &cmp2.Password)

	// fmt.Println(cmp2, err3, 4)
	// redisClient.Set(ctx, "key", 123, 24*time.Hour)
	// fmt.Println(redisClient.Get(ctx, "key"))

}
