import { Processor, WorkerHost } from '@nestjs/bullmq';
import { Job } from 'bullmq';
import { SYNC_QUEUE } from './sync-jobs.service';
@Processor(SYNC_QUEUE)
export class SyncJobsProcessor extends WorkerHost {
  process(job: Job<{ providerSlug: string }>) {
    return Promise.resolve({ processed: job.data.providerSlug });
  }
}
