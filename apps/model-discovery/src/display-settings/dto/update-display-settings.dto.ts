import { IsIn } from 'class-validator';

const currencyUnits = ['USD', 'IRR', 'TOMAN'] as const;
const dateCalendars = ['GREGORIAN', 'SHAMSI'] as const;
const priceScales = ['PER_MILLION_TOKENS', 'PER_TOKEN'] as const;

export class UpdateDisplaySettingsDto {
  @IsIn(currencyUnits)
  currencyUnit!: string;

  @IsIn(dateCalendars)
  dateCalendar!: string;

  @IsIn(priceScales)
  priceScale!: string;
}
