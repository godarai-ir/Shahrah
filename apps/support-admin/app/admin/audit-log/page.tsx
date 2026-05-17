import { apiGet } from '@/lib/api';

type AuditRow = { id: string; action: string; entityType: string; entityId: string; createdAt: string };

export default async function AdminAuditLogPage() {
  let rows: AuditRow[] = [];
  let error: string | null = null;
  try {
    rows = await apiGet<AuditRow[]>('/admin/models/audit-log');
  } catch (err) {
    error = err instanceof Error ? err.message : 'Failed to load audit log';
  }
  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Audit Log</h1>
      {error ? <p className="text-sm text-amber-300">{error}</p> : null}
      <div className="space-y-2">
        {rows.map((r) => (
          <div key={r.id} className="rounded border border-zinc-800 bg-zinc-900 p-3 text-sm">
            <div>{r.action} {r.entityType} {r.entityId}</div>
            <div className="text-xs text-zinc-500">{new Date(r.createdAt).toLocaleString()}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
