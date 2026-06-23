import { Global, Module } from '@nestjs/common';
import { ScopesGuard } from '../common/guards/scopes.guard';
import { ServiceAccountGuard } from '../common/guards/service-account.guard';

@Global()
@Module({
  providers: [ServiceAccountGuard, ScopesGuard],
  exports: [ServiceAccountGuard, ScopesGuard],
})
export class AuthModule {}
