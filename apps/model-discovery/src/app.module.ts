import { BullModule } from '@nestjs/bullmq';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
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
        connection: {
          host: cfg.get<string>('REDIS_HOST', '127.0.0.1'),
          port: Number(cfg.get<number>('REDIS_PORT', 6379)),
          password: cfg.get<string>('REDIS_PASSWORD') || undefined,
          // BullMQ workers are long-lived and must survive transient DNS or
          // Redis outages instead of exhausting ioredis' default 20 retries.
          maxRetriesPerRequest: null,
          connectTimeout: 10_000,
          retryStrategy: (attempt: number) => Math.min(attempt * 200, 5_000),
        },
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
