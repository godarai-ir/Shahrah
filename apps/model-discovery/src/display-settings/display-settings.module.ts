import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { DatabaseModule } from '../database/database.module';
import { DisplaySettingsController } from './display-settings.controller';
import { DisplaySettingsService } from './display-settings.service';
import { ExchangeRateService } from './exchange-rate.service';
import { MockExchangeRateProvider } from './exchange-rate.provider';

@Module({
  imports: [DatabaseModule, ConfigModule],
  controllers: [DisplaySettingsController],
  providers: [
    DisplaySettingsService,
    ExchangeRateService,
    MockExchangeRateProvider,
  ],
  exports: [DisplaySettingsService, ExchangeRateService],
})
export class DisplaySettingsModule {}
