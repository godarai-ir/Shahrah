import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';

export class RuntimePricingRateDto {
  @ApiProperty()
  dimension!: string;

  @ApiProperty()
  unit!: string;

  @ApiProperty({
    description: 'Exact decimal USD price serialized as a string.',
    example: '0.000003',
  })
  unitPriceUsd!: string;

  @ApiProperty()
  currency!: string;

  @ApiPropertyOptional({ type: Object })
  condition?: Record<string, unknown> | null;

  @ApiProperty()
  region!: string;

  @ApiProperty()
  effectiveFrom!: string;

  @ApiPropertyOptional()
  effectiveTo?: string | null;

  @ApiProperty()
  sourceUrl!: string;
}

export class RuntimeProfileRevisionDto {
  @ApiProperty()
  version!: number;

  @ApiProperty({ type: Object })
  profile!: Record<string, unknown>;

  @ApiPropertyOptional()
  sourceUrl?: string | null;

  @ApiProperty()
  effectiveFrom!: string;
}

export class RuntimeSnapshotModelDto {
  @ApiProperty()
  canonicalId!: string;

  @ApiProperty()
  slug!: string;

  @ApiProperty({ enum: ['active', 'preview', 'deprecated', 'unavailable'] })
  status!: string;

  @ApiPropertyOptional()
  deprecationDate?: string | null;

  @ApiPropertyOptional()
  retirementDate?: string | null;

  @ApiPropertyOptional()
  replacementCanonicalId?: string | null;

  @ApiProperty({ type: RuntimeProfileRevisionDto })
  runtimeProfile!: RuntimeProfileRevisionDto;

  @ApiProperty({ type: [RuntimePricingRateDto] })
  pricingRates!: RuntimePricingRateDto[];

  @ApiProperty({ type: [String] })
  sources!: string[];
}

export class RuntimeSnapshotDto {
  @ApiProperty()
  provider!: string;

  @ApiProperty({
    description:
      'SHA-256 revision of the ordered runtime and pricing payload.',
  })
  revision!: string;

  @ApiProperty()
  generatedAt!: string;

  @ApiProperty({ type: [RuntimeSnapshotModelDto] })
  models!: RuntimeSnapshotModelDto[];
}
