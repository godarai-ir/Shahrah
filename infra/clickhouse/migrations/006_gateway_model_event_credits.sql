ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS credits_mc Int64 DEFAULT 0;
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS credits_avoided_mc Int64 DEFAULT 0;
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS credit_product LowCardinality(String) DEFAULT '';
ALTER TABLE gateway_model_events ADD COLUMN IF NOT EXISTS rate_card_version UInt32 DEFAULT 0;
