import { IsObject, IsOptional, IsString } from 'class-validator';

export class UpsertProviderConnectionProfileDto {
  @IsString() providerSlug!: string;
  @IsString() providerType!: string;
  @IsOptional() @IsString() authMode?: string;
  @IsObject() configJson!: Record<string, unknown>;
}
