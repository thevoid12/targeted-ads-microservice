package dbpkg

import (
	"context"

	"github.com/jackc/pgx/pgtype"
)

const getCampaignByID = `-- name: GetCampaignByID :one
SELECT id, campaign_string_id, name, image_url, cta, activity_status, created_at, created_by, updated_at, updated_by, is_deleted
FROM campaigns
WHERE id= $1 AND is_deleted = false
`

func (conn *Dbconn) GetCampaignByID(ctx context.Context, id pgtype.UUID) (Campaign, error) {
	row := conn.Db.QueryRow(ctx, getCampaignByID, id)
	var i Campaign
	err := row.Scan(
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
	)
	return i, err
}

const getTargetRulesByID = `-- name: GetTargetRulesByID :one
SELECT id, campaigns_id, is_included, category, value, created_at, created_by, updated_at, updated_by, is_deleted
FROM targeting_rules
WHERE id= $1 AND is_deleted = false
`

func (conn *Dbconn) GetTargetRulesByID(ctx context.Context, id pgtype.UUID) (TargetingRule, error) {
	row := conn.Db.QueryRow(ctx, getTargetRulesByID, id)
	var i TargetingRule
	err := row.Scan(
		&i.ID,
		&i.CampaignsID,
		&i.IsIncluded,
		&i.Category,
		&i.Value,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
		&i.IsDeleted,
	)
	return i, err
}

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
