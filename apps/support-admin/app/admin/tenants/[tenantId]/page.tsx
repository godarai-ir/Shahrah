import { TenantDetailWorkspace } from './tenant-detail-workspace';

export default async function TenantDetailPage({ params }: { params: Promise<{ tenantId: string }> }) {
  const { tenantId } = await params;
  return <TenantDetailWorkspace tenantId={tenantId} />;
}
