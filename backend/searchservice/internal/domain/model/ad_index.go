package model

import (
	pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"
)

// ================ Rich model for Ad Index ================

type AdIndex struct {
	id          string
	title       string
	description string
	price       int64 // in cents
	category    Category
	mainImage   string
}

func NewAdIndex(
	id, title, description string,
	price int64,
	rawCategory, mainImage string,
) (*AdIndex, error) {
	if id == "" {
		return nil, pkgerrs.NewValueRequiredError("id")
	}
	if title == "" {
		return nil, pkgerrs.NewValueRequiredError("title")
	}
	if price < 0 {
		return nil, pkgerrs.NewValueInvalidError("price")
	}

	category, err := NewCategory(rawCategory)
	if err != nil {
		return nil, err
	}

	if mainImage == "" {
		return nil, pkgerrs.NewValueRequiredError("main_image")
	}

	return &AdIndex{
		id:          id,
		title:       title,
		description: description,
		price:       price,
		category:    category,
		mainImage:   mainImage,
	}, nil
}

func RestoreAdIndex(
	id, title, description string,
	price int64,
	category Category,
	mainImage string,
) *AdIndex {
	return &AdIndex{
		id:          id,
		title:       title,
		description: description,
		price:       price,
		category:    category,
		mainImage:   mainImage,
	}
}

// ================ Read-Only ================

func (ind *AdIndex) ID() string          { return ind.id }
func (ind *AdIndex) Title() string       { return ind.title }
func (ind *AdIndex) Description() string { return ind.description }
func (ind *AdIndex) Price() int64        { return ind.price }
func (ind *AdIndex) Category() Category  { return ind.category }
func (ind *AdIndex) MainImage() string   { return ind.mainImage }
