import { ApiPropertyOptional } from '@nestjs/swagger';
import { Transform, Type } from 'class-transformer';
import {
  IsArray,
  IsNumber,
  IsOptional,
  IsString,
  Max,
  Min,
} from 'class-validator';

export class QueryModelDto {
  @ApiPropertyOptional() @IsOptional() @IsString() provider?: string;
  @ApiPropertyOptional() @IsOptional() @IsString() status?: string;
  @ApiPropertyOptional({ type: [String] })
  @IsOptional()
  @Transform(({ value }): string[] => {
    if (value === undefined || value === null) return [];
    if (Array.isArray(value)) return value.map(String);
    return [String(value)];
  })
  @IsArray()
  @IsString({ each: true })
  capabilities?: string[];
  @ApiPropertyOptional() @IsOptional() @IsString() modality?: string;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsNumber()
  contextWindowMin?: number;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsNumber()
  contextWindowMax?: number;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsNumber()
  minInputPrice?: number;
  @ApiPropertyOptional()
  @Type(() => Number)
  @IsOptional()
  @IsNumber()
  maxInputPrice?: number;
  @ApiPropertyOptional({ default: 1 })
  @Type(() => Number)
  @IsOptional()
  @Min(1)
  page = 1;
  @ApiPropertyOptional({ default: 20 })
  @Type(() => Number)
  @IsOptional()
  @Min(1)
  @Max(100)
  pageSize = 20;
  @ApiPropertyOptional({ default: 'createdAt:desc' })
  @IsOptional()
  @IsString()
  sort = 'createdAt:desc';
}
