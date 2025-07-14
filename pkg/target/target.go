package target

import (
	"context"
	"errors"
	dbpkg "targetad/pkg/db"
	"targetad/pkg/target/model"

	"github.com/google/uuid"
)

var TargetCache model.TargetingData // decaring the cache globally

// fetches the data from pgsql db and initializes the cache
func InitCache(ctx context.Context) (TargetCache *model.TargetingData, err error) {
	conn := dbpkg.GetConn()
	if conn == nil {
		return nil, errors.New("database connection is nil")
	}

	campaigns, err := conn.ListAllValidCampaigns(ctx)
	if err != nil {
		return nil, err
	}
	TargetCache.TargetMutex.Lock()
	defer TargetCache.TargetMutex.Unlock()

	TargetCache.Campaigns = make(map[uuid.UUID]*model.Campaign)
	TargetCache.ExcludeCountryIndex = make(map[string][]uuid.UUID)
	TargetCache.IncludeCountryIndex = make(map[string][]uuid.UUID)
	TargetCache.IncludeOSIndex = make(map[string][]uuid.UUID)
	TargetCache.IncludeAppIndex = make(map[string][]uuid.UUID)

	for _, campaign := range campaigns {
		TargetCache.Campaigns[campaign.ID.Bytes] = &model.Campaign{
			ID:             campaign.ID.Bytes,
			Name:           campaign.Name,
			ImageUrl:       campaign.ImageUrl,
			CTA:            campaign.Cta,
			ActivityStatus: campaign.ActivityStatus,
			IsDeleted:      campaign.IsDeleted,
		}
	}

	// get all valid targetting rules from the database
	dbvals, err := conn.ListValidTargetingRules(ctx)
	if err != nil {
		return nil, err
	}

	for _, val := range dbvals {
		switch val.Category {
		case int32(model.TargetCategoryAppID):
			TargetCache.IncludeAppIndex[val.Value] = append(TargetCache.IncludeAppIndex[val.Value], val.CampaignID.Bytes)
		case int32(model.TargetCategoryCountry):
			if val.IsIncluded {
				TargetCache.IncludeCountryIndex[val.Value] = append(TargetCache.IncludeCountryIndex[val.Value], val.CampaignID.Bytes)
			} else {
				TargetCache.ExcludeCountryIndex[val.Value] = append(TargetCache.ExcludeCountryIndex[val.Value], val.CampaignID.Bytes)
			}
		case int32(model.TargetCategoryOS):
			TargetCache.IncludeOSIndex[val.Value] = append(TargetCache.IncludeOSIndex[val.Value], val.CampaignID.Bytes)
		}
	}

	return TargetCache, nil
}
