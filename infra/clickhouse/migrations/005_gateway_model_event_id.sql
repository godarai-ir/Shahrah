ALTER TABLE gateway_model_events
    ADD COLUMN IF NOT EXISTS event_id String DEFAULT '' AFTER ts;
