import { IsBoolean, IsOptional, IsString } from 'class-validator';

export class UpsertProviderDto {
  @IsString() slug!: string;
  @IsString() name!: string;
  @IsOptional() @IsString() description?: string;
  @IsOptional() @IsString() websiteUrl?: string;
  @IsOptional() @IsString() docsUrl?: string;
  @IsOptional() @IsString() logoUrl?: string;
  @IsOptional() @IsString() status?: string;
  @IsOptional() @IsBoolean() enabled?: boolean;
  @IsOptional() @IsBoolean() supportsByok?: boolean;
}
