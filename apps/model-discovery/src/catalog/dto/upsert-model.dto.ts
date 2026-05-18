import {
  IsArray,
  IsBoolean,
  IsInt,
  IsOptional,
  IsString,
} from 'class-validator';

export class UpsertModelDto {
  @IsString() providerSlug!: string;
  @IsString() canonicalId!: string;
  @IsString() slug!: string;
  @IsString() displayName!: string;
  @IsOptional() @IsString() shortDescription?: string;
  @IsOptional() @IsString() longDescription?: string;
  @IsOptional() @IsString() status?: string;
  @IsOptional() @IsBoolean() enabled?: boolean;
  @IsOptional() @IsString() lifecycleStatus?: string;
  @IsOptional() @IsString() modelSeriesSlug?: string;
  @IsOptional() @IsString() categorySlug?: string;
  @IsOptional() @IsInt() contextWindow?: number;
  @IsOptional() @IsInt() maxOutputTokens?: number;
  @IsOptional() @IsBoolean() openSource?: boolean;
  @IsOptional() @IsBoolean() free?: boolean;
  @IsOptional() @IsArray() inputModalities?: string[];
  @IsOptional() @IsArray() outputModalities?: string[];
  @IsOptional() @IsArray() supportedParameters?: string[];
  @IsOptional() @IsArray() capabilities?: string[];
  @IsOptional() @IsArray() regions?: string[];
  @IsOptional() pricing?: {
    currency?: string;
    inputPricePerMillionTokens?: number;
    outputPricePerMillionTokens?: number;
    cachedInputPricePerMillionTokens?: number;
    sourceUrl?: string;
  };
}
