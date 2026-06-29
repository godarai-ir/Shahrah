import { Controller, Get, ServiceUnavailableException } from '@nestjs/common';
import { ApiTags } from '@nestjs/swagger';
import { PrismaService } from '../database/prisma.service';
import { IndexingService } from '../indexing/indexing.service';

@ApiTags('health')
@Controller({ version: '1' })
export class HealthController {
  constructor(
    private readonly prisma: PrismaService,
    private readonly indexing: IndexingService,
  ) {}

  @Get('health')
  health() {
    return {
      status: 'healthy',
      service: 'model-discovery',
    };
  }

  @Get('ready')
  async ready() {
    let typesenseAvailable = false;

    try {
      await this.prisma.$queryRaw`SELECT 1`;
    } catch {
      throw new ServiceUnavailableException({
        status: 'not_ready',
        service: 'model-discovery',
        database: 'unavailable',
      });
    }

    typesenseAvailable = await this.indexing.ping();
    if (!typesenseAvailable && this.indexing.isReadyStrict()) {
      throw new ServiceUnavailableException({
        status: 'not_ready',
        service: 'model-discovery',
        database: 'available',
        typesense: 'unavailable',
      });
    }

    return {
      status: 'ready',
      service: 'model-discovery',
      database: 'available',
      typesense: typesenseAvailable ? 'available' : 'degraded',
    };
  }
}
