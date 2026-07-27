import { ApiPropertyOptional } from '@nestjs/swagger';
import { Transform, Type } from 'class-transformer';
import {
  IsBoolean,
  IsIn,
  IsInt,
  IsNumber,
  IsOptional,
  IsString,
  Max,
  Min,
} from 'class-validator';

const toArray = (value: unknown): string[] => {
  if (!value) return [];
  if (Array.isArray(value)) return value.map(String);
  return String(value)
    .split(',')
    .map((v) => v.trim())
    .filter(Boolean);
};

export class ModelsQueryDto {
  @ApiPropertyOptional() @IsOptional() @IsString() q?: string;
  @ApiPropertyOptional() @IsOptional() @IsString() provider?: string;
  @ApiPropertyOptional() @IsOptional() @IsString() category?: string;
  @ApiPropertyOptional()
  @IsOptional()
  @Transform(({ value }) => toArray(value))
  mode?: string[];
  @ApiPropertyOptional()
  @IsOptional()
  @Transform(({ value }) => toArray(value))
  input_modalities?: string[];
  @ApiPropertyOptional()
  @IsOptional()
  @Transform(({ value }) => toArray(value))
  output_modalities?: string[];
  @ApiPropertyOptional()
  @IsOptional()
  @Transform(({ value }) => toArray(value))
  supported_parameters?: string[];
  @ApiPropertyOptional()
  @IsOptional()
  @Transform(({ value }) => toArray(value))
  capabilities?: string[];
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsInt()
  min_context_window?: number;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsInt()
  max_context_window?: number;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsInt()
  min_output_tokens?: number;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsInt()
  max_output_tokens?: number;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsNumber()
  @Min(0)
  min_price?: number;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsNumber()
  @Min(0)
  max_price?: number;
  @ApiPropertyOptional({
    enum: ['input', 'output', 'cache_read', 'cache_write'],
  })
  @IsOptional()
  @IsIn(['input', 'output', 'cache_read', 'cache_write'])
  price_usage?: string;
  @ApiPropertyOptional() @IsOptional() @IsString() price_unit?: string;
  @ApiPropertyOptional() @IsOptional() @IsString() price_region?: string;
  @ApiPropertyOptional() @IsOptional() @IsString() price_resolution?: string;
  @ApiPropertyOptional({ enum: ['standard', 'batch', 'priority'] })
  @IsOptional()
  @IsIn(['standard', 'batch', 'priority'])
  price_service_tier?: string;
  @ApiPropertyOptional()
  @Type(() => Boolean)
  @IsOptional()
  @IsBoolean()
  free_only?: boolean;
  @ApiPropertyOptional()
  @Type(() => Boolean)
  @IsOptional()
  @IsBoolean()
  open_source_only?: boolean;
  @ApiPropertyOptional() @IsOptional() @IsString() status?: string;
  @ApiPropertyOptional() @IsOptional() @IsString() region?: string;
  @ApiPropertyOptional({
    enum: [
      'newest',
      'name',
      'price_asc',
      'price_desc',
      'largest_context',
      'most_popular',
      'highest_benchmark_score',
    ],
  })
  @IsOptional()
  @IsIn([
    'newest',
    'name',
    'price_asc',
    'price_desc',
    'largest_context',
    'most_popular',
    'highest_benchmark_score',
  ])
  sort: string = 'newest';
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @Min(1)
  page: number = 1;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @Min(1)
  @Max(100)
  limit: number = 24;
}
