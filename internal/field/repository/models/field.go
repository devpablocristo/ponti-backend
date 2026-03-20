package models

import (
	"time"

	"github.com/shopspring/decimal"

	fielddom "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	invdom "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	leasetypemod "github.com/devpablocristo/ponti-backend/internal/lease-type/repository/models"
	leasetypedom "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"
	lotmod "github.com/devpablocristo/ponti-backend/internal/lot/repository/models"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type Field struct {
	ID               int64            `gorm:"primaryKey;autoIncrement;column:id"`
	Name             string           `gorm:"size:100;not null;column:name"`
	ProjectID        int64            `gorm:"not null;index;column:project_id"`
	LeaseTypeID      int64            `gorm:"not null;column:lease_type_id"`
	LeaseTypePercent *decimal.Decimal `gorm:"column:lease_type_percent"`
	LeaseTypeValue   *decimal.Decimal `gorm:"column:lease_type_value"`
	sharedmodels.Base
	FieldInvestors []FieldInvestor         `gorm:"foreignKey:FieldID;references:ID"`
	Lots           []lotmod.Lot            `gorm:"foreignKey:FieldID"`
	LeaseType      *leasetypemod.LeaseType `gorm:"foreignKey:LeaseTypeID;references:ID"`
}

type FieldInvestor struct {
	FieldID    int64 `gorm:"primaryKey;autoIncrement:false;column:field_id"`
	InvestorID int64 `gorm:"primaryKey;autoIncrement:false;column:investor_id"`
	Percentage int   `gorm:"not null;column:percentage"`
	sharedmodels.Base

	Investor invdom.Investor `gorm:"foreignKey:InvestorID;references:ID"`
}

// FROM DOMAIN (para INSERT: solo escalares)
func FromDomain(d *fielddom.Field) *Field {
	return &Field{
		ID:               d.ID,
		Name:             d.Name,
		ProjectID:        d.ProjectID,
		LeaseTypeID:      d.LeaseType.ID,
		LeaseTypePercent: d.LeaseTypePercent,
		LeaseTypeValue:   d.LeaseTypeValue,
		Base: sharedmodels.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		},
	}
}

func (m *Field) ToDomain() *fielddom.Field {
	base := shareddomain.Base{
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	var archivedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		archivedAt = &t
	}
	ltName := ""
	if m.LeaseType != nil {
		ltName = m.LeaseType.Name
	}
	return &fielddom.Field{
		ID:               m.ID,
		Name:             m.Name,
		ProjectID:        m.ProjectID,
		LeaseType:        &leasetypedom.LeaseType{ID: m.LeaseTypeID, Name: ltName},
		LeaseTypePercent: m.LeaseTypePercent,
		LeaseTypeValue:   m.LeaseTypeValue,
		ArchivedAt:       archivedAt,
		Base:             base,
	}
}
