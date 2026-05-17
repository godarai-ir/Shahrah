import { PartialType } from '@nestjs/swagger';
import { CreateModelPricingDto } from './create-model-pricing.dto';
export class UpdateModelPricingDto extends PartialType(CreateModelPricingDto) {}
