package repository

type SheetsRepository interface {
}

type sheetsRepository struct{}

func NewSheetsRepository() SheetsRepository {
	return &sheetsRepository{}
}
