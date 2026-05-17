import { BullModule } from '@nestjs/bullmq';
import { Module } from '@nestjs/common';
import { IndexingController } from './indexing.controller';
import { IndexingProcessor } from './indexing.processor';
import { INDEXING_QUEUE, IndexingService } from './indexing.service';
@Module({
  imports: [BullModule.registerQueue({ name: INDEXING_QUEUE })],
  controllers: [IndexingController],
  providers: [IndexingService, IndexingProcessor],
})
export class IndexingModule {}
