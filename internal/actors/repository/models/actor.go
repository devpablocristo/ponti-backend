package models

// Actor es un modelo GORM mínimo de la tabla actors, suficiente para asociaciones
// de SOLO LECTURA (ej. resolver el nombre del arrendatario al precargar field_lessees).
// La escritura del registro de identidad vive en internal/actors (SQL crudo / resolver),
// no acá: este struct no se usa para crear ni actualizar actores.
type Actor struct {
	ID          int64  `gorm:"primaryKey;column:id"`
	DisplayName string `gorm:"column:display_name"`
}

func (Actor) TableName() string { return "actors" }
