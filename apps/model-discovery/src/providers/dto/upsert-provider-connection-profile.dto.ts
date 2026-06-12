import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { IsObject, IsOptional, IsString } from 'class-validator';

export class UpsertProviderConnectionProfileDto {
  @ApiProperty()
  @IsString()
  providerSlug!: string;

  @ApiProperty()
  @IsString()
  providerType!: string;

  @ApiPropertyOptional()
  @IsOptional()
  @IsString()
  authMode?: string;

  @ApiProperty({ additionalProperties: true })
  @IsObject()
  configJson!: Record<string, unknown>;
}
