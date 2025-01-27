package postgresrepository

import (
	"context"
	"database/sql"
	"fmt"
	"solution/internal/models"
	"solution/internal/service"
	"solution/pkg/db/postgres"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type PostgresRepo struct {
	db *postgres.DB
}

func New(db *postgres.DB) *PostgresRepo {
	return &PostgresRepo{db}
}
func (pr *PostgresRepo) TestCompanyRegistration(ctx context.Context, company models.Company) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM companies WHERE email = $1)`
	var exist bool
	err := pr.db.Db.QueryRow(q, company.Email).Scan(&exist)
	if err != nil {
		return false, err
	}
	return exist, nil
}
func (pr *PostgresRepo) GetCompanyByEmail(ctx context.Context, company models.Company) (*models.Company, error) {
	var res models.Company
	err := sq.Select("company_id", "name", "password").
		From("companies").
		Where(sq.Eq{"email": company.Email}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		QueryRow().
		Scan(&res.CompanyID, &res.Name, &res.Password)
	if err != nil {
		return nil, err
	}
	res.Email = company.Email
	return &res, nil
}
func (pr *PostgresRepo) GetCompanyById(ctx context.Context, company models.Company) (*models.Company, error) {
	var res models.Company
	err := sq.Select("email", "name", "password").
		From("companies").
		Where(sq.Eq{"company_id": company.CompanyID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		QueryRow().
		Scan(&res.Email, &res.Name, &res.Password)
	if err != nil {
		return nil, err
	}
	res.CompanyID = company.CompanyID
	return &res, nil
}
func (pr *PostgresRepo) AddCompany(ctx context.Context, company models.Company) error {
	_, err := sq.Insert("companies").
		Columns("company_id", "email", "name", "password").
		Values(company.CompanyID, company.Email, company.Name, company.Password).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return err
	}
	return nil
}
func (pr *PostgresRepo) CreatePromo(ctx context.Context, promo *models.Promo) error {
	_, err := sq.Insert("promos").
		Columns("description,image_url,target,max_count,active_from,active_until,mode,promo_common,promo_unique,used_promo_unique,promo_id,company_id,company_name,like_count,used_count,comment_count,active").
		Values(promo.Description, promo.ImageUrl, promo.Target, promo.MaxCount,
			promo.ActiveFrom, promo.ActiveUntil, promo.Mode, promo.PromoCommon, promo.PromoUnique, models.StringSlice{},
			promo.PromoId, promo.CompanyId, promo.CompanyName, promo.LikeCount,
			promo.UsedCount, 0, promo.Active).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return err
	}
	return nil
}
func (pr *PostgresRepo) GetPromos(ctx context.Context, sortRules *models.CompanySort) ([]models.GetPromoResponse, int, error) {
	promos := make([]models.GetPromoResponse, 0)
	selectBuilder := sq.Select("description,image_url,target,max_count,active_from,active_until,mode,promo_common,promo_unique,promo_id,company_id,company_name,like_count,used_count,active").
		From("promos").
		Where(sq.Eq{"company_id": sortRules.CompanyId}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db)
	if sortRules.SortBy != "" {
		selectBuilder = selectBuilder.OrderBy(fmt.Sprintf("%s DESC", sortRules.SortBy))
	}
	selectBuilder = selectBuilder.OrderBy("id DESC")
	var rows *sql.Rows
	var err error
	if sortRules.Countries != nil {
		rows, err = selectBuilder.Where(sq.Or{sq.Eq{"lower(target ->> 'country')": sortRules.Countries},
			sq.Eq{"target ->> 'country'": nil}}).Query()
	} else {
		rows, err = selectBuilder.Query()
	}
	if err != nil {
		return nil, 0, err
	}
	var count int = 0
	for rows.Next() {
		var promo models.GetPromoResponse
		var ActiveFrom, ActiveUntil *int64
		err := rows.Scan(&promo.Description, &promo.ImageUrl, &promo.Target, &promo.MaxCount, &ActiveFrom, &ActiveUntil, &promo.Mode, &promo.PromoCommon, &promo.PromoUnique, &promo.PromoId, &promo.CompanyId, &promo.CompanyName, &promo.LikeCount, &promo.UsedCount, &promo.Active) //
		if err != nil {
			return nil, 0, err
		}
		if sortRules.Offset <= count && len(promos) < sortRules.Limit {

			if ActiveFrom != nil {
				t := time.Unix(*ActiveFrom, 0).Format("2006-01-02")
				promo.ActiveFrom = &t
			}
			if ActiveUntil != nil {
				t := time.Unix(*ActiveUntil, 0).Format("2006-01-02")
				promo.ActiveUntil = &t
			}

			promos = append(promos, promo)
		}
		count++

	}
	return promos, count, nil
}
func (pr *PostgresRepo) GetPromo(ctx context.Context, promo models.Promo) (*models.GetPromoResponse, error) {
	var resp models.GetPromoResponse
	var ActiveFrom, ActiveUntil *int64
	err := sq.Select("description,image_url,target,max_count,active_from,active_until,mode,promo_common,promo_unique,promo_id,company_id,company_name,like_count,used_count,active").
		From("promos").
		Where(sq.And{sq.Eq{"promo_id": promo.PromoId}, sq.Eq{"company_id": promo.CompanyId}}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Scan(&resp.Description, &resp.ImageUrl, &resp.Target, &resp.MaxCount, &ActiveFrom, &ActiveUntil, &resp.Mode, &resp.PromoCommon, &resp.PromoUnique, &resp.PromoId, &resp.CompanyId, &resp.CompanyName, &resp.LikeCount, &resp.UsedCount, &resp.Active)
	if err != nil {
		return nil, err
	}
	if ActiveFrom != nil {
		t := time.Unix(*ActiveFrom, 0).Format("2006-01-02")
		resp.ActiveFrom = &t
	}
	if ActiveUntil != nil {
		t := time.Unix(*ActiveUntil, 0).Format("2006-01-02")
		resp.ActiveUntil = &t
	}
	return &resp, nil
}
func (pr *PostgresRepo) GetPromoById(ctx context.Context, promo models.Promo) (*models.Promo, error) {
	var resp models.Promo
	var ActiveFrom, ActiveUntil *int64
	err := sq.Select("description,image_url,target,max_count,active_from,active_until,mode,promo_common,promo_unique,used_promo_unique,promo_id,company_id,company_name,like_count,used_count,comment_count,active").
		From("promos").
		Where(sq.Eq{"promo_id": promo.PromoId}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Scan(&resp.Description, &resp.ImageUrl, &resp.Target, &resp.MaxCount, &ActiveFrom, &ActiveUntil, &resp.Mode, &resp.PromoCommon, &resp.PromoUnique, &promo.UsedPromoUnique, &resp.PromoId, &resp.CompanyId, &resp.CompanyName, &resp.LikeCount, &resp.UsedCount, &resp.CommentCount, &resp.Active)
	if err != nil {
		return nil, err
	}
	if ActiveFrom != nil {
		t := time.Unix(*ActiveFrom, 0).Unix()
		resp.ActiveFrom = &t
	}
	if ActiveUntil != nil {
		t := time.Unix(*ActiveUntil, 0).Unix()
		resp.ActiveUntil = &t
	}
	return &resp, nil
}
func (pr *PostgresRepo) EditPromo(ctx context.Context, promo *models.Promo) (*models.GetPromoResponse, error) {
	var resp models.GetPromoResponse
	sets := make(map[string]interface{})
	if promo.Description != nil {
		sets["description"] = promo.Description
	}
	if promo.ImageUrl != nil {
		sets["image_url"] = promo.ImageUrl
	}
	if promo.Target != nil {
		sets["target"] = promo.Target
	}
	if promo.MaxCount != nil {

		sets["max_count"] = promo.MaxCount
	}
	if promo.ActiveFrom != nil {
		sets["active_from"] = promo.ActiveFrom
	}
	if promo.ActiveUntil != nil {
		sets["active_until"] = promo.ActiveUntil
	}

	updateBuileder := sq.Update("promos").
		Where(sq.Eq{"promo_id": *promo.PromoId}).Set("active", promo.Active).
		SetMap(sets).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Suffix("RETURNING description,image_url,target,max_count,active_from,active_until,mode,promo_common,promo_unique,promo_id,company_id,company_name,like_count,used_count,active")

	var ActiveFrom, ActiveUntil *int64
	err := updateBuileder.QueryRow().Scan(&resp.Description, &resp.ImageUrl, &resp.Target, &resp.MaxCount, &ActiveFrom, &ActiveUntil, &resp.Mode, &resp.PromoCommon, &resp.PromoUnique, &resp.PromoId, &resp.CompanyId, &resp.CompanyName, &resp.LikeCount, &resp.UsedCount, &resp.Active)
	if err != nil {
		return nil, err
	}
	if ActiveFrom != nil {
		t := time.Unix(*ActiveFrom, 0).Format("2006-01-02")
		resp.ActiveFrom = &t
	}
	if ActiveUntil != nil {
		t := time.Unix(*ActiveUntil, 0).Format("2006-01-02")
		resp.ActiveUntil = &t
	}
	return &resp, nil
}
func (pr *PostgresRepo) GetPromoStat(ctx context.Context, promo models.GetPromoStatRequest) (*models.GetPromoStatResponse, error) {
	var resp models.GetPromoStatResponse
	rows, err := sq.Select("country").
		From("activations").
		Where(sq.Eq{"promo_id": promo.PromoID}).
		PlaceholderFormat(sq.Dollar).
		OrderBy("activate_time DESC").
		RunWith(pr.db.Db).Query()
	if err != nil {
		return nil, err
	}
	uses := make(map[string]int)
	for rows.Next() {
		var country string
		err = rows.Scan(&country)
		if err != nil {
			return nil, err
		}
		uses[country]++
	}
	var activations_count int
	for k, v := range uses {
		var country models.Country
		country.Country = k
		country.ActivationsCount = v
		activations_count += v
		resp.Countries = append(resp.Countries, country)
	}
	resp.ActivationsCount = activations_count
	return &resp, nil
}
func (pr *PostgresRepo) TestUserRegistration(ctx context.Context, user models.User) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exist bool
	err := pr.db.Db.QueryRow(q, user.Email).Scan(&exist)
	if err != nil {
		return false, err
	}
	return exist, nil
}
func (pr *PostgresRepo) AddUser(ctx context.Context, user models.User) error {
	_, err := sq.Insert("users").
		Columns("id", "name", "surname", "email", "avatar_url", "other", "password").
		Values(user.ID, user.Name, user.SurName, user.Email, user.AvatarUrl, user.Other, user.Password).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return err
	}
	return nil
}
func (pr *PostgresRepo) GetUserByEmail(ctx context.Context, User models.User) (*models.User, error) {
	var res models.User
	err := sq.Select("id", "name", "surname", "avatar_url", "other", "password").
		From("users").
		Where(sq.Eq{"email": User.Email}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		QueryRow().
		Scan(&res.ID, &res.Name, &res.SurName, &res.AvatarUrl, &res.Other, &res.Password)
	if err != nil {
		return nil, err
	}
	res.Email = User.Email
	return &res, nil
}
func (pr *PostgresRepo) GetUserById(ctx context.Context, User models.User) (*models.User, error) {
	var res models.User
	err := sq.Select("email", "name", "surname", "avatar_url", "other", "password").
		From("users").
		Where(sq.Eq{"id": User.ID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		QueryRow().
		Scan(&res.Email, &res.Name, &res.SurName, &res.AvatarUrl, &res.Other, &res.Password)
	if err != nil {
		return nil, err
	}
	res.ID = User.ID
	return &res, nil
}
func (pr *PostgresRepo) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	var resp models.User
	sets := make(map[string]interface{})
	if user.Name != nil {
		sets["name"] = user.Name
	}
	if user.SurName != nil {
		sets["surname"] = user.SurName
	}
	if user.AvatarUrl != nil {
		sets["avatar_url"] = user.AvatarUrl
	}
	if user.Password != nil {
		sets["password"] = user.Password
	}
	updateBuileder := sq.Update("users").
		Where(sq.Eq{"id": *user.ID}).
		SetMap(sets).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Suffix("RETURNING name,surname,email,avatar_url,other,password")

	err := updateBuileder.QueryRow().Scan(&resp.Name, &resp.SurName, &resp.Email, &resp.AvatarUrl, &resp.Other, &resp.Password)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
func (pr *PostgresRepo) FeedUser(ctx context.Context, sortRules *models.UserSort) ([]models.FeedUserResponse, int, error) {
	promos := make([]models.FeedUserResponse, 0)
	target := models.Target{}
	q := `SELECT description, image_url, promo_id, company_id, active_from, 
	active_until, mode, promo_unique, used_promo_unique, max_count, used_count, 
	company_name, like_count, comment_count, active, target FROM promos
	WHERE (lower(target ->> 'country') = $1 OR target ->> 'country' IS NULL) 
	AND ((target ->> 'age_from' <= $2 OR target ->> 'age_from' IS NULL) 
	AND (target ->> 'age_until' >= $3 OR target ->> 'age_until' IS NULL))`
	t := 4
	args := make([]interface{}, 0)
	country := strings.ToLower(sortRules.Country)
	args = append(args, &country, sortRules.Age, sortRules.Age)
	if sortRules.Active != nil {
		q += fmt.Sprintf("AND (active = $%d)", t)
		args = append(args, sortRules.Active)
		t++
	}
	if sortRules.Category != nil {
		q += fmt.Sprintf("AND ((lower(target->>'categories'))::jsonb ? $%d)", t)
		args = append(args, strings.ToLower(*sortRules.Category))
		t++
	}
	q += ` ORDER BY id DESC`
	rows, err := pr.db.Db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	var count int = 0
	for rows.Next() {
		if sortRules.Offset <= count && len(promos) < sortRules.Limit {
			var ActiveFrom, ActiveUntil *int64
			var promo models.FeedUserResponse
			var mode string
			var uniq, usedUniq models.StringSlice
			var maxCount, usedCount int
			scanErr := rows.Scan(&promo.Description, &promo.ImageUrl, &promo.PromoId, &promo.CompanyId, &ActiveFrom, &ActiveUntil, &mode, &uniq, &usedUniq, &maxCount, &usedCount, &promo.CompanyName, &promo.LikeCount, &promo.CommentCount, &promo.Active, &target) //
			if scanErr != nil {
				return nil, 0, scanErr
			}
			count, until, from := *promo.Active, *promo.Active, *promo.Active
			now := time.Now().UTC().Add(3 * time.Hour).Unix()
			if mode == "COMMON" {
				if *promo.Active != (maxCount > usedCount) {
					count = maxCount > usedCount
				}
			} else {
				if *promo.Active != (len(uniq) > len(usedUniq)) {
					count = len(uniq) > len(usedUniq)
				}
			}
			if ActiveFrom != nil {
				if *promo.Active != (*ActiveFrom > now) {
					from = *ActiveFrom <= now
				}

			}
			if ActiveUntil != nil {
				if *promo.Active != (*ActiveUntil < now) {
					until = *ActiveUntil >= now
				}
			}
			if (count && until && from) != *promo.Active {
				_, checkErr := sq.Update("promos").
					Where(sq.Eq{"promo_id": promo.PromoId}).
					Set("active", count && until && from).
					RunWith(pr.db.Db).
					Exec()
				if checkErr != nil {
					return nil, 0, checkErr
				}
				continue
			}
			statErr := sq.Select("is_liked_by_user").
				From("promosstat").
				Where(sq.And{sq.Eq{"id": sortRules.Id}, sq.Eq{"promo_id": *promo.PromoId}}).
				PlaceholderFormat(sq.Dollar).
				RunWith(pr.db.Db).QueryRow().Scan(&promo.IsLiked)
			if statErr != nil && statErr != sql.ErrNoRows {
				return nil, 0, statErr
			}
			falseVal := false
			trueVal := true
			if statErr == sql.ErrNoRows {
				promo.IsLiked = &falseVal
			}
			var activateTime *int64
			activeErr := sq.Select("activate_time").
				From("activations").
				Where(sq.And{sq.Eq{"id": sortRules.Id}, sq.Eq{"promo_id": *promo.PromoId}}).
				PlaceholderFormat(sq.Dollar).
				RunWith(pr.db.Db).QueryRow().Scan(activateTime)
			if activeErr != nil && activeErr != sql.ErrNoRows {
				return nil, 0, activeErr
			}
			if activeErr == sql.ErrNoRows {
				promo.IsActivated = &falseVal
			} else {
				promo.IsActivated = &trueVal
			}
			promos = append(promos, promo)
		}
		count++

	}
	return promos, count, nil
}
func (pr *PostgresRepo) UserGetPromo(ctx context.Context, req models.UserPromoRequest) (*models.FeedUserResponse, error) {
	promo := models.FeedUserResponse{}
	target := models.Target{}
	var ActiveFrom, ActiveUntil *int64
	var mode string
	var uniq, usedUniq models.StringSlice
	var maxCount, usedCount int
	err := sq.Select("description", "image_url", "promo_id", "company_id", "active_from", "active_until", "mode", "promo_unique", "used_promo_unique", "max_count", "used_count", "company_name", "like_count", "comment_count", "active", "target").
		From("promos").
		Where(sq.Eq{"promo_id": req.PromoId}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).Scan(&promo.Description, &promo.ImageUrl, &promo.PromoId, &promo.CompanyId, &ActiveFrom, &ActiveUntil, &mode, &uniq, &usedUniq, &maxCount, &usedCount, &promo.CompanyName, &promo.LikeCount, &promo.CommentCount, &promo.Active, &target) //
	if err != nil {
		return nil, err
	}

	count, until, from := *promo.Active, *promo.Active, *promo.Active
	now := time.Now().UTC().Add(3 * time.Hour).Unix()
	if mode == "COMMON" {
		if *promo.Active != (maxCount > usedCount) {
			count = maxCount > usedCount
		}
	} else {
		if *promo.Active != (len(uniq) > len(usedUniq)) {
			count = len(uniq) > len(usedUniq)
		}
	}
	if ActiveFrom != nil {
		if *promo.Active != (*ActiveFrom > now) {
			from = *ActiveFrom <= now
		}

	}
	if ActiveUntil != nil {
		if *promo.Active != (*ActiveUntil < now) {
			until = *ActiveUntil >= now
		}
	}
	if (count && until && from) != *promo.Active {
		_, checkErr := sq.Update("promos").
			Where(sq.Eq{"promo_id": promo.PromoId}).
			Set("active", count && until && from).
			RunWith(pr.db.Db).
			Exec()
		if checkErr != nil {
			return nil, checkErr
		}
		*promo.Active = count && until && from
	}
	statErr := sq.Select("is_liked_by_user").
		From("promosstat").
		Where(sq.And{sq.Eq{"id": req.ID}, sq.Eq{"promo_id": *promo.PromoId}}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).QueryRow().Scan(&promo.IsLiked)
	if statErr != nil && statErr != sql.ErrNoRows {
		return nil, statErr
	}
	falseVal := false
	trueVal := true
	if statErr == sql.ErrNoRows {
		promo.IsLiked = &falseVal
	}
	var activateTime int64
	activeErr := sq.Select("activate_time").
		From("activations").
		Where(sq.And{sq.Eq{"id": req.ID}, sq.Eq{"promo_id": *promo.PromoId}}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).QueryRow().Scan(&activateTime)
	if activeErr != nil && activeErr != sql.ErrNoRows {
		return nil, activeErr
	}
	if activeErr == sql.ErrNoRows {
		promo.IsActivated = &falseVal
	} else {
		promo.IsActivated = &trueVal
	}

	return &promo, nil
}
func (pr *PostgresRepo) CheckISLiked(ctx context.Context, promo models.UserPromoRequest) (bool, error) {
	var liked bool
	statErr := sq.Select("is_liked_by_user").
		From("promosstat").
		Where(sq.And{sq.Eq{"id": promo.ID}, sq.Eq{"promo_id": promo.PromoId}}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).QueryRow().Scan(&liked)
	if statErr != nil {
		return false, statErr
	}
	return liked, nil
}
func (pr *PostgresRepo) UserLikePromo(ctx context.Context, promo models.UserLikedPromo) error {
	_, err := sq.Insert("promosstat").
		Columns("promo_id", "id", "is_liked_by_user").
		Values(promo.PromoId, promo.UserID, promo.IsLiked).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return err
	}
	_, err = sq.Update("promos").
		Where(sq.Eq{"promo_id": promo.PromoId}).
		Set("like_count", promo.LikeCount).PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return err
	}
	return nil
}
func (pr *PostgresRepo) UserUpdateLike(ctx context.Context, promo models.UserLikedPromo) error {
	_, err := sq.Update("promosstat").
		Where(sq.And{sq.Eq{"promo_id": promo.PromoId}, sq.Eq{"id": promo.UserID}}).
		Set("is_liked_by_user", promo.IsLiked).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return err
	}
	_, err = sq.Update("promos").
		Where(sq.Eq{"promo_id": promo.PromoId}).
		Set("like_count", promo.LikeCount).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return err
	}
	return nil
}
func (pr *PostgresRepo) UserCreateComment(ctx context.Context, comment models.UserCommentCreateRequest) (*models.Comment, error) {
	var comm models.Comment
	err := sq.Insert("comments").
		Columns("id", "user_id", "promo_id", "text", "date", "author").
		Values(comment.CommentId, comment.UserID, comment.PromoID, comment.Text, comment.Date, comment.Author).
		PlaceholderFormat(sq.Dollar).
		Suffix("RETURNING id, text, date, author").
		RunWith(pr.db.Db).
		Scan(&comm.CommentId, &comm.Text, &comm.Date, &comm.Author)
	if err != nil {
		return nil, err
	}
	_, err = sq.Update("promos").
		Where(sq.Eq{"promo_id": comment.PromoID}).
		Set("comment_count", *comment.CommentCount).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return nil, err
	}
	return &comm, nil
}
func (pr *PostgresRepo) UserGetComments(ctx context.Context, sortRules *models.CommentSort) ([]models.Comment, int, error) {
	comments := make([]models.Comment, 0)
	rows, err := sq.Select("id,text,date,author").
		From("comments").
		Where(sq.Eq{"promo_id": sortRules.PromoId}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		OrderBy("serial_number DESC").
		Query()

	if err != nil {
		return nil, 0, err
	}
	var count int = 0
	for rows.Next() {
		if sortRules.Offset <= count && len(comments) < sortRules.Limit {
			var comment models.Comment
			err := rows.Scan(&comment.CommentId, &comment.Text, &comment.Date, &comment.Author)
			if err != nil {
				return nil, 0, err
			}
			comments = append(comments, comment)
		}
		count++

	}
	return comments, count, nil
}
func (pr *PostgresRepo) UserGetComment(ctx context.Context, comment models.UserGetComment) (*models.Comment, error) {
	var comm models.Comment
	err := sq.Select("id,text,date,author").
		From("comments").
		Where(sq.And{sq.Eq{"promo_id": comment.PromoID}, sq.Eq{"id": comment.CommentId}}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		QueryRow().
		Scan(&comm.CommentId, &comm.Text, &comm.Date, &comm.Author)
	if err != nil {
		return nil, err
	}
	return &comm, nil
}
func (pr *PostgresRepo) CheckComment(ctx context.Context, comment models.UserCheckComments) (bool, error) {
	var user string
	err := sq.Select("user_id").
		From("comments").
		Where(sq.And{sq.Eq{"promo_id": comment.PromoID}, sq.Eq{"id": comment.CommentId}}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		QueryRow().
		Scan(&user)
	if err != nil {
		return false, err
	}
	if user != *comment.UserID {
		return false, nil
	}
	return true, nil
}
func (pr *PostgresRepo) UserEditComment(ctx context.Context, comment models.UserEditCommentRequest) (*models.Comment, error) {
	var comm models.Comment
	err := sq.Update("comments").
		Where(sq.Eq{"user_id": comment.UserID}).
		Where(sq.Eq{"id": comment.CommentId}).
		Where(sq.Eq{"promo_id": comment.PromoID}).
		Set("text", comment.Text).
		PlaceholderFormat(sq.Dollar).
		Suffix("RETURNING id,text,date,author").
		RunWith(pr.db.Db).
		QueryRow().
		Scan(&comm.CommentId, &comm.Text, &comm.Date, &comm.Author)
	if err != nil {
		return nil, err
	}
	return &comm, nil
}
func (pr *PostgresRepo) UserDeleteComment(ctx context.Context, comment models.UserDeleteCommentRequest) error {
	_, err := sq.Delete("comments").
		Where(sq.Eq{"user_id": comment.UserID}).
		Where(sq.Eq{"id": comment.CommentId}).
		Where(sq.Eq{"promo_id": comment.PromoID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return err
	}
	_, err = sq.Update("promos").
		Where(sq.Eq{"promo_id": comment.PromoID}).
		Set("comment_count", comment.CommentCount).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return err
	}
	return err
}
func (pr *PostgresRepo) UserActivatePromo(ctx context.Context, promo models.ActivateRequest) (string, error) {
	var promocode models.Promo

	var ActiveFrom, ActiveUntil *int64
	err := sq.Select("max_count", "active_from", "active_until", "mode", "promo_common", "promo_unique", "used_promo_unique", "used_count", "active").
		From("promos").
		Where(sq.Eq{"promo_id": promo.PromoID}).
		Where(sq.Or{sq.Eq{"lower(target ->> 'country')": promo.Country}, sq.Eq{"target ->> 'country'": nil}}).
		Where(sq.Or{sq.LtOrEq{"target ->> 'age_from'": promo.Age}, sq.Eq{"target ->> 'age_from'": nil}}).
		Where(sq.Or{sq.GtOrEq{"target ->> 'age_until'": promo.Age}, sq.Eq{"target ->> 'age_until'": nil}}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Scan(&promocode.MaxCount, &ActiveFrom, &ActiveUntil, &promocode.Mode, &promocode.PromoCommon, &promocode.PromoUnique, &promocode.UsedPromoUnique, &promocode.UsedCount, &promocode.Active)
	if err != nil {
		return "", err
	}
	var res string
	count, until, from := *promocode.Active, *promocode.Active, *promocode.Active
	if *promocode.Mode == "COMMON" {
		count = *promocode.MaxCount > *promocode.UsedCount
		if count {
			res = *promocode.PromoCommon
			*promocode.UsedCount += 1
		}
	} else {
		count = len(promocode.PromoUnique) > len(promocode.UsedPromoUnique)
		if count {
			used := make(map[string]int)
			if promocode.UsedPromoUnique[0] == "" {
				res = promocode.PromoUnique[0]

				promocode.UsedPromoUnique = append(models.StringSlice{}, promocode.PromoUnique[0])
				*promocode.UsedCount += 1
			} else {
				for _, code := range promocode.UsedPromoUnique {
					used[code] += 1
				}

				for _, code := range promocode.PromoUnique {
					_, ok := used[code]
					if ok {
						used[code] -= 1
						if v := used[code]; v == 0 {
							delete(used, code)
						}
					}
					if !ok {
						res = code
						promocode.UsedPromoUnique = append(promocode.UsedPromoUnique, code)
						*promocode.UsedCount += 1
						break
					}
				}
			}
		}
	}
	now := time.Now().UTC().Add(3 * time.Hour).Unix()
	if ActiveUntil != nil {
		until = *ActiveUntil >= now

	}
	if ActiveFrom != nil {
		from = *ActiveFrom <= now

	}
	active := count && until && from
	if !active {
		if active != *promocode.Active {
			_, err := sq.Update("promos").
				Set("active", active).
				Where(sq.Eq{"promo_id": promo.PromoID}).
				PlaceholderFormat(sq.Dollar).
				RunWith(pr.db.Db).
				Exec()
			if err != nil {
				return "", service.ErrNoPermission
			}
		}
		return "", service.ErrNoPermission
	}
	if *promocode.Mode == "COMMON" {
		count = *promocode.MaxCount > *promocode.UsedCount
		promocode.UsedPromoUnique = models.StringSlice{}
	} else {
		count = len(promocode.PromoUnique) > len(promocode.UsedPromoUnique)

	}
	active = count && until && from
	_, err = sq.Update("promos").
		Set("active", active).
		Set("used_count", promocode.UsedCount).
		Set("used_promo_unique", promocode.UsedPromoUnique).
		Where(sq.Eq{"promo_id": promo.PromoID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return "", service.ErrNoPermission
	}
	_, err = sq.Insert("activations").
		Columns("activate_time", "country", "promo_id", "id").
		Values(now, promo.Country, promo.PromoID, promo.UserID).
		PlaceholderFormat(sq.Dollar).
		RunWith(pr.db.Db).
		Exec()
	if err != nil {
		return "", err
	}
	return res, nil
}
func (pr *PostgresRepo) GetUserHistory(ctx context.Context, sortRules *models.HistorySort) ([]models.FeedUserResponse, int, error) {
	activations := []models.FeedUserResponse{}
	rows, err := sq.Select("promo_id,activate_time").
		From("activations").
		Where(sq.Eq{"id": sortRules.UserID}).
		PlaceholderFormat(sq.Dollar).
		OrderBy("seq_id DESC").
		RunWith(pr.db.Db).Query()
	if err != nil {
		return nil, 0, err
	}
	var count int = 0
	for rows.Next() {
		if sortRules.Offset <= count && len(activations) < sortRules.Limit {
			var ActiveFrom, ActiveUntil *int64
			var promo models.FeedUserResponse
			var mode string
			var uniq, usedUniq models.StringSlice
			var maxCount, usedCount int
			var activate_time int64
			scanErr := rows.Scan(&promo.PromoId, &activate_time)
			if scanErr != nil {
				return nil, 0, scanErr
			}
			err := sq.Select("company_id", "company_name", "description", "image_url", "active", "like_count", "comment_count", "used_count", "max_count", "active_from", "active_until", "mode", "promo_unique", "used_promo_unique").
				From("promos").
				Where(sq.Eq{"promo_id": promo.PromoId}).
				PlaceholderFormat(sq.Dollar).
				RunWith(pr.db.Db).
				Scan(&promo.CompanyId, &promo.CompanyName, &promo.Description, &promo.ImageUrl, &promo.Active, &promo.LikeCount, &promo.CommentCount, &usedCount, &maxCount, &ActiveFrom, &ActiveUntil, &mode, &uniq, &usedUniq)
			if err != nil {
				return nil, 0, err
			}

			count, until, from := *promo.Active, *promo.Active, *promo.Active
			now := time.Now().UTC().Add(3 * time.Hour).Unix()
			if mode == "COMMON" {
				if *promo.Active != (maxCount > usedCount) {
					count = maxCount > usedCount
				}
			} else {
				if *promo.Active != (len(uniq) > len(usedUniq)) {
					count = len(uniq) > len(usedUniq)
				}
			}
			if ActiveFrom != nil {
				if *promo.Active != (*ActiveFrom > now) {
					from = *ActiveFrom <= now
				}

			}
			if ActiveUntil != nil {
				if *promo.Active != (*ActiveUntil < now) {
					until = *ActiveUntil >= now
				}
			}
			l, _ := zap.NewProduction()
			if (count && until && from) != *promo.Active {
				_, checkErr := sq.Update("promos").
					Where(sq.Eq{"promo_id": promo.PromoId}).
					Set("active", count && until && from).
					RunWith(pr.db.Db).
					Exec()
				if checkErr != nil {
					return nil, 0, checkErr
				}
				l.Sugar().Info(promo.PromoId, " not match")
			}
			*promo.Active = count && until && from
			statErr := sq.Select("is_liked_by_user").
				From("promosstat").
				Where(sq.And{sq.Eq{"id": sortRules.UserID}, sq.Eq{"promo_id": *promo.PromoId}}).
				PlaceholderFormat(sq.Dollar).
				RunWith(pr.db.Db).QueryRow().Scan(&promo.IsLiked)
			if statErr != nil && statErr != sql.ErrNoRows {
				return nil, 0, statErr
			}
			falseVal := false
			trueVal := true
			if statErr == sql.ErrNoRows {
				promo.IsLiked = &falseVal
			}
			promo.IsActivated = &trueVal
			l.Sugar().Info(activate_time, " ", *promo.PromoId)
			activations = append(activations, promo)
		}
		count++

	}
	return activations, count, nil
}
