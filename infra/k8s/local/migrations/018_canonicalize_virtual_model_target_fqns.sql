-- +goose Up
UPDATE public.virtual_model_targets
SET model = split_part(model, ':', 1) || ':' || split_part(model, ':', 2) || ':' ||
            substring(split_part(model, ':', 3) from length(split_part(model, ':', 1)) + 2)
WHERE array_length(string_to_array(model, ':'), 1) = 3
  AND lower(split_part(model, ':', 3)) LIKE lower(split_part(model, ':', 1)) || '/%';

-- +goose Down
-- Data canonicalization is intentionally not reversed.
