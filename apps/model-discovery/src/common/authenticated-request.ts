import type { Request } from 'express';

export type ServicePrincipal = {
  serviceAccountId: string;
  serviceAccountName: string;
  scopes: string[];
  actorId?: string;
  actorRole?: string;
  requestId?: string;
};

export type AuthenticatedRequest = Request & { user: ServicePrincipal };

export function auditPrincipal(principal: ServicePrincipal) {
  return {
    actorId: principal.actorId ?? null,
    actorRole: principal.actorRole ?? null,
    serviceAccountId: principal.serviceAccountId,
    requestId: principal.requestId ?? null,
  };
}

export function auditActor(request: AuthenticatedRequest) {
  return auditPrincipal(request.user);
}
