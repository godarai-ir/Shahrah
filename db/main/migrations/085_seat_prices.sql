-- +goose Up
-- Support sets each paid seat's price; the credit-list-price formula only
-- suggests one. NULL keeps the suggested price.
ALTER TABLE public.seat_types
    ADD COLUMN IF NOT EXISTS monthly_price_irr bigint,
    ADD COLUMN IF NOT EXISTS yearly_price_irr bigint,
    ADD CONSTRAINT chk_seat_types_monthly_price CHECK (monthly_price_irr IS NULL OR monthly_price_irr > 0),
    ADD CONSTRAINT chk_seat_types_yearly_price CHECK (yearly_price_irr IS NULL OR yearly_price_irr > 0);

-- +goose Down
ALTER TABLE public.seat_types
    DROP CONSTRAINT IF EXISTS chk_seat_types_yearly_price,
    DROP CONSTRAINT IF EXISTS chk_seat_types_monthly_price,
    DROP COLUMN IF EXISTS yearly_price_irr,
    DROP COLUMN IF EXISTS monthly_price_irr;
