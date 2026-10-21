'use client';

import { useCallback, useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { type BankTransferItem, type BankTransferList, kindLabels, receiptHref } from '@/lib/support-bank-transfers';
import { controlPlaneRequest, formatDate, formatIRR } from '@/lib/support-tenants';

const PAGE_SIZE = 25;
const statuses = [['pending_review', 'Pending review'], ['approved', 'Approved'], ['rejected', 'Rejected'], ['all', 'All']];

/** Review queue for bank transfer receipts: view the receipt, then approve or reject it. */
export function BankTransfersWorkspace() {
  const [status, setStatus] = useState('pending_review');
  const [query, setQuery] = useState('');
  const [page, setPage] = useState(0);
  const [list, setList] = useState<BankTransferList>({ data: [], total: 0 });
  const [selected, setSelected] = useState<BankTransferItem | null>(null);
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    const params = new URLSearchParams({ status, limit: String(PAGE_SIZE), offset: String(page * PAGE_SIZE) });
    if (query.trim()) params.set('q', query.trim());
    try {
      setList(await controlPlaneRequest<BankTransferList>(`bank-transfers?${params}`));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to load bank transfers');
    } finally {
      setLoading(false);
    }
  }, [status, query, page]);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => { void load(); }, [load]);
  /* eslint-enable react-hooks/set-state-in-effect */

  function select(item: BankTransferItem) {
    setSelected(item);
    setReason('');
    setMessage('');
  }

  async function review(action: 'approve' | 'reject') {
    if (!selected || busy) return;
    if (action === 'reject' && !reason.trim()) { setError('Enter the rejection reason.'); return; }
    const verb = action === 'approve' ? 'Approve' : 'Reject';
    if (!window.confirm(`${verb} ${formatIRR(selected.amount_irr)} from ${selected.tenant_slug} (${kindLabels[selected.kind]})?`)) return;
    setBusy(true);
    setError('');
    setMessage('');
    try {
      await controlPlaneRequest(`tenants/${encodeURIComponent(selected.tenant_id)}/wallet/bank-transfers/${encodeURIComponent(selected.id)}/${action}`, {
        method: 'POST',
        // A stable key per transfer makes a double click or retry a no-op instead of a second credit.
        headers: { 'idempotency-key': `bank-transfer-${action}-${selected.id}` },
        body: JSON.stringify(action === 'reject' ? { reason: reason.trim() } : { idempotency_key: `bank_transfer:${selected.id}` }),
      });
      setMessage(`Transfer ${action === 'approve' ? 'approved' : 'rejected'}.`);
      setSelected(null);
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `${verb} failed`);
    } finally {
      setBusy(false);
    }
  }

  const receipt = selected ? receiptHref(selected) : undefined;
  const pages = Math.max(1, Math.ceil(list.total / PAGE_SIZE));

  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-semibold">Bank transfers</h1>
        <p className="mt-1 text-sm text-zinc-400">Check the receipt against the amount and tracking number, then approve or reject. Approval credits the wallet, credit pack or seat order it pays.</p>
      </div>
      {error ? <p role="alert" className="text-sm text-red-300">{error}</p> : null}
      {message ? <p role="status" className="text-sm text-emerald-300">{message}</p> : null}
      <div className="flex flex-wrap items-center gap-2">
        <select aria-label="Status" className="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm" value={status} onChange={(e) => { setStatus(e.target.value); setPage(0); }}>
          {statuses.map(([value, label]) => <option key={value} value={value}>{label}</option>)}
        </select>
        <Input aria-label="Search tenant or tracking number" className="w-64" placeholder="Tenant or tracking number" value={query} onChange={(e) => { setQuery(e.target.value); setPage(0); }} />
      </div>
      <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
        <Card className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="text-zinc-400"><tr><th className="py-2">Tenant</th><th>Type</th><th>Amount</th><th>Tracking</th><th>Submitted</th><th>Status</th></tr></thead>
            <tbody className="divide-y divide-zinc-800">
              {list.data.map((item) => (
                <tr key={item.id} className={item.id === selected?.id ? 'bg-zinc-800/60' : 'cursor-pointer hover:bg-zinc-800/40'} onClick={() => select(item)}>
                  <td className="py-3"><button type="button" className="text-cyan-300 hover:underline" onClick={() => select(item)}>{item.tenant_slug}</button></td>
                  <td>{kindLabels[item.kind]}</td>
                  <td>{formatIRR(item.amount_irr)}</td>
                  <td>{item.tracking_number}</td>
                  <td>{formatDate(item.created_at)}</td>
                  <td>{item.status.replace('_', ' ')}</td>
                </tr>
              ))}
              {!loading && list.data.length === 0 ? <tr><td colSpan={6} className="py-6 text-center text-zinc-400">No bank transfers.</td></tr> : null}
            </tbody>
          </table>
          <div className="mt-3 flex items-center justify-between text-sm text-zinc-400">
            <span>{loading ? 'Loading…' : `${list.total} total`}</span>
            <div className="flex items-center gap-2">
              <Button disabled={page === 0 || loading} onClick={() => setPage(page - 1)}>Previous</Button>
              <span>{page + 1} / {pages}</span>
              <Button disabled={page + 1 >= pages || loading} onClick={() => setPage(page + 1)}>Next</Button>
            </div>
          </div>
        </Card>
        <Card>
          {selected ? (
            <div className="space-y-3 text-sm">
              <h2 className="text-lg font-semibold">{selected.tenant_slug}</h2>
              <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1">
                <dt className="text-zinc-400">Type</dt><dd>{kindLabels[selected.kind]}</dd>
                <dt className="text-zinc-400">Amount</dt><dd>{formatIRR(selected.amount_irr)}</dd>
                <dt className="text-zinc-400">Tracking</dt><dd>{selected.tracking_number}</dd>
                <dt className="text-zinc-400">Bank date</dt><dd>{formatDate(selected.bank_transaction_at)}</dd>
                <dt className="text-zinc-400">Status</dt><dd>{selected.status.replace('_', ' ')}</dd>
                {selected.reviewed_by ? <><dt className="text-zinc-400">Reviewed</dt><dd>{selected.reviewed_by} · {formatDate(selected.reviewed_at)}</dd></> : null}
                {selected.rejection_reason ? <><dt className="text-zinc-400">Reason</dt><dd>{selected.rejection_reason}</dd></> : null}
              </dl>
              {receipt ? (
                <a href={receipt} target="_blank" rel="noreferrer" className="block">
                  {/* eslint-disable-next-line @next/next/no-img-element -- private, session-gated image; next/image cannot proxy it. */}
                  <img src={receipt} alt={`Receipt ${selected.tracking_number}`} className="max-h-96 w-full rounded-md border border-zinc-800 object-contain" />
                  <span className="mt-1 block text-cyan-300 hover:underline">Open full size</span>
                </a>
              ) : <p className="text-zinc-400">No receipt image was uploaded.</p>}
              {selected.status === 'pending_review' ? (
                <div className="space-y-2">
                  <Input aria-label="Rejection reason" placeholder="Rejection reason (required to reject)" value={reason} onChange={(e) => setReason(e.target.value)} />
                  <div className="flex gap-2">
                    <Button disabled={busy} onClick={() => void review('approve')}>Approve</Button>
                    <Button disabled={busy} className="bg-red-500 text-white hover:bg-red-400" onClick={() => void review('reject')}>Reject</Button>
                  </div>
                </div>
              ) : null}
            </div>
          ) : <p className="text-sm text-zinc-400">Select a transfer to see its receipt.</p>}
        </Card>
      </div>
    </div>
  );
}
