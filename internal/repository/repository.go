package repository

type Repository struct {
	Ads   *AdsRepository
	Users *UserRepository
}

func InitRepository() *Repository {
	return &Repository{
		Ads:   InitAdsRepository(),
		Users: InitUserRepository(),
	}
}
