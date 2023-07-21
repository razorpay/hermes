package models

import (
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Team struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Name is the name of the team
	Name string `gorm:"default:null;index;not null;type:citext;unique"`

	// BUName is the business unit that this team belongs to
	BUID uuid.UUID `gorm:"default:null;not null;type:citext;"`

	// UserSubscribers are the users that subscribed to this product.
	BU Product
}

// Upsert upserts a team along with its associated BU into the database.
// If a BU with the given name already exists, it is used; otherwise, an error is returned.
func (t *Team) Upsert(db *gorm.DB, prdName string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Check if the BU with the given name already exists.
		var existingPrd Product
		if err := tx.Where("name = ?", prdName).First(&existingPrd).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("BU '%s' does not exist", prdName)
			}
			return fmt.Errorf("error querying BU: %w", err)
		}

		// Set the BUID of the team to the ID of the BU.
		t.BUID = existingPrd.ID
		t.BU = existingPrd

		// Upsert the team.
		if err := tx.Where(Team{Name: t.Name}).
			Preload("BU").
			Omit(clause.Associations).
			Assign(*t).
			Clauses(clause.OnConflict{DoNothing: true}).
			FirstOrCreate(&t).Error; err != nil {
			return fmt.Errorf("error upserting team: %w", err)
		}

		// Update the BU.
		existingPrd.Teams = append(existingPrd.Teams, *t)
		if err := tx.Save(&existingPrd).Error; err != nil {
			return err
		}

		return nil
	})
}

// Get gets a team from database db by name, and assigns it back to the
// receiver.
func (t *Team) Get(db *gorm.DB) error {
	if err := validation.ValidateStruct(t,
		validation.Field(
			&t.ID,
			validation.When(t.Name == "",
				validation.Required.Error("either ID or Name is required")),
		),
		validation.Field(
			&t.Name,
			validation.When(t.ID == uuid.Nil,
				validation.Required.Error("either ID or Name is required"))),
	); err != nil {
		return err
	}

	return db.
		Where(Team{Name: t.Name}).
		Preload(clause.Associations).
		First(&t).
		Error
}
