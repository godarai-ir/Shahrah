import { ApiProperty, ApiPropertyOptional } from '@nestjs/swagger';
import { IsOptional, IsString } from 'class-validator';
export class SearchModelsDto {
  @ApiProperty() @IsString() q!: string;
  @ApiPropertyOptional() @IsOptional() @IsString() filterBy?: string;
}
