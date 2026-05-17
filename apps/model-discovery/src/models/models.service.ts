import { Injectable } from '@nestjs/common';
import { Prisma } from '@prisma/client';
import { PrismaService } from '../prisma/prisma.service';
import { CreateModelDto } from './dto/create-model.dto';
import { QueryModelDto } from './dto/query-model.dto';
import { UpdateModelDto } from './dto/update-model.dto';

@Injectable()
export class ModelsService {
  constructor(private readonly prisma: PrismaService) {}
  create(dto: CreateModelDto) {
    return this.prisma.model.create({ data: dto });
  }
  findAll(q: QueryModelDto) {
    const where: Prisma.ModelWhereInput = {
      status: q.status as never,
      modality: q.modality,
      provider: q.provider ? { slug: q.provider } : undefined,
      contextWindow: { gte: q.contextWindowMin, lte: q.contextWindowMax },
      capabilities: q.capabilities?.length
        ? { some: { name: { in: q.capabilities } } }
        : undefined,
      pricing:
        q.minInputPrice || q.maxInputPrice
          ? {
              some: {
                inputPer1M: { gte: q.minInputPrice, lte: q.maxInputPrice },
              },
            }
          : undefined,
    };
    return this.prisma.model.findMany({
      where,
      include: { provider: true, pricing: true, capabilities: true },
      skip: (q.page - 1) * q.pageSize,
      take: q.pageSize,
      orderBy: { createdAt: 'desc' },
    });
  }
  findOne(id: string) {
    return this.prisma.model.findUnique({ where: { id } });
  }
  update(id: string, dto: UpdateModelDto) {
    return this.prisma.model.update({ where: { id }, data: dto });
  }
  publish(id: string) {
    return this.prisma.model.update({
      where: { id },
      data: { status: 'PUBLISHED' },
    });
  }
  deprecate(id: string) {
    return this.prisma.model.update({
      where: { id },
      data: { status: 'DEPRECATED' },
    });
  }
  remove(id: string) {
    return this.prisma.model.delete({ where: { id } });
  }
}
