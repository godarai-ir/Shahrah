import { Injectable } from '@nestjs/common';
import { PrismaService } from '../database/prisma.service';
import { UpdateDisplaySettingsDto } from './dto/update-display-settings.dto';
import { DisplaySettings, DisplaySettingsResponse } from './display-settings.types';
import { ExchangeRateService } from './exchange-rate.service';

const DEFAULT_SETTINGS: DisplaySettings = {
  currencyUnit: 'USD',
  dateCalendar: 'GREGORIAN',
  priceScale: 'PER_MILLION_TOKENS',
};

@Injectable()
export class DisplaySettingsService {
  constructor(
    private readonly prisma: PrismaService,
    private readonly exchangeRate: ExchangeRateService,
  ) {}

  async getDisplaySettings(): Promise<DisplaySettingsResponse> {
    const [settings, rate] = await Promise.all([
      this.ensureSettings(),
      this.exchangeRate.getCachedOrFetch(),
    ]);

    return {
      currencyUnit: settings.currencyUnit as DisplaySettings['currencyUnit'],
      dateCalendar: settings.dateCalendar as DisplaySettings['dateCalendar'],
      priceScale: settings.priceScale as DisplaySettings['priceScale'],
      exchangeRate: rate,
    };
  }

  async updateDisplaySettings(dto: UpdateDisplaySettingsDto): Promise<DisplaySettingsResponse> {
    await this.prisma.displaySetting.upsert({
      where: { id: 'default' },
      create: {
        id: 'default',
        ...dto,
      },
      update: dto,
    });

    return this.getDisplaySettings();
  }

  private async ensureSettings() {
    return this.prisma.displaySetting.upsert({
      where: { id: 'default' },
      create: {
        id: 'default',
        ...DEFAULT_SETTINGS,
      },
      update: {},
    });
  }
}
