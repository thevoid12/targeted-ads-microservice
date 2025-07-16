package target

import (
	"context"
	"errors"
	dbpkg "targetad/pkg/db"
	"targetad/pkg/target/model"

	"github.com/google/uuid"
)

var TargetCache *model.TargetingData // decaring the cache globally

// fetches the data from pgsql db and initializes the cache
func InitCache(ctx context.Context) (*model.TargetingData, error) {
	TargetCache = &model.TargetingData{}
	conn := dbpkg.GetConn()
	if conn == nil {
		return nil, errors.New("database connection is nil")
	}

	campaigns, err := conn.ListAllValidCampaigns(ctx)
	if err != nil {
		return nil, err
	}
	TargetCache.TargetMutex.Lock()

	TargetCache.Campaigns = make(map[uuid.UUID]*model.Campaign)
	TargetCache.ExcludeCountryIndex = make(map[string][]uuid.UUID)
	TargetCache.IncludeCountryIndex = make(map[string][]uuid.UUID)
	TargetCache.IncludeOSIndex = make(map[string][]uuid.UUID)
	TargetCache.IncludeAppIndex = make(map[string][]uuid.UUID)

	for _, campaign := range campaigns {
		TargetCache.Campaigns[campaign.ID.Bytes] = &model.Campaign{
			ID:               campaign.ID.Bytes,
			CampaignStringID: campaign.CampaignStringID,
			Name:             campaign.Name,
			ImageUrl:         campaign.ImageUrl,
			CTA:              campaign.Cta,
			ActivityStatus:   campaign.ActivityStatus,
			IsDeleted:        campaign.IsDeleted,
		}
	}
	TargetCache.TargetMutex.Unlock()
	// get all valid targetting rules from the database
	dbvals, err := conn.ListValidTargetingRules(ctx)
	if err != nil {
		return nil, err
	}
	TargetCache.TargetMutex.Lock()
	for _, val := range dbvals {
		switch val.Category {
		case int32(model.TargetCategoryAppID):
			TargetCache.IncludeAppIndex[val.Value] = append(TargetCache.IncludeAppIndex[val.Value], val.CampaignsID.Bytes)
		case int32(model.TargetCategoryCountry):
			if val.IsIncluded {
				TargetCache.IncludeCountryIndex[val.Value] = append(TargetCache.IncludeCountryIndex[val.Value], val.CampaignsID.Bytes)
			} else {
				TargetCache.ExcludeCountryIndex[val.Value] = append(TargetCache.ExcludeCountryIndex[val.Value], val.CampaignsID.Bytes)
			}
		case int32(model.TargetCategoryOS):
			TargetCache.IncludeOSIndex[val.Value] = append(TargetCache.IncludeOSIndex[val.Value], val.CampaignsID.Bytes)
		}
	}
	TargetCache.TargetMutex.Unlock()

	return TargetCache, nil
}

