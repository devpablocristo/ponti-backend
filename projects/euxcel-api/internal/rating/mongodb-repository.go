package rating

import (
	mongo "github.com/devpablocristo/monorepo/pkg/databases/nosql/mongodb/mongo-driver"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/rating/usecases/domain"
)

type MongoDbRepository struct {
	repository mongo.Repository
}

func NewMongoDbRepository(r mongo.Repository) Repository { //(Repository, error) {
	// r, err := mongo.Bootstrap("", "", "", "", "")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to initialize MongoDB client: %w", err)
	// }

	//	return &MongoDbRepository{
	//		repository: r,
	//	}, nil

	return &MongoDbRepository{
		repository: r,
	}
}

func (rat *MongoDbRepository) CreateRating(r *domain.Rating) (*domain.Rating, error) {
	return nil, nil
}

func (rat *MongoDbRepository) GetRatingByID(ID string) (*domain.Rating, error) {
	return nil, nil
}

func (rat *MongoDbRepository) GetRatingByTarget(targetID string, targetType domain.TargetType) ([]*domain.Rating, error) {
	return nil, nil
}

func (rat *MongoDbRepository) UpdateRating(r *domain.Rating) (*domain.Rating, error) {
	return nil, nil
}

func (rat *MongoDbRepository) GetRatingByRaterAndTarget(raterID, targetID string, targetType domain.TargetType) (*domain.Rating, error) {
	return nil, nil
}
