package domain

import (
	"time"
)

type TargetType string

const (
	TargetEvent TargetType = "event"
	TargetUser  TargetType = "user"
)

type Rating struct {
	ID         string     `json:"id,omitempty" bson:"_id,omitempty"`
	RaterID    string     `json:"rater_id" bson:"rater_id" validate:"required"`   // ID del usuario que califica
	TargetID   string     `json:"target_id" bson:"target_id" validate:"required"` // ID del evento o usuario que recibe la calificación
	TargetType TargetType `json:"target_type" bson:"target_type" validate:"required,oneof=event user"`
	Score      float32    `json:"score" bson:"score" validate:"required,min=0,max=5"`
	Comment    string     `json:"comment,omitempty" bson:"comment"`
	CreatedAt  time.Time  `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at,omitempty" bson:"updated_at"`
}
