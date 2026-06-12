import { ApiProperty } from '@nestjs/swagger';
import { IsString } from 'class-validator';

export class UpsertModelSeriesDto {
  @ApiProperty()
  @IsString()
  slug!: string;

  @ApiProperty()
  @IsString()
  name!: string;
}
