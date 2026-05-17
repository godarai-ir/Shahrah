import { BullModule } from '@nestjs/bullmq';
import { Module } from '@nestjs/common';
import { ConfigModule, ConfigService } from '@nestjs/config';
import { AdminModule } from './admin/admin.module';
import { AppController } from './app.controller';
import { AuthModule } from './auth/auth.module';
import { CommonModule } from './common/common.module';
import { IndexingModule } from './indexing/indexing.module';
import { ModelBenchmarksModule } from './model-benchmarks/model-benchmarks.module';
import { ModelCapabilitiesModule } from './model-capabilities/model-capabilities.module';
import { ModelPricingModule } from './model-pricing/model-pricing.module';
import { ModelsModule } from './models/models.module';
import { PrismaModule } from './prisma/prisma.module';
import { ProvidersModule } from './providers/providers.module';
import { RedisModule } from './redis/redis.module';
import { SearchModule } from './search/search.module';
import { SyncJobsModule } from './sync-jobs/sync-jobs.module';
import { TypesenseModule } from './typesense/typesense.module';

@Module({
  imports: [
    ConfigModule.forRoot({ isGlobal: true }),
    BullModule.forRootAsync({
      inject: [ConfigService],
      useFactory: (cfg: ConfigService) => ({
        connection: {
          host: cfg.get<string>('REDIS_HOST', '127.0.0.1'),
          port: Number(cfg.get('REDIS_PORT', 6379)),
          password: cfg.get<string>('REDIS_PASSWORD') || undefined,
        },
      }),
    }),
    CommonModule,
    PrismaModule,
    RedisModule,
    TypesenseModule,
    AuthModule,
    ProvidersModule,
    ModelsModule,
    ModelPricingModule,
    ModelCapabilitiesModule,
    ModelBenchmarksModule,
    SearchModule,
    IndexingModule,
    SyncJobsModule,
    AdminModule,
  ],
  controllers: [AppController],
})
export class AppModule {}
