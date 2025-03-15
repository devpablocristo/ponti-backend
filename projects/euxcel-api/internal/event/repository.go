package event

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	mng "github.com/devpablocristo/monorepo/pkg/databases/nosql/mongodb/mongo-driver"

	models "github.com/devpablocristo/monorepo/projects/qh/internal/event/repository/models"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/event/usecases/domain"
)

type mongoRepository struct {
	repository mng.Repository
}

func NewRepository(r mng.Repository) Repository {
	return &mongoRepository{
		repository: r,
	}
}

func (r *mongoRepository) ListEvents(ctx context.Context) ([]domain.Event, error) {
	cursor, err := r.repository.DB().Collection("events").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := cursor.Close(ctx); closeErr != nil {
			log.Printf("Error closing cursor: %v", closeErr)
		}
	}()

	var es models.EventList
	if err := cursor.All(ctx, &es); err != nil {
		return nil, err
	}

	return es.ToDomain(), nil
}

func (ed *mongoRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
	var e models.CreateEvent
	e.CreatedAt = time.Now()
	if _, err := ed.repository.DB().Collection("events").InsertOne(ctx, e); err != nil {
		return err
	}
	return nil
}

// func (ed *mongoRepository) FindByID(ctx context.Context, ID string) (*domain.Event, error) {
// 	eventID, err := primitive.ObjectIDFromHex(ID)
// 	if err != nil {
// 		log.Printf("Error converting ID: %v", err)
// 		return nil, err
// 	}

// 	filter := bson.M{"_id": eventID}
// 	dao := &Repository{}
// 	if err := ed.repository.DB().Collection("events").FindOne(ctx, filter).Decode(dao); err != nil {
// 		log.Printf("Error decoding event: %v", err)
// 		return nil, err
// 	}
// 	return EventDaoToDomain(dao), nil
// }

// func (ed *mongoRepository) Update(ctx context.Context, event *domain.Event, ID string) (*domain.Event, error) {
// 	eventID, err := primitive.ObjectIDFromHex(ID)
// 	if err != nil {
// 		log.Printf("Error converting ID: %v", err)
// 		return nil, err
// 	}

// 	dao := EventDomainToDao(event)
// 	updateFields := checkEventFields(dao)

// 	if updateFields == nil {
// 		log.Printf("Error: event is nil")
// 		return nil, fmt.Errorf("event is nil")
// 	}

// 	if len(updateFields) == 0 {
// 		log.Printf("Error: no fields to update")
// 		return nil, fmt.Errorf("no fields to update")
// 	}

// 	updateFields["updated_at"] = time.Now()

// 	filter := bson.M{"_id": eventID}
// 	update := bson.M{"$set": updateFields}

// 	result, err := ed.repository.DB().Collection("events").UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		log.Printf("Error updating event: %v", err)
// 		return nil, err
// 	}

// 	if result.MatchedCount == 0 {
// 		log.Printf("Error: no event found with ID: %s", eventID.Hex())
// 		return nil, fmt.Errorf("no event found with ID: %s", eventID.Hex())
// 	}

// 	var updatedDAO Repository
// 	if err := ed.repository.DB().Collection("events").FindOne(ctx, filter).Decode(&updatedDAO); err != nil {
// 		log.Printf("Error decoding updated event: %v", err)
// 		return nil, err
// 	}

// 	log.Printf("updated %d events", result.ModifiedCount)

// 	return EventDaoToDomain(&updatedDAO), nil
// }

// func (ed *mongoRepository) HardDelete(ctx context.Context, ID string) (*domain.Event, error) {
// 	eventID, err := primitive.ObjectIDFromHex(ID)
// 	if err != nil {
// 		log.Printf("Error converting ID: %v", err)
// 		return nil, err
// 	}

// 	filter := bson.M{"_id": eventID}

// 	deletedDAO := &Repository{}
// 	if err := ed.repository.DB().Collection("events").FindOne(ctx, filter).Decode(deletedDAO); err != nil {
// 		log.Printf("Error decoding deleted event: %v", err)
// 		return nil, err
// 	}

