import { Module } from '@nestjs/common';
import { IndexingModule } from '../indexing/indexing.module';
import { HealthController } from './health.controller';

@Module({ imports: [IndexingModule], controllers: [HealthController] })
export class HealthModule {}
