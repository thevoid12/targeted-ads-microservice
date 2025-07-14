-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS targeting_rules(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id uuid NOT NULL,
    is_included BOOLEAN NOT NULL, -- if false then exclude
    category INTEGER NOT NULL, -- category 1,2,3,4,5 1 for appID, 2 for Country, 3 for OS.
    value TEXT NOT NULL, -- value for the corresponding category, e.g. appID, country code, OS name
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

create TABLE IF NOT EXISTS campaigns(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    image_url TEXT NOT NULL,
    CTA TEXT NOT NULL, -- Call to Action text
    activity_status BOOLEAN NOT NULL DEFAULT TRUE, -- true for active, false for inactive
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists targeting_rules;
drop table if exists campaigns;
-- +goose StatementEnd
