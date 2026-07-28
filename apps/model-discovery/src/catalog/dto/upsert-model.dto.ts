import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import {
  IsArray,
  IsBoolean,
  IsObject,
  IsInt,
  IsIn,
  IsISO8601,
  Matches,
  IsOptional,
  IsString,
  Min,
  ValidateNested,
} from 'class-validator';
import { Type } from 'class-transformer';

export class UpsertModelPricingTierDto {
  @ApiPropertyOptional()
  @IsOptional()
  @IsInt()
  @Min(0)
  from?: number | null;

  @ApiPropertyOptional()
  @IsOptional()
  @IsInt()
  @Min(0)
  to?: number | null;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  mode?: string | null;
}

export class UpsertModelPricingRateDto {
  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  key?: string;

  @ApiProperty()
  @IsIn(['input', 'output', 'cache_read', 'cache_write'])
  usage!: string;

  @ApiProperty()
  @IsString()
  unit!: string;

  @ApiProperty()
  @Matches(/^\d+(?:\.\d+)?$/)
  amount!: string;

  @ApiProperty()
  @IsInt()
  @Min(1)
  quantity!: number;

  @ApiPropertyOptional()
  @IsOptional()
  @IsIn(['standard', 'batch', 'priority'])
  serviceTier?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  resolution?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsInt()
  @Min(0)
  tierFrom?: number;

  @ApiPropertyOptional()
  @IsOptional()
  @IsInt()
  @Min(0)
  tierTo?: number;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  tierMode?: string;

  @ApiPropertyOptional({ type: () => UpsertModelPricingTierDto })
  @IsOptional()
  @ValidateNested()
  @Type(() => UpsertModelPricingTierDto)
  tier?: UpsertModelPricingTierDto | null;

  @ApiPropertyOptional()
  @IsOptional()
  @IsObject()
  qualifiers?: Record<string, string | number | boolean>;
}

export class UpsertModelPricingProfileDto {
  @ApiProperty()
  @IsString()
  region!: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  currency?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  sourceUrl?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsISO8601()
  effectiveFrom?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsISO8601()
  effectiveTo?: string;

  @ApiProperty({ type: () => [UpsertModelPricingRateDto] })
  @IsArray()
  @ValidateNested({ each: true })
  @Type(() => UpsertModelPricingRateDto)
  rates!: UpsertModelPricingRateDto[];
}

export class UpsertModelDto {
  @ApiProperty()
  @IsString()
  providerSlug!: string;

  @ApiProperty()
  @IsString()
  canonicalId!: string;

  @ApiProperty()
  @IsString()
  slug!: string;

  @ApiProperty()
  @IsString()
  displayName!: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  shortDescription?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  longDescription?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  status?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsBoolean()
  enabled?: boolean;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  lifecycleStatus?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  modelSeriesSlug?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  categorySlug?: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsInt()
  contextWindow?: number;

  @ApiPropertyOptional()
  @IsOptional()
  @IsInt()
  maxOutputTokens?: number;

  @ApiPropertyOptional()
  @IsOptional()
  @IsBoolean()
  openSource?: boolean;

  @ApiPropertyOptional()
  @IsOptional()
  @IsIn(['unknown', 'priced', 'free'])
  pricingStatus?: string;

  @ApiPropertyOptional({ type: () => [String] })
  @IsOptional()
  @IsArray()
  mode?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  @IsOptional()
  @IsArray()
  inputModalities?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  @IsOptional()
  @IsArray()
  outputModalities?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  @IsOptional()
  @IsArray()
  supportedParameters?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  @IsOptional()
  @IsArray()
  capabilities?: string[];

  @ApiPropertyOptional({ type: () => [String] })
  @IsOptional()
  @IsArray()
  regions?: string[];

  @ApiPropertyOptional({ type: () => [UpsertModelPricingProfileDto] })
  @IsOptional()
  @IsArray()
  @ValidateNested({ each: true })
  @Type(() => UpsertModelPricingProfileDto)
  pricingProfiles?: UpsertModelPricingProfileDto[];
}

export class UpdateModelPricingDto {
  @ApiPropertyOptional()
  @IsOptional()
  @IsIn(['unknown', 'priced', 'free'])
  pricingStatus?: string;

  @ApiPropertyOptional({ type: () => [UpsertModelPricingProfileDto] })
  @IsOptional()
  @IsArray()
  @ValidateNested({ each: true })
  @Type(() => UpsertModelPricingProfileDto)
  pricingProfiles?: UpsertModelPricingProfileDto[];
}

export function resolvePricingTier(rate: UpsertModelPricingRateDto): {
  from?: number;
  to?: number;
  mode?: string;
} {
  if (rate.tier) {
    return {
      from: rate.tier.from ?? undefined,
      to: rate.tier.to ?? undefined,
      mode: rate.tier.mode ?? undefined,
    };
  }
  return {
    from: rate.tierFrom,
    to: rate.tierTo,
    mode: rate.tierMode,
  };
}
