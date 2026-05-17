import { BullModule } from '@nestjs/bullmq';
import { Module } from '@nestjs/common';
import { SyncJobsController } from './sync-jobs.controller';
import { SyncJobsProcessor } from './sync-jobs.processor';
import { SYNC_QUEUE, SyncJobsService } from './sync-jobs.service';
@Module({
  imports: [BullModule.registerQueue({ name: SYNC_QUEUE })],
  controllers: [SyncJobsController],
  providers: [SyncJobsService, SyncJobsProcessor],
})
export class SyncJobsModule {}
