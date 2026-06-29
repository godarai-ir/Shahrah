import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import {
  IsArray,
  IsBoolean,
  IsInt,
  IsOptional,
  IsString,
} from 'class-validator';

export class UpsertModelPricingDto {
  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  currency?: string;

  @ApiPropertyOptional()
  inputPricePerMillionTokens?: number;

  @ApiPropertyOptional()
  outputPricePerMillionTokens?: number;

  @ApiPropertyOptional()
  cachedInputPricePerMillionTokens?: number;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  sourceUrl?: string;
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
  @IsBoolean()
  free?: boolean;

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

  @ApiPropertyOptional({ type: () => UpsertModelPricingDto })
  @IsOptional()
  pricing?: UpsertModelPricingDto;
}
