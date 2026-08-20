-- +goose Up
-- +goose StatementBegin
ALTER TABLE chatbots
    ADD COLUMN IF NOT EXISTS default_locale text NOT NULL DEFAULT 'fa-IR',
    ADD COLUMN IF NOT EXISTS answer_language text NOT NULL DEFAULT 'query';

ALTER TABLE chatbots
    ADD CONSTRAINT chatbot_default_locale_check CHECK (default_locale IN ('fa-IR','en-US')),
    ADD CONSTRAINT chatbot_answer_language_check CHECK (answer_language IN ('query','fa','en'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chatbots
    DROP CONSTRAINT IF EXISTS chatbot_answer_language_check,
    DROP CONSTRAINT IF EXISTS chatbot_default_locale_check,
    DROP COLUMN IF EXISTS answer_language,
    DROP COLUMN IF EXISTS default_locale;
-- +goose StatementEnd