// ProcessRedisStreamDataService is a function for processing data from the Redis stream.
func ProcessRedisStreamDataService(ctx context.Context, tableName string, id string, isDeleted bool) error {

	conn := dbpkg.GetConn()
	if conn == nil {
		return errors.New("database connection is nil")
	}

	switch tableName {
	case string(dbpkg.CampaignsTable):
		if isDeleted {
			TargetCache.TargetMutex.Lock()
			delete(TargetCache.Campaigns, uuid.MustParse(id))
			TargetCache.TargetMutex.Unlock()
		} else {
			campaign, err := conn.GetCampaignByID(ctx, uuid.MustParse(id))
			if err != nil {
				return err
			}
			TargetCache.TargetMutex.Lock()
			TargetCache.Campaigns[campaign.ID.Bytes] = &model.Campaign{
				ID:               campaign.ID.Bytes,
				CampaignStringID: campaign.CampaignStringID,
				Name:             campaign.Name,
				ImageUrl:         campaign.ImageUrl,
				CTA:              campaign.Cta,
				ActivityStatus:   campaign.ActivityStatus,
				IsDeleted:        campaign.IsDeleted,
			}
			TargetCache.TargetMutex.Unlock()
		}
	case string(dbpkg.TargetingRulesTable):
		targetRule, err := conn.GetTargetRulesByID(ctx, uuid.MustParse(id))
		if err != nil {
			return err
		}
		TargetCache.TargetMutex.Lock()
		defer TargetCache.TargetMutex.Unlock()
		switch targetRule.Category {
		case int32(model.TargetCategoryAppID):
			TargetCache.IncludeAppIndex[targetRule.Value] = append(TargetCache.IncludeAppIndex[targetRule.Value], targetRule.CampaignsID.Bytes)
		case int32(model.TargetCategoryCountry):
			if targetRule.IsIncluded {
				TargetCache.IncludeCountryIndex[targetRule.Value] = append(TargetCache.IncludeCountryIndex[targetRule.Value], targetRule.CampaignsID.Bytes)
			} else {
				TargetCache.ExcludeCountryIndex[targetRule.Value] = append(TargetCache.ExcludeCountryIndex[targetRule.Value], targetRule.CampaignsID.Bytes)
			}
		case int32(model.TargetCategoryOS):
			TargetCache.IncludeOSIndex[targetRule.Value] = append(TargetCache.IncludeOSIndex[targetRule.Value], targetRule.CampaignsID.Bytes)
		}

		if isDeleted {
			// TODO: do the same check and remove the campaign from the cache
		}
	default:
		return errors.New("unknown table name")
	}

	return nil
}

// DeliveryService handles the delivery service request and returns the response based on the targeting rules
// I am trying to use inverted indexing. It checks the cache for the campaigns that match the request criteria and returns them.
// I am iterating over the cache to find the campaigns that match the request criteria
// I am using a map to ensure that each campaign is only returned once, even if it matches multiple criteria.
// the reason behind this approach of iterating is because it is clearly meantioned in the requirements
// that the number of campaings will be few thousands and the number of requests will be in millions
// so this approach is efficient enough to handle the load.
func DeliveryService(ctx context.Context, req *model.DeliveryServiceRequest) (res []*model.DeliveryServiceResponse, err error) {

	uniqueCampaigns := make(map[uuid.UUID]bool)
	TargetCache.TargetMutex.RLock()
	defer TargetCache.TargetMutex.RUnlock()
	// Check for AppID targeting
	if appIDs, ok := TargetCache.IncludeAppIndex[req.AppID]; ok {
		for _, campaignID := range appIDs {
			if _, exists := uniqueCampaigns[campaignID]; !exists {
				uniqueCampaigns[campaignID] = true
			}
		}
	}
	// Check for OS targeting
	if osIDs, ok := TargetCache.IncludeOSIndex[req.OS]; ok {
		for _, campaignID := range osIDs {
			if _, exists := uniqueCampaigns[campaignID]; !exists {
				uniqueCampaigns[campaignID] = true
			}
		}
	}
	// Check for Included Country targeting
	if includedCountries, ok := TargetCache.IncludeCountryIndex[req.Country]; ok {
		for _, campaignID := range includedCountries {
			if _, exists := uniqueCampaigns[campaignID]; !exists {
				uniqueCampaigns[campaignID] = true
			}
		}
	}

	// removing campaigns that are excluded based on country
	if excludedCountries, ok := TargetCache.ExcludeCountryIndex[req.Country]; ok {
		for _, campaignID := range excludedCountries {
			delete(uniqueCampaigns, campaignID)
		}
	}

	for campaignID := range uniqueCampaigns {
		campaign, exists := TargetCache.Campaigns[campaignID]
		if exists && campaign.ActivityStatus && !campaign.IsDeleted {
			res = append(res, &model.DeliveryServiceResponse{
				CampaignStringID: campaign.CampaignStringID,
				Image:            campaign.ImageUrl,
				Cta:              campaign.CTA,
			})
		}
	}

	return res, nil
}
