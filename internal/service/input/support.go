package input

type GetSupportMessages struct {
	ThreadID int
	Limit    int
	BeforeID *int
}

type CreateSupportMessage struct {
	ThreadID int
	Text     string
}
