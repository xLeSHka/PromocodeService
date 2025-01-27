package models

type Company struct {
	CompanyID string `json:"company_id" db:"company_id" redis:"company_id"`
	Name      string `json:"name" db:"name" redis:"name"`
	Email     string `json:"email" db:"email" redis:"email"`
	Password  []byte `json:"password" db:"password" redis:"password"`
}

