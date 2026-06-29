import {
  CanActivate,
  ExecutionContext,
  Injectable,
  UnauthorizedException,
} from '@nestjs/common';
import { createHash, timingSafeEqual } from 'crypto';
import type { Request } from 'express';
import { PrismaService } from '../../database/prisma.service';
import type { ServicePrincipal } from '../authenticated-request';

const TOKEN_PATTERN = /^sa_([A-Za-z0-9_-]{8,64})\.([A-Za-z0-9_-]{32,256})$/;
const ACTOR_PATTERN = /^[A-Za-z0-9@._:-]{1,128}$/;

@Injectable()
export class ServiceAccountGuard implements CanActivate {
  constructor(private readonly prisma: PrismaService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context
      .switchToHttp()
      .getRequest<Request & { user?: ServicePrincipal }>();
    const authorization = request.headers.authorization;
    const token = authorization?.startsWith('Bearer ')
      ? authorization.slice(7)
      : '';
    const match = TOKEN_PATTERN.exec(token);
    if (!match) throw new UnauthorizedException('Invalid service credential');

    const credential = await this.prisma.serviceAccountCredential.findUnique({
      where: { id: match[1] },
      include: { serviceAccount: true },
    });
    const now = new Date();
    if (
      !credential ||
      credential.revokedAt ||
      (credential.expiresAt && credential.expiresAt <= now) ||
      !credential.serviceAccount.active ||
      !this.secretMatches(match[2], credential.secretHash)
    ) {
      throw new UnauthorizedException('Invalid service credential');
    }

    const actorId = this.validHeader(request.headers['x-shahrah-actor-id']);
    const actorRole = this.validHeader(request.headers['x-shahrah-actor-role']);
    const requestId = this.validHeader(request.headers['x-request-id']);
    request.user = {
      serviceAccountId: credential.serviceAccount.id,
      serviceAccountName: credential.serviceAccount.name,
      scopes: credential.serviceAccount.scopes,
      ...(actorId ? { actorId } : {}),
      ...(actorRole ? { actorRole } : {}),
      ...(requestId ? { requestId } : {}),
    };

    if (
      !credential.lastUsedAt ||
      now.getTime() - credential.lastUsedAt.getTime() > 300_000
    ) {
      await this.prisma.serviceAccountCredential.update({
        where: { id: credential.id },
        data: { lastUsedAt: now },
      });
    }
    return true;
  }

  private secretMatches(secret: string, expectedHex: string): boolean {
    const actual = createHash('sha256').update(secret).digest();
    const expected = Buffer.from(expectedHex, 'hex');
    return (
      expected.length === actual.length && timingSafeEqual(actual, expected)
    );
  }

  private validHeader(
    value: string | string[] | undefined,
  ): string | undefined {
    const normalized = Array.isArray(value) ? value[0] : value;
    return normalized && ACTOR_PATTERN.test(normalized)
      ? normalized
      : undefined;
  }
}
