import { InjectQueue } from '@nestjs/bullmq';
import { Injectable } from '@nestjs/common';
import { Queue } from 'bullmq';
export const SYNC_QUEUE = 'sync-jobs';
@Injectable()
export class SyncJobsService {
  constructor(@InjectQueue(SYNC_QUEUE) private readonly queue: Queue) {}
  enqueue(providerSlug: string) {
    return this.queue.add('sync-provider-metadata', { providerSlug });
  }
}