// 	result, err := ed.repository.DB().Collection("events").DeleteOne(ctx, filter)
// 	if err != nil {
// 		log.Printf("Error deleting event: %v", err)
// 		return nil, err
// 	}

// 	log.Printf("deleted %d events", result.DeletedCount)
// 	return EventDaoToDomain(deletedDAO), nil
// }

// func (ed *mongoRepository) SoftDelete(ctx context.Context, ID string) (*domain.Event, error) {
// 	eventID, err := primitive.ObjectIDFromHex(ID)
// 	if err != nil {
// 		log.Printf("Error converting ID: %v", err)
// 		return nil, err
// 	}

// 	filter := bson.M{"_id": eventID}
// 	update := bson.M{
// 		"$set": bson.M{
// 			"deleted_at": time.Now(),
// 			"deleted":    true,
// 		},
// 	}

// 	result, err := ed.repository.DB().Collection("events").UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		log.Printf("Error soft deleting event: %v", err)
// 		return nil, err
// 	}

// 	var updatedDAO *domain.Repository
// 	err = ed.repository.DB().Collection("events").FindOne(ctx, filter).Decode(&updatedDAO)
// 	if err != nil {
// 		log.Printf("Error decoding soft deleted event: %v", err)
// 		return nil, err
// 	}

// 	log.Printf("softdeleted %d events", result.ModifiedCount)

// 	return EventDaoToDomain(updatedDAO), nil
// }

// func (ed *mongoRepository) SoftUndelete(ctx context.Context, ID string) (*domain.Event, error) {
// 	eventID, err := primitive.ObjectIDFromHex(ID)
// 	if err != nil {
// 		log.Printf("Error converting ID: %v", err)
// 		return nil, err
// 	}

// 	filter := bson.M{"_id": eventID}
// 	update := bson.M{
// 		"$set": bson.M{
// 			"deleted_at": nil,
// 			"deleted":    false,
// 		},
// 	}

// 	result, err := ed.repository.DB().Collection("events").UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		log.Printf("Error soft undeleting event: %v", err)
// 		return nil, err
// 	}

// 	var updatedDAO *domain.Repository
// 	err = ed.repository.DB().Collection("events").FindOne(ctx, filter).Decode(&updatedDAO)
// 	if err != nil {
// 		log.Printf("Error decoding soft undeleted event: %v", err)
// 		return nil, err
// 	}

// 	log.Printf("softundeleted %d events", result.ModifiedCount)

// 	return EventDaoToDomain(updatedDAO), nil
// }

// func (ed *mongoRepository) AddUserToEvent(ctx context.Context, eventID string, user *usr.User) (*domain.Event, error) {
// 	// Implementación pendiente
// 	return nil, fmt.Errorf("AddUserToEvent method not implemented")
// }

// // helpers
// func checkEventFields(event *domain.Repository) map[string]any {
// 	if event == nil {
// 		return nil
// 	}

// 	updateFields := make(map[string]any)

// 	if event.EventName != "" {
// 		updateFields["event_name"] = event.EventName
// 	}

// 	if event.Description != "" {
// 		updateFields["description"] = event.Description
// 	}

// 	if event.Date != "" {
// 		updateFields["date"] = event.Date
// 	}

// 	if event.Location != nil && (event.Location.Address != "" || event.Location.City != "" || event.Location.State != "" || event.Location.Country != "" || event.Location.PostalCode != "") {
// 		updateFields["location"] = event.Location
// 	}

// 	if len(event.Attendees) > 0 {
// 		updateFields["attendees"] = event.Attendees
// 	}

// 	if len(event.Planners) > 0 {
// 		updateFields["planners"] = event.Planners
// 	}

// 	if len(event.Tags) > 0 {
// 		updateFields["tags"] = event.Tags
// 	}

// 	return updateFields
// }
