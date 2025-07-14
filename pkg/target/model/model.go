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
	// MatchAllCampaigns   map[string]struct{}
}

type TargetCategory int

const (
	TargetCategoryAppID TargetCategory = iota + 1
	TargetCategoryCountry
	TargetCategoryOS
)

type Campaign struct {
	ID             uuid.UUID
	Name           string
	ImageUrl       string
	CTA            string
	ActivityStatus bool
	IsDeleted      bool
}
