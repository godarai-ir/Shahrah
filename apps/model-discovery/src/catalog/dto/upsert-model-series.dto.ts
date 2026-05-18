import { IsString } from 'class-validator';

export class UpsertModelSeriesDto {
  @IsString() slug!: string;
  @IsString() name!: string;
}
