package models

type CompanySort struct {
	CompanyId string
	Limit     int
	Offset    int
	SortBy    string
	Countries []string
}
type UserSort struct {
	Id       string
	Limit    int
	Offset   int
	Category *string
	Country  string
	Age      int
	Active   *bool
}
type CommentSort struct {
	PromoId string
	Limit   int
	Offset  int
}
type HistorySort struct {
	UserID  string
	Limit   int
	Offset  int
}
