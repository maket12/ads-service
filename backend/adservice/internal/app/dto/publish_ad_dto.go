package dto

import "github.com/google/uuid"

type PublishAdInput struct {
	AdID uuid.UUID
}

type PublishAdOutput struct {
	Success bool
}
