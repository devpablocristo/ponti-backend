package book

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/book/usecases/domain"
)

type Repository interface {
	GetBook(context.Context, *domain.Book, int) (*domain.Book, error)
	AddBook(context.Context, *domain.Book) (int, error)
	UpdateBook(context.Context, *domain.Book) (int64, error)
	RemoveBook(context.Context, int) (int64, error)
}
