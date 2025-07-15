package dbpkg

import "github.com/jackc/pgx/pgtype"

type Campaign struct {
	ID               pgtype.UUID
	CampaignStringID string
	Name             string
	ImageUrl         string
	Cta              string
	ActivityStatus   bool
	CreatedAt        pgtype.Timestamp
	CreatedBy        string
	UpdatedAt        pgtype.Timestamp
	UpdatedBy        string
	IsDeleted        bool
}

type ListValidTargetingRulesRow struct {
	CampaignsID pgtype.UUID
	IsIncluded  bool
	Category    int32
	Value       string
}

type PgsqlTableName string

const (
	CampaignsTable      PgsqlTableName = "campaigns"
	TargetingRulesTable PgsqlTableName = "targeting_rules"
)
