import { ApiProperty } from '@nestjs/swagger';
import { IsIn } from 'class-validator';

const currencyUnits = ['USD', 'IRR', 'TOMAN'] as const;
const dateCalendars = ['GREGORIAN', 'SHAMSI'] as const;
const priceScales = ['PER_MILLION_TOKENS', 'PER_TOKEN'] as const;

export class UpdateDisplaySettingsDto {
  @ApiProperty({ enum: currencyUnits })
  @IsIn(currencyUnits)
  currencyUnit!: string;

  @ApiProperty({ enum: dateCalendars })
  @IsIn(dateCalendars)
  dateCalendar!: string;

  @ApiProperty({ enum: priceScales })
  @IsIn(priceScales)
  priceScale!: string;
}
