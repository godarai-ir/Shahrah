import { BullModule } from '@nestjs/bullmq';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { buildRedisConnectionOptions } from './common/redis-connection';
import { AuthModule } from './auth/auth.module';
import { CatalogModule } from './catalog/catalog.module';
import { DatabaseModule } from './database/database.module';
import { FacetsModule } from './facets/facets.module';
import { HealthModule } from './health/health.module';
import { ImportModule } from './import/import.module';
import { IndexingModule } from './indexing/indexing.module';
import { ProvidersModule } from './providers/providers.module';
import { DisplaySettingsModule } from './display-settings/display-settings.module';

@Module({
  imports: [
    ConfigModule.forRoot({ isGlobal: true }),
    BullModule.forRootAsync({
      inject: [ConfigService],
      useFactory: (cfg: ConfigService) => ({
        connection: buildRedisConnectionOptions(cfg),
      }),
    }),
    DatabaseModule,
    AuthModule,
    CatalogModule,
    ProvidersModule,
    DisplaySettingsModule,
    FacetsModule,
    HealthModule,
    IndexingModule,
    ImportModule,
  ],
})
export class AppModule {}
