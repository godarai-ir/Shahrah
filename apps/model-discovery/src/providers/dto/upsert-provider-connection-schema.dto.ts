import { ApiProperty } from '@nestjs/swagger';
import { IsArray, IsBoolean, IsInt, IsString, Min } from 'class-validator';

export class UpsertProviderConnectionSchemaDto {
  @ApiProperty()
  @IsString()
  providerSlug!: string;

  @ApiProperty()
  @IsString()
  authMode!: string;

  @ApiProperty({ minimum: 1 })
  @IsInt()
  @Min(1)
  version!: number;

  @ApiProperty()
  @IsBoolean()
  runtimeReady!: boolean;

  @ApiProperty({ type: 'array', items: { type: 'object' } })
  @IsArray()
  fields!: Record<string, unknown>[];
}
