package book

import (
	"context"

	pg "github.com/devpablocristo/monorepo/pkg/databases/sql/postgresql/pq"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/book/usecases/domain"
)

type repository struct {
	pgInst pg.Repository
}

func NewRepository(inst pg.Repository) Repository {
	return &repository{
		pgInst: inst,
	}
}

func (b *repository) GetBook(ctx context.Context, book *domain.Book, id int) (*domain.Book, error) {
	row := b.pgInst.DB().QueryRow("SELECT * FROM books WHERE id=$1", id)
	err := row.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
	return book, err
}

func (b *repository) AddBook(ctx context.Context, book *domain.Book) (int, error) {
	var id int
	err := b.pgInst.DB().QueryRow("INSERT INTO books (title, author, year) VALUES($1, $2, $3) RETURNING id;",
		book.Title, book.Author, book.Year).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (b *repository) UpdateBook(ctx context.Context, book *domain.Book) (int64, error) {
	result, err := b.pgInst.DB().Exec("UPDATE books SET title=$1, author=$2, year=$3 WHERE id=$4",
		book.Title, book.Author, book.Year, book.ID)
	if err != nil {
		return 0, err
	}
	rowsUpdated, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsUpdated, nil
}

func (b *repository) RemoveBook(ctx context.Context, id int) (int64, error) {
	result, err := b.pgInst.DB().Exec("DELETE FROM books WHERE id=$1", id)
	if err != nil {
		return 0, err
	}
	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsDeleted, nil
}
