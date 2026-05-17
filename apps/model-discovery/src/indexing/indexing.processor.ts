import { Processor, WorkerHost } from '@nestjs/bullmq';
import { Job } from 'bullmq';
import { INDEXING_QUEUE } from './indexing.service';
@Processor(INDEXING_QUEUE)
export class IndexingProcessor extends WorkerHost {
  process(job: Job<{ modelId: string }>) {
    return Promise.resolve({ processed: job.data.modelId });
  }
}
