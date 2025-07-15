-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS targeting_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaigns_id uuid NOT NULL, -- primary key from campaigns table
    is_included BOOLEAN NOT NULL, -- if false then exclude
    category INTEGER NOT NULL, -- category 1,2,3,4,5 1 for appID, 2 for Country, 3 for OS.
    value TEXT NOT NULL, -- value for the corresponding category, e.g. appID, country code, OS name
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

create TABLE IF NOT EXISTS campaigns (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_string_id TEXT NOT NULL,
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


CREATE OR REPLACE FUNCTION notify_change_with_id() RETURNS TRIGGER AS $$
DECLARE
    row_id TEXT;
BEGIN
    IF TG_OP = 'DELETE' THEN
        row_id := OLD.id::TEXT;
    ELSE
        row_id := NEW.id::TEXT;
    END IF;

    PERFORM pg_notify('table_changes', TG_TABLE_NAME || ':' || row_id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- i am creating triggers which will trigger our function when some CUD opperation is performed on the table
CREATE TRIGGER targeting_rules_change_notify
AFTER INSERT OR UPDATE OR DELETE ON targeting_rules
FOR EACH ROW
EXECUTE FUNCTION notify_change_with_id();

CREATE TRIGGER campaigns_change_notify
AFTER INSERT OR UPDATE OR DELETE ON campaigns
FOR EACH ROW
EXECUTE FUNCTION notify_change_with_id();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists targeting_rules;
drop table if exists campaigns;
drop function if exists notify_change_with_id;
drop trigger if exists targeting_rules_change_notify on targeting_rules;
drop trigger if exists campaigns_change_notify on campaigns;
-- +goose StatementEnd
