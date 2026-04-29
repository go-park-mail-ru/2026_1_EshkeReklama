package input

type RegisterAdvertiser struct {
	Name     string
	Email    string
	Phone    string
	Password string
}

type UpdateAdvertiserProfile struct {
	AdvertiserID int
	Name         *string
	Surname      *string
	Email        *string
	Phone        *string
	Company      *string
	City         *string
	Tariff       *string
}
