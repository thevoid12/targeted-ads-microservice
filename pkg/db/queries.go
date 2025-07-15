package dbpkg

import (
	"context"
)

const listAllValidCampaigns = `-- name: ListAllValidCampaigns :many
SELECT id, campaign_string_id, name, image_url, cta, activity_status, created_at, created_by, updated_at, updated_by, is_deleted
FROM campaigns
WHERE is_deleted = false
`

func (conn *Dbconn) ListAllValidCampaigns(ctx context.Context) ([]Campaign, error) {
	rows, err := conn.Db.Query(ctx, listAllValidCampaigns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Campaign
	for rows.Next() {
		var i Campaign
		if err := rows.Scan(
			&i.ID,
			&i.CampaignStringID,
			&i.Name,
			&i.ImageUrl,
			&i.Cta,
			&i.ActivityStatus,
			&i.CreatedAt,
			&i.CreatedBy,
			&i.UpdatedAt,
			&i.UpdatedBy,
			&i.IsDeleted,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listValidTargetingRules = `-- name: ListValidTargetingRules :many
SELECT campaigns_id, is_included, category, value
FROM targeting_rules
WHERE is_deleted = false
`

func (conn *Dbconn) ListValidTargetingRules(ctx context.Context) ([]ListValidTargetingRulesRow, error) {
	rows, err := conn.Db.Query(ctx, listValidTargetingRules)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListValidTargetingRulesRow
	for rows.Next() {
		var i ListValidTargetingRulesRow
		if err := rows.Scan(
			&i.CampaignsID,
			&i.IsIncluded,
			&i.Category,
			&i.Value,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
