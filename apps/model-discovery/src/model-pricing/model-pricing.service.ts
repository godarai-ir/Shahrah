import { Injectable } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';
import { CreateModelPricingDto } from './dto/create-model-pricing.dto';
import { UpdateModelPricingDto } from './dto/update-model-pricing.dto';

@Injectable()
export class ModelPricingService {
  constructor(private readonly prisma: PrismaService) {}
  create(dto: CreateModelPricingDto) {
    return this.prisma.modelPricing.create({ data: dto });
  }
  findAll() {
    return this.prisma.modelPricing.findMany({ include: { model: true } });
  }
  update(id: string, dto: UpdateModelPricingDto) {
    return this.prisma.modelPricing.update({ where: { id }, data: dto });
  }
  remove(id: string) {
    return this.prisma.modelPricing.delete({ where: { id } });
  }
}
