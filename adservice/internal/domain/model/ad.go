package model

import (
	"errors"
	"time"

	pkgerrs "github.com/maket12/ads-service/pkg/errs"

	"github.com/google/uuid"
)

var (
	ErrAdCantBePublished = errors.New("ad cannot be published")
	ErrAdCantBeRejected  = errors.New("ad cannot be rejected")
	ErrAdCantBeDeleted   = errors.New("ad cannot be deleted")
)

type AdStatus string

const (
	AdPublished    AdStatus = "published"
	AdOnModeration AdStatus = "on_moderation"
	AdRejected     AdStatus = "rejected"
	AdDeleted      AdStatus = "deleted"
)

const (
	minTitleLen       = 5
	maxDescriptionLen = 2048
)

// ================ Rich model for Ad ================

type Ad struct {
	id          uuid.UUID
	sellerID    uuid.UUID
	title       string
	description *string
	price       int64 // in cents
	status      AdStatus
	images      []string
	createdAt   time.Time
	updatedAt   time.Time
}

func NewAd(
	sellerID uuid.UUID,
	title string,
	description *string,
	price int64,
	images []string,
) (*Ad, error) {
	if sellerID == uuid.Nil {
		return nil, pkgerrs.NewValueInvalidError("seller_id")
	}
	if title == "" {
		return nil, pkgerrs.NewValueRequiredError("title")
	}
	if len(title) < minTitleLen {
		return nil, pkgerrs.NewValueInvalidError("title")
	}
	if description != nil {
		if len(*description) == 0 {
			return nil, pkgerrs.NewValueRequiredError("description")
		} else if len(*description) > maxDescriptionLen {
			return nil, pkgerrs.NewValueInvalidError("description")
		}
	}
	if price < 0 {
		return nil, pkgerrs.NewValueInvalidError("price")
	}
	if images != nil {
		if len(images) == 0 {
			return nil, pkgerrs.NewValueInvalidError("images")
		}
	}

	var imagesCopy []string
	if images != nil {
		imagesCopy = make([]string, len(images))
		copy(imagesCopy, images)
	}

	now := time.Now()

	return &Ad{
		id:          uuid.New(),
		sellerID:    sellerID,
		title:       title,
		description: description,
		price:       price,
		status:      AdOnModeration,
		images:      imagesCopy,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func RestoreAd(
	id, sellerID uuid.UUID,
	title string,
	description *string,
	price int64,
	status AdStatus,
	images []string,
	createdAt time.Time,
	updatedAt time.Time,
) *Ad {
	var imagesCopy []string
	if images != nil {
		imagesCopy = make([]string, len(images))
		copy(imagesCopy, images)
	}
	return &Ad{
		id:          id,
		sellerID:    sellerID,
		title:       title,
		description: description,
		price:       price,
		status:      status,
		images:      imagesCopy,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// ================ Read-Only ================

func (ad *Ad) ID() uuid.UUID        { return ad.id }
func (ad *Ad) SellerID() uuid.UUID  { return ad.sellerID }
func (ad *Ad) Title() string        { return ad.title }
func (ad *Ad) Description() *string { return ad.description }
func (ad *Ad) Price() int64         { return ad.price }
func (ad *Ad) Status() AdStatus     { return ad.status }
func (ad *Ad) Images() []string {
	if ad.images == nil {
		return nil
	}
	cp := make([]string, len(ad.images))
	copy(cp, ad.images)
	return cp
}
func (ad *Ad) CreatedAt() time.Time { return ad.createdAt }
func (ad *Ad) UpdatedAt() time.Time { return ad.updatedAt }

func (ad *Ad) IsPublished() bool    { return ad.status == AdPublished }
func (ad *Ad) IsOnModeration() bool { return ad.status == AdOnModeration }
func (ad *Ad) IsRejected() bool     { return ad.status == AdRejected }
func (ad *Ad) IsDeleted() bool      { return ad.status == AdDeleted }

func (ad *Ad) CanBePublished() bool { return ad.IsOnModeration() }
func (ad *Ad) CanBeRejected() bool  { return ad.IsOnModeration() }
func (ad *Ad) CanBeDeleted() bool   { return ad.IsPublished() }

// ================ Mutation ================

func (ad *Ad) Publish() error {
	if !ad.CanBePublished() {
		return ErrAdCantBePublished
	}

	ad.status = AdPublished
	ad.updatedAt = time.Now()

	return nil
}

func (ad *Ad) Reject() error {
	if !ad.CanBeRejected() {
		return ErrAdCantBeRejected
	}

	ad.status = AdRejected
	ad.updatedAt = time.Now()

	return nil
}

func (ad *Ad) Delete() error {
	if !ad.CanBeDeleted() {
		return ErrAdCantBeDeleted
	}

	ad.status = AdDeleted
	ad.updatedAt = time.Now()

	return nil
}

func (ad *Ad) Update(title, description *string, price *int64, images []string) error {
	if title != nil && len(*title) < minTitleLen {
		return pkgerrs.NewValueInvalidError("title")
	}
	if description != nil && len(*description) > maxDescriptionLen {
		return pkgerrs.NewValueInvalidError("description")
	}
	if price != nil && *price < 0 {
		return pkgerrs.NewValueInvalidError("price")
	}

	if title != nil {
		ad.title = *title
	}
	if description != nil {
		ad.description = description
	}
	if price != nil {
		ad.price = *price
	}
	if images != nil {
		ad.images = make([]string, len(images))
		copy(ad.images, images)
	}

	ad.updatedAt = time.Now()

	return nil
}
