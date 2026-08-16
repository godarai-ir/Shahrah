import {
  activeModelPricingWhere,
  isPricingWindowActive,
  nextPricingTransition,
  sortModelsByActivePrice,
} from './pricing-window';

describe('pricing windows', () => {
  const start = new Date('2026-08-13T00:00:00Z');
  const end = new Date('2027-01-01T00:00:00Z');
  const window = { effectiveFrom: start, effectiveTo: end };

  it.each([
    ['before effectiveFrom', '2026-08-12T23:59:59.999Z', false],
    ['at effectiveFrom', '2026-08-13T00:00:00Z', true],
    ['before effectiveTo', '2026-12-31T23:59:59.999Z', true],
    ['at effectiveTo', '2027-01-01T00:00:00Z', false],
    ['after effectiveTo', '2027-01-01T00:00:00.001Z', false],
  ])('%s', (_name, asOf, expected) => {
    expect(isPricingWindowActive(window, new Date(asOf))).toBe(expected);
  });

  it('treats missing boundaries as unbounded', () => {
    expect(
      isPricingWindowActive(
        { effectiveFrom: null, effectiveTo: null },
        new Date('2026-08-17T00:00:00Z'),
      ),
    ).toBe(true);
  });

  it('finds the earliest future start or end boundary', () => {
    expect(
      nextPricingTransition(
        [
          window,
          {
            effectiveFrom: new Date('2027-01-01T00:00:00Z'),
            effectiveTo: null,
          },
          {
            effectiveFrom: new Date('2026-08-13T00:00:00Z'),
            effectiveTo: new Date('2026-10-01T00:00:00Z'),
          },
        ],
        new Date('2026-08-17T00:00:00Z'),
      ),
    ).toEqual(new Date('2026-10-01T00:00:00Z'));
  });

  it('builds a half-open Prisma predicate from one reference instant', () => {
    const asOf = new Date('2026-08-17T00:00:00Z');

    expect(activeModelPricingWhere(asOf)).toEqual({
      AND: [
        {
          OR: [{ effectiveFrom: null }, { effectiveFrom: { lte: asOf } }],
        },
        {
          OR: [{ effectiveTo: null }, { effectiveTo: { gt: asOf } }],
        },
      ],
    });
  });

  it('sorts globally by the active price and puts missing prices last', () => {
    const rows = [
      { displayName: 'No price', pricing: [] },
      {
        displayName: 'Future-expensive',
        pricing: [{ inputPricePerMillionTokens: 1.5 }],
      },
      {
        displayName: 'Current-cheap',
        pricing: [{ inputPricePerMillionTokens: 0.75 }],
      },
    ];

    expect(
      sortModelsByActivePrice(rows, 'inputPricePerMillionTokens').map(
        (row) => row.displayName,
      ),
    ).toEqual(['Current-cheap', 'Future-expensive', 'No price']);
  });
});
