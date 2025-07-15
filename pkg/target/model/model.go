package model

import (
	"sync"

	"github.com/google/uuid"
)

type TargetingData struct {
	TargetMutex         sync.RWMutex
	Campaigns           map[uuid.UUID]*Campaign
	IncludeCountryIndex map[string][]uuid.UUID
	ExcludeCountryIndex map[string][]uuid.UUID
	IncludeOSIndex      map[string][]uuid.UUID
	IncludeAppIndex     map[string][]uuid.UUID
}

type TargetCategory int

const (
	TargetCategoryAppID TargetCategory = iota + 1
	TargetCategoryCountry
	TargetCategoryOS
)

type Campaign struct {
	ID               uuid.UUID
	CampaignStringID string
	Name             string
	ImageUrl         string
	CTA              string
	ActivityStatus   bool
	IsDeleted        bool
}

type DeliveryServiceRequest struct {
	AppID   string `json:"app" validate:"required"`
	OS      string `json:"os" validate:"required"`
	Country string `json:"country" validate:"required"`
}

type DeliveryServiceResponse struct {
	CampaignStringID string `json:"cid"`
	Image            string `json:"img"`
	Cta              string `json:"cta"`
}
