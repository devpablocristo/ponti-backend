package rating

import domain "github.com/devpablocristo/monorepo/projects/qh/internal/rating/usecases/domain"

type Repository interface {
	CreateRating(r *domain.Rating) (*domain.Rating, error)
	GetRatingByID(ID string) (*domain.Rating, error)
	GetRatingByTarget(targetID string, targetType domain.TargetType) ([]*domain.Rating, error)
	UpdateRating(r *domain.Rating) (*domain.Rating, error)
	GetRatingByRaterAndTarget(raterID, targetID string, targetType domain.TargetType) (*domain.Rating, error) // Para verificar si ya existe una calificación de un usuario a un evento
}

type GrpcClient any

type UseCases any
