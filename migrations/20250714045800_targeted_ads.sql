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
    is_deleted BOOLEAN;
BEGIN
    IF TG_OP = 'DELETE' THEN
        row_id := OLD.id::TEXT;
        is_deleted := TRUE;
    ELSE
        row_id := NEW.id::TEXT;
        is_deleted := FALSE;
    END IF;

    PERFORM pg_notify('table_changes', TG_TABLE_NAME || ':' || row_id || ':' || is_deleted::TEXT);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


INSERT INTO campaigns (id,campaign_string_id, name, image_url, CTA, activity_status, created_by, updated_by) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11','spotify', 'Spotify - Music for everyone', 'https://somelink', 'Download', TRUE, 'admin', 'admin'),
('b1cdef00-d1e2-4f33-aa44-7cc0cd191b22','duolingo', 'Duolingo: Best way to learn', 'https://somelink2', 'Install', TRUE, 'admin', 'admin'),
('c2def011-e2f3-4a55-bb66-8dd1de202c33','subwaysurfer', 'Subway Surfer', 'https://somelink3', 'Play', TRUE, 'admin', 'admin');
INSERT INTO campaigns (campaign_string_id, name, image_url, CTA, activity_status, created_by, updated_by) VALUES
('netflix', 'Netflix - Watch TV Shows Online', 'https://example.com/netflix_img.jpg', 'Watch Now', TRUE, 'admin', 'admin');
INSERT INTO campaigns (campaign_string_id, name, image_url, CTA, activity_status, created_by, updated_by) VALUES
('amazon_prime', 'Amazon Prime Video - Movies & TV', 'https://example.com/prime_img.png', 'Stream It', TRUE, 'admin', 'admin');
INSERT INTO campaigns (campaign_string_id, name, image_url, CTA, activity_status, created_by, updated_by) VALUES
('youtube_music', 'YouTube Music - Official Music', 'https://example.com/youtube_music_img.webp', 'Listen Now', TRUE, 'admin', 'admin');
INSERT INTO campaigns (campaign_string_id, name, image_url, CTA, activity_status, created_by, updated_by) VALUES
('Maps', 'Google Maps - Navigation & Transit', 'https://example.com/Maps_img.jpeg', 'Get Directions', TRUE, 'admin', 'admin');
INSERT INTO campaigns (campaign_string_id, name, image_url, CTA, activity_status, created_by, updated_by) VALUES
('whatsapp', 'WhatsApp - Simple. Secure. Reliable.', 'https://example.com/whatsapp_img.png', 'Chat Now', TRUE, 'admin', 'admin');


-- Rules for Spotify (assuming a campaign_id for Spotify)
INSERT INTO targeting_rules (campaigns_id, is_included, category, value, created_by, updated_by) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', TRUE, 2, 'US', 'admin', 'admin'); -- Include Country: US
INSERT INTO targeting_rules (campaigns_id, is_included, category, value, created_by, updated_by) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', TRUE, 2, 'Canada', 'admin', 'admin'); -- Include Country: Canada

-- Rules for Duolingo (assuming a campaign_id for Duolingo)
INSERT INTO targeting_rules (campaigns_id, is_included, category, value, created_by, updated_by) VALUES
('b1cdef00-d1e2-4f33-aa44-7cc0cd191b22', TRUE, 3, 'Android', 'admin', 'admin'); -- Include OS: Android
INSERT INTO targeting_rules (campaigns_id, is_included, category, value, created_by, updated_by) VALUES
('b1cdef00-d1e2-4f33-aa44-7cc0cd191b22', TRUE, 3, 'iOS', 'admin', 'admin'); -- Include OS: iOS
INSERT INTO targeting_rules (campaigns_id, is_included, category, value, created_by, updated_by) VALUES
('b1cdef00-d1e2-4f33-aa44-7cc0cd191b22', FALSE, 2, 'US', 'admin', 'admin'); -- Exclude Country: US

-- Rules for Subway Surfer (assuming a campaign_id for Subway Surfer)
INSERT INTO targeting_rules (campaigns_id, is_included, category, value, created_by, updated_by) VALUES
('c2def011-e2f3-4a55-bb66-8dd1de202c33', TRUE, 3, 'Android', 'admin', 'admin'); -- Include OS: Android
INSERT INTO targeting_rules (campaigns_id, is_included, category, value, created_by, updated_by) VALUES
('c2def011-e2f3-4a55-bb66-8dd1de202c33', TRUE, 1, 'com.gametion.ludokinggame', 'admin', 'admin'); -- Include App: com.gametion.ludokinggame

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
