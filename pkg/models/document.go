package models

import (
	// "errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Document is a model for a document.
type Document struct {
	gorm.Model

	// GoogleFileID is the Google Drive file ID of the document.
	GoogleFileID string `gorm:"index;not null;unique"`

	// Reviewers is the list of users whose approval is requested for the
	// document.
	Reviewers []*User `gorm:"many2many:document_reviews;"`

	// last date to review a doc for a reviewer
	DueDate string
	// Contributors are users who have contributed to the document.
	Contributors []*User `gorm:"many2many:document_contributors;"`

	// Contributors are users who have contributed to the document.
	ReviewedBy []*User `gorm:"many2many:document_reviewedBy;"`

	// CustomFields contains custom fields.
	CustomFields []*DocumentCustomField

	// DocumentCreatedAt is the time of document creation.
	DocumentCreatedAt time.Time

	// DocumentModifiedAt is the time the document was last modified.
	DocumentModifiedAt time.Time

	// DocumentNumber is a document number unique to each bu. It
	// pairs with the product abbreviation to form a document identifier
	// (e.g., "TF-123").
	// DocumentNumber int `gorm:"index:latest_product_number"`

	// DocumentType is the document type.
	DocumentType   DocumentType
	DocumentTypeID uint

	// Imported is true if the document was not created through the application.
	Imported bool

	// Locked is true if the document cannot be updated (may be in a bad state).
	Locked bool

	// Owner is the owner of the document.
	Owner   *User `gorm:"default:null;not null"`
	OwnerID *uint `gorm:"default:null"`

	// Product is the product or area that the document relates to.
	Product   Product
	ProductID uuid.UUID `gorm:"index:latest_product_number"`

	// Team is the team/pod inside the Product
	Team   Team
	TeamID uuid.UUID `gorm:"index"` // Foreign key column referencing the teams table

	// Project is the project inside specific team
	Project   Project
	ProjectID uuid.UUID `gorm:"index"` // Foreign key column referencing the projects table

	// Status is the status of the document.
	Status DocumentStatus

	// Summary is a summary of the document.
	Summary string

	// Title is the title of the document. It only contains the title, and not the
	// product abbreviation, document number, or document type.
	Title string
}

// Documents is a slice of documents.
type Documents []Document

// DocumentStatus is the status of the document (e.g., "Draft", "In-Review",
// "Reviewed", "Obsolete").
type DocumentStatus int

const (
	UnspecifiedDocumentStatus DocumentStatus = iota
	DraftDocumentStatus
	InReviewDocumentStatus
	ReviewedDocumentStatus
	ObsoleteDocumentStatus
)

// BeforeSave is a hook used to find associations before saving.
func (d *Document) BeforeSave(tx *gorm.DB) error {
	if err := d.getAssociations(tx); err != nil {
		return fmt.Errorf("error getting associations: %w", err)
	}

	return nil
}

// Create creates a document in database db.
func (d *Document) Create(db *gorm.DB) error {
	if err := validation.ValidateStruct(d,
		validation.Field(
			&d.ID,
			validation.When(d.GoogleFileID == "",
				validation.Required.Error("either ID or GoogleFileID is required"),
			),
		),
		validation.Field(
			&d.GoogleFileID,
			validation.When(d.ID == 0,
				validation.Required.Error("either ID or GoogleFileID is required"),
			),
		),
	); err != nil {
		return err
	}

	if err := d.createAssocations(db); err != nil {
		return fmt.Errorf("error creating associations: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&d).
			Where(Document{GoogleFileID: d.GoogleFileID}).
			Omit(clause.Associations). // We get associations in the BeforeSave hook.
			Create(&d).
			Error; err != nil {
			return err
		}

		if err := d.replaceAssocations(tx); err != nil {
			return fmt.Errorf("error replacing associations: %w", err)
		}

		return nil
	})
}

// Find finds all documents from database db with the provided query, and
// assigns them to the receiver.
func (d *Documents) Find(
	db *gorm.DB, query interface{}, queryArgs ...interface{}) error {

	return db.
		Where(query, queryArgs...).
		Preload(clause.Associations).
		Find(&d).Error
}

// FirstOrCreate finds the first document by Google file ID or creates a new
// record if it does not exist.
// func (d *Document) FirstOrCreate(db *gorm.DB) error {
// 	return db.
// 		Where(Document{GoogleFileID: d.GoogleFileID}).
// 		Preload(clause.Associations).
// 		FirstOrCreate(&d).Error
// }

// Get gets a document from database db by Google file ID, and assigns it to the
// receiver.
func (d *Document) Get(db *gorm.DB) error {
	if err := validation.ValidateStruct(d,
		validation.Field(
			&d.ID,
			validation.When(d.GoogleFileID == "",
				validation.Required.Error("either ID or GoogleFileID is required"),
			),
		),
		validation.Field(
			&d.GoogleFileID,
			validation.When(d.ID == 0,
				validation.Required.Error("either ID or GoogleFileID is required"),
			),
		),
	); err != nil {
		return err
	}

	if err := db.
		Where(Document{GoogleFileID: d.GoogleFileID}).
		Preload(clause.Associations).
		First(&d).
		Error; err != nil {
		return err
	}

	if err := d.getAssociations(db); err != nil {
		return fmt.Errorf("error getting associations: %w", err)
	}

	return nil
}

// GetLatestProductNumber gets the latest document number for a product.
// func GetLatestProductNumber(db *gorm.DB,
// 	documentTypeName, productName string) (int, error) {
// 	// Validate required fields.
// 	if err := validation.Validate(db, validation.Required); err != nil {
// 		return 0, err
// 	}
// 	if err := validation.Validate(productName, validation.Required); err != nil {
// 		return 0, err
// 	}

// 	// Get document type.
// 	dt := DocumentType{
// 		Name: documentTypeName,
// 	}
// 	if err := dt.Get(db); err != nil {
// 		return 0, fmt.Errorf("error getting document type: %w", err)
// 	}

// 	// Get product.
// 	p := Product{
// 		Name: productName,
// 	}
// 	if err := p.Get(db); err != nil {
// 		return 0, fmt.Errorf("error getting product: %w", err)
// 	}

// 	// Get document with largest document number.
// 	var d Document
// 	if err := db.
// 		Where(Document{
// 			DocumentTypeID: dt.ID,
// 			ProductID:      p.ID,
// 		}).
// 		Order("document_number desc").
// 		First(&d).
// 		Error; err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return 0, nil
// 		} else {
// 			return 0, err
// 		}
// 	}

// 	return d.DocumentNumber, nil
// }

// Upsert updates or inserts the receiver document into database db.
func (d *Document) Upsert(db *gorm.DB) error {
	if err := validation.ValidateStruct(d,
		validation.Field(
			&d.ID,
			validation.When(d.GoogleFileID == "",
				validation.Required.Error("either ID or GoogleFileID is required"),
			),
		),
		validation.Field(
			&d.GoogleFileID,
			validation.When(d.ID == 0,
				validation.Required.Error("either ID or GoogleFileID is required"),
			),
		),
	); err != nil {
		return err
	}

	// Create required associations.
	if err := d.createAssocations(db); err != nil {
		return fmt.Errorf("error creating associations: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&d).
			Where(Document{GoogleFileID: d.GoogleFileID}).
			Omit(clause.Associations). // We manage associations in the BeforeSave hook.
			Assign(*d).
			FirstOrCreate(&d).
			Error; err != nil {
			return err
		}

		// Replace has-many associations because we may have removed instances.
		if err := d.replaceAssocations(tx); err != nil {
			return fmt.Errorf("error replacing associations: %w", err)
		}

		if err := d.Get(tx); err != nil {
			return fmt.Errorf("error getting the document after upsert: %w", err)
		}

		return nil
	})
}

// createAssocations creates required assocations for a document.
func (d *Document) createAssocations(db *gorm.DB) error {
	// Find or create reviewers.
	var reviewers []*User
	for _, a := range d.Reviewers {
		if err := a.FirstOrCreate(db); err != nil {
			return fmt.Errorf("error finding or creating reviewer: %w", err)
		}
		reviewers = append(reviewers, a)
	}
	d.Reviewers = reviewers

	// Find or create contributors.
	var contributors []*User
	for _, c := range d.Contributors {
		if err := c.FirstOrCreate(db); err != nil {
			return fmt.Errorf("error finding or creating contributor: %w", err)
		}
		contributors = append(contributors, c)
	}
	d.Contributors = contributors

	// Find or create owner.
	if d.Owner != nil && d.Owner.EmailAddress != "" {
		if err := d.Owner.FirstOrCreate(db); err != nil {
			return fmt.Errorf("error finding or creating owner: %w", err)
		}
		d.OwnerID = &d.Owner.ID
	}

	// Get product if ProductID is not set.
	if d.ProductID == uuid.Nil && d.Product.Name != "" {
		if err := d.Product.Get(db); err != nil {
			return fmt.Errorf("error getting product: %w", err)
		}
		d.ProductID = d.Product.ID
	}

	// Get team if TeamID is not set.
	if d.TeamID == uuid.Nil && d.Team.Name != "" {
		if err := d.Team.Get(db); err != nil {
			return fmt.Errorf("error getting product: %w", err)
		}
		d.TeamID = d.Team.ID
	}

	// Get project if project id is not set.
	if d.ProjectID == uuid.Nil && d.Project.Name != "" {
		// Search for the project by its ID in the project table.
		project := d.Project
		if err := db.
			Where(project).
			Preload(clause.Associations).
			First(&project).
			Error; err != nil {
			return fmt.Errorf("error getting project: %w", err)
		}

		// Assign the found project to the document's Project field.
		d.Project = project
		d.ProjectID = project.ID
	}

	return nil
}

// getAssociations gets associations.
func (d *Document) getAssociations(db *gorm.DB) error {
	// Get reviewers.
	var reviewers []*User
	for _, a := range d.Reviewers {
		if err := a.Get(db); err != nil {
			return fmt.Errorf("error getting reviewer: %w", err)
		}
		reviewers = append(reviewers, a)
	}
	d.Reviewers = reviewers

	// Get contributors.
	var contributors []*User
	for _, c := range d.Contributors {
		if err := c.FirstOrCreate(db); err != nil {
			return fmt.Errorf("error getting contributor: %w", err)
		}
		contributors = append(contributors, c)
	}
	d.Contributors = contributors

	// Get custom fields.
	var customFields []*DocumentCustomField
	for _, c := range d.CustomFields {
		// If we already know the document type custom field ID, get the rest of its
		// data.
		if c.DocumentTypeCustomFieldID != 0 {
			if err := db.
				Model(&c.DocumentTypeCustomField).
				Where(DocumentTypeCustomField{
					Model: gorm.Model{
						ID: c.DocumentTypeCustomFieldID,
					},
				}).
				First(&c.DocumentTypeCustomField).
				Error; err != nil {
				return fmt.Errorf(
					"error getting document type custom field with known ID: %w", err)
			}
		}
		c.DocumentTypeCustomField.DocumentType.Name = d.DocumentType.Name
		if err := c.DocumentTypeCustomField.Get(db); err != nil {
			return fmt.Errorf("error getting document type custom field: %w", err)
		}
		c.DocumentTypeCustomFieldID = c.DocumentTypeCustomField.DocumentType.ID
		customFields = append(customFields, c)
	}
	d.CustomFields = customFields

	// Get document type.
	dt := d.DocumentType
	if err := dt.Get(db); err != nil {
		return fmt.Errorf("error getting document type: %w", err)
	}
	d.DocumentType = dt
	d.DocumentTypeID = dt.ID

	// Get owner.
	if d.Owner != nil && d.Owner.EmailAddress != "" {
		if err := d.Owner.Get(db); err != nil {
			return fmt.Errorf("error getting owner: %w", err)
		}
		d.OwnerID = &d.Owner.ID
	}

	// Get product.
	if d.Product.Name != "" {
		if err := d.Product.Get(db); err != nil {
			return fmt.Errorf("error getting product: %w", err)
		}
		d.ProductID = d.Product.ID
	}

	// get team
	if d.Team.Name != "" {
		if err := d.Team.Get(db); err != nil {
			return fmt.Errorf("error getting product: %w", err)
		}
		d.TeamID = d.Team.ID
	}

	// Get project if project id is not set.
	if d.Project.Name != "" {
		// Search for the project by its ID in the project table.
		project := d.Project
		if err := db.
			Where(project).
			Preload(clause.Associations).
			First(&project).
			Error; err != nil {
			return fmt.Errorf("error getting project: %w", err)
		}

		// Assign the found project to the document's Project field.
		d.Project = project
		d.ProjectID = project.ID
	}

	return nil
}

// replaceAssocations replaces assocations for a document.
func (d *Document) replaceAssocations(db *gorm.DB) error {
	// Replace reviewers.
	if err := db.
		Session(&gorm.Session{SkipHooks: true}).
		Model(&d).
		Association("Reviewers").
		Replace(d.Reviewers); err != nil {
		return err
	}

	// Replace contributors.
	if err := db.
		Session(&gorm.Session{SkipHooks: true}).
		Model(&d).
		Association("Contributors").
		Replace(d.Contributors); err != nil {
		return err
	}

	// Replace custom fields.
	if err := db.
		Session(&gorm.Session{SkipHooks: true}).
		Model(&d).
		Association("CustomFields").
		Replace(d.CustomFields); err != nil {
		return err
	}

	return nil
}
