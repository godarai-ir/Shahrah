import { Injectable } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';
import { CreateProviderDto } from './dto/create-provider.dto';
import { UpdateProviderDto } from './dto/update-provider.dto';

@Injectable()
export class ProvidersService {
  constructor(private readonly prisma: PrismaService) {}
  create(dto: CreateProviderDto) {
    return this.prisma.provider.create({ data: dto });
  }
  findAll() {
    return this.prisma.provider.findMany();
  }
  update(id: string, dto: UpdateProviderDto) {
    return this.prisma.provider.update({ where: { id }, data: dto });
  }
  remove(id: string) {
    return this.prisma.provider.delete({ where: { id } });
  }
}
