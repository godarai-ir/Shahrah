import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { IsInt, IsOptional, IsString, Min } from 'class-validator';

export class CreateModelDto {
  @ApiProperty() @IsString() providerId!: string;
  @ApiProperty() @IsString() slug!: string;
  @ApiProperty() @IsString() name!: string;
  @ApiPropertyOptional() @IsOptional() @IsString() version?: string;
  @ApiPropertyOptional() @IsOptional() @IsString() status?:
    | 'DRAFT'
    | 'PUBLISHED'
    | 'DEPRECATED';
  @ApiPropertyOptional() @IsOptional() @IsString() modality?: string;
  @ApiPropertyOptional() @IsOptional() @IsInt() @Min(1) contextWindow?: number;
}
