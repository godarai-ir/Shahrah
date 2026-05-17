import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { IsNumber, IsOptional, IsString, Min } from 'class-validator';

export class CreateModelPricingDto {
  @ApiProperty() @IsString() modelId!: string;
  @ApiProperty() @IsString() currency!: string;
  @ApiProperty() @IsNumber() @Min(0) inputPer1M!: number;
  @ApiProperty() @IsNumber() @Min(0) outputPer1M!: number;
  @ApiPropertyOptional()
  @IsOptional()
  @IsNumber()
  @Min(0)
  cacheReadPer1M?: number;
}
