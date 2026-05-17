import { BullModule } from '@nestjs/bullmq';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { AuthModule } from './auth/auth.module';
import { CatalogModule } from './catalog/catalog.module';
import { DatabaseModule } from './database/database.module';
import { FacetsModule } from './facets/facets.module';
import { ImportModule } from './import/import.module';
import { IndexingModule } from './indexing/indexing.module';
import { ProvidersModule } from './providers/providers.module';

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
        },
      }),
    }),
    DatabaseModule,
    AuthModule,
    CatalogModule,
    ProvidersModule,
    FacetsModule,
    IndexingModule,
    ImportModule,
  ],
})
export class AppModule {}
