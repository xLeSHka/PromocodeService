package models

type Company struct {
	CompanyID string `json:"company_id" db:"company_id" redis:"company_id"`
	Name      string `json:"name" db:"name" redis:"name"`
	Email     string `json:"email" db:"email" redis:"email"`
	Password  []byte `json:"password" db:"password" redis:"password"`
}

// func (c *Company) GetName() string {
// 	name := make([]uint16, 0)
// 	json.Unmarshal(c.Name, &name)
// 	return string(utf16.Decode(name))
// }
// func (c *Company) SetName(name string) {
// 	data, _ := json.Marshal(utf16.Encode([]rune(name)))
// 	c.Name = data
// }
