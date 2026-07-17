'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';

type SchemaField = {
  key: string;
  label: string;
  type: string;
  secret: boolean;
  required?: boolean;
  scope: 'connection' | 'credential';
  help_text?: string;
  documentation_url?: string;
  validation?: { default?: string };
};

type ConnectionSchema = {
  provider_slug: string;
  auth_mode: string;
  version: number;
  runtime_ready: boolean;
  fields: SchemaField[];
};

type CredentialVersion = {
  id: string;
  version: number;
  configured: boolean;
  display_hint?: string;
  status: string;
  created_at: string;
  expires_at?: string;
};

type ProviderConnection = {
  id: string;
  provider_slug: string;
  name: string;
  environment: string;
  auth_mode: string;
  schema_version: number;
  config: Record<string, unknown>;
  status: string;
  is_primary: boolean;
  revision: number;
  last_test_status?: string;
  last_test_category?: string;
  last_test_summary?: string;
  active_credential_version_id?: string;
  credential_versions?: CredentialVersion[];
};

type ProviderOption = {
  slug: string;
  name: string;
  enabled: boolean;
  status: string;
};

async function platformRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api/admin/control-plane/${path}`, {
    cache: 'no-store',
    ...init,
    headers: {
      ...(init?.body ? { 'content-type': 'application/json' } : {}),
      ...init?.headers,
    },
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(payload?.error?.message ?? payload?.error ?? `Request failed (${response.status})`);
  }
  return payload.data as T;
}

function statusClass(status: string) {
  if (status === 'active' || status === 'validated' || status === 'success') return 'bg-emerald-950 text-emerald-300';
  if (status === 'disabled' || status === 'revoked' || status === 'failed') return 'bg-rose-950 text-rose-300';
  return 'bg-amber-950 text-amber-300';
}

function defaultFieldValues(schema: ConnectionSchema | null): Record<string, string> {
  return Object.fromEntries(
    (schema?.fields ?? []).flatMap((field) => field.validation?.default ? [[field.key, field.validation.default]] : []),
  );
}

export default function ProviderConnectionsPage() {
  const [connections, setConnections] = useState<ProviderConnection[]>([]);
  const [selectedId, setSelectedId] = useState('');
  const [providers, setProviders] = useState<ProviderOption[]>([]);
  const [provider, setProvider] = useState('');
  const [schemas, setSchemas] = useState<ConnectionSchema[]>([]);
  const [authMode, setAuthMode] = useState('');
  const [name, setName] = useState('');
  const [values, setValues] = useState<Record<string, string>>({});
  const [selectedSchemas, setSelectedSchemas] = useState<ConnectionSchema[]>([]);
  const [rotationValues, setRotationValues] = useState<Record<string, string>>({});
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState('');

  const selected = connections.find((item) => item.id === selectedId) ?? connections[0] ?? null;
  const schema = useMemo(() => schemas.find((item) => item.auth_mode === authMode) ?? null, [schemas, authMode]);

  const load = useCallback(async () => {
    const data = await platformRequest<ProviderConnection[]>('provider-connections');
    setConnections(data);
    setSelectedId((current) => current || data[0]?.id || '');
  }, []);

  useEffect(() => {
    let cancelled = false;
    platformRequest<ProviderConnection[]>('provider-connections')
      .then((data) => {
        if (cancelled) return;
        setConnections(data);
        setSelectedId((current) => current || data[0]?.id || '');
      })
      .catch((error) => {
        if (!cancelled) setMessage(error.message);
      });
    return () => { cancelled = true; };
  }, []);
  useEffect(() => {
    fetch('/api/admin/providers', { cache: 'no-store' })
      .then(async (response) => {
        const payload = await response.json();
        if (!response.ok) throw new Error(payload?.error ?? 'Provider catalog could not be loaded.');
        const items = (Array.isArray(payload) ? payload : payload.data ?? []) as ProviderOption[];
        const active = items.filter((item) => item.enabled && item.status === 'active');
        setProviders(active);
        setProvider((current) => current || active[0]?.slug || '');
      })
      .catch((error) => setMessage(error.message));
  }, []);
  useEffect(() => {
    if (!provider) return;
    platformRequest<ConnectionSchema[]>(`provider-connection-schemas/${provider}`)
      .then((items) => {
        setSchemas(items);
        const next = items.find((item) => item.runtime_ready) ?? items[0] ?? null;
        setAuthMode(next?.auth_mode ?? '');
        setValues(defaultFieldValues(next));
      })
      .catch((error) => setMessage(error.message));
  }, [provider]);

  useEffect(() => {
    if (!selected) return;
    platformRequest<ConnectionSchema[]>(`provider-connection-schemas/${selected.provider_slug}`)
      .then(setSelectedSchemas)
      .catch((error) => setMessage(error.message));
  }, [selected]);

  const connectionFields = useMemo(() => schema?.fields.filter((field) => field.scope === 'connection') ?? [], [schema]);
  const credentialFields = useMemo(() => schema?.fields.filter((field) => field.scope === 'credential') ?? [], [schema]);
  const selectedSchema = selectedSchemas.find((item) => item.auth_mode === selected?.auth_mode && item.version === selected?.schema_version) ?? null;
  const rotationFields = selectedSchema?.fields.filter((field) => field.scope === 'credential') ?? [];

  async function run(operation: () => Promise<unknown>, success: string) {
    setBusy(true);
    setMessage('');
    try {
      await operation();
      await load();
      setMessage(success);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Operation failed.');
    } finally {
      setBusy(false);
    }
  }

  async function createConnection() {
    const missingRequired = schema?.fields.some((field) => field.required && !values[field.key]?.trim());
    if (!schema || !name.trim() || missingRequired) {
      setMessage('Connection name and all required provider fields are required.');
      return;
    }
    await run(async () => {
      const config = Object.fromEntries(connectionFields.map((field) => [field.key, values[field.key]?.trim()]).filter(([, value]) => value));
      const connection = await platformRequest<ProviderConnection>('provider-connections', {
        method: 'POST',
        headers: { 'idempotency-key': crypto.randomUUID() },
        body: JSON.stringify({ provider_slug: provider, name: name.trim(), environment: 'production', auth_mode: schema.auth_mode, schema_version: schema.version, config, weight: 1 }),
      });
      await platformRequest<CredentialVersion>(`provider-connections/${connection.id}/credential-versions`, {
        method: 'POST',
        headers: { 'idempotency-key': crypto.randomUUID() },
        body: JSON.stringify({ secrets: Object.fromEntries(credentialFields.map((field) => [field.key, values[field.key]?.trim()]).filter(([, value]) => value)) }),
      });
      setSelectedId(connection.id);
      setName('');
      setValues({});
    }, 'Connection created. Test and activate its credential before traffic can use it.');
  }

  async function rotateCredential() {
    if (!selected || rotationFields.some((field) => field.required && !rotationValues[field.key]?.trim())) return;
    await run(() => platformRequest(`provider-connections/${selected.id}/credential-versions`, {
      method: 'POST', headers: { 'idempotency-key': crypto.randomUUID() }, body: JSON.stringify({ secrets: Object.fromEntries(rotationFields.map((field) => [field.key, rotationValues[field.key]?.trim()]).filter(([, value]) => value)) }),
    }), 'A new write-only credential version was created.');
    setRotationValues({});
  }

  async function versionAction(version: number, action: 'test' | 'activate' | 'rollback' | 'revoke') {
    if (!selected) return;
    if (action === 'revoke' && !window.confirm(`Revoke credential version ${version}? This cannot be undone.`)) return;
    await run(() => platformRequest(`provider-connections/${selected.id}/credential-versions/${version}/${action}`, {
      method: 'POST',
      headers: action === 'activate' ? { 'idempotency-key': crypto.randomUUID() } : undefined,
      body: action === 'test' ? undefined : JSON.stringify({ revision: selected.revision, reason: `${action} from support-admin` }),
    }), `Credential ${action} completed.`);
  }

  async function disableConnection() {
    if (!selected || !window.confirm(`Disable ${selected.name}? New requests will fail closed.`)) return;
    await run(() => platformRequest(`provider-connections/${selected.id}/disable`, {
      method: 'POST', body: JSON.stringify({ revision: selected.revision, reason: 'disabled from support-admin' }),
    }), 'Connection disabled and removed from runtime selection.');
  }

  return (
    <div className="space-y-5" dir="auto">
      <div>
        <h1 className="text-xl font-semibold">Platform provider credentials</h1>
        <p className="mt-1 text-sm text-zinc-400">Encrypted, versioned credentials for wallet-billed provider traffic. Existing secret values are never returned.</p>
      </div>
      {message ? <div role="status" className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm">{message}</div> : null}

      <div className="grid gap-5 xl:grid-cols-[minmax(20rem,0.8fr)_minmax(30rem,1.2fr)]">
        <div className="space-y-5">
          <Card>
            <h2 className="mb-3 font-medium">New connection</h2>
            <div className="space-y-3">
              <label className="block text-sm">Provider<select className="mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2" value={provider} onChange={(event) => { setSchemas([]); setAuthMode(''); setProvider(event.target.value); }}>{providers.map((item) => <option key={item.slug} value={item.slug}>{item.name} ({item.slug})</option>)}</select></label>
              {schemas.length > 1 ? <label className="block text-sm">Authentication mode<select className="mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2" value={authMode} onChange={(event) => { const nextMode = event.target.value; setAuthMode(nextMode); setValues(defaultFieldValues(schemas.find((item) => item.auth_mode === nextMode) ?? null)); }}>{schemas.map((item) => <option key={`${item.auth_mode}:${item.version}`} value={item.auth_mode}>{item.auth_mode}{item.runtime_ready ? '' : ' — runtime not implemented'}</option>)}</select></label> : null}
              <label className="block text-sm">Connection name<Input className="mt-1" value={name} onChange={(event) => setName(event.target.value)} autoComplete="off" /></label>
              {connectionFields.map((field) => <label key={field.key} className="block text-sm">{field.label}{field.required ? ' *' : ''}<Input className="mt-1" type={field.type === 'url' ? 'url' : 'text'} value={values[field.key] ?? ''} onChange={(event) => setValues((current) => ({ ...current, [field.key]: event.target.value }))} required={field.required} />{field.help_text ? <span className="mt-1 block text-xs text-zinc-500">{field.help_text}</span> : null}</label>)}
              {credentialFields.map((field) => <label key={field.key} className="block text-sm">{field.label}{field.required ? ' *' : ''}<Input className="mt-1" type="password" value={values[field.key] ?? ''} onChange={(event) => setValues((current) => ({ ...current, [field.key]: event.target.value }))} autoComplete="new-password" required={field.required} /><span className="mt-1 block text-xs text-zinc-500">Write-only. It is encrypted immediately and cannot be read back.</span></label>)}
              {schema && !schema.runtime_ready ? <p className="rounded-md border border-amber-800 bg-amber-950/60 p-3 text-sm text-amber-200">This official authentication schema is documented, but its token refresh, signing, or request translator is not implemented. It cannot be activated.</p> : null}
              <Button type="button" disabled={busy || !schema?.runtime_ready} onClick={createConnection}>Create draft</Button>
            </div>
          </Card>

          <Card>
            <h2 className="mb-3 font-medium">Connections</h2>
            <div className="space-y-2">
              {connections.length === 0 ? <p className="text-sm text-zinc-500">No platform connections yet.</p> : connections.map((item) => <button key={item.id} type="button" onClick={() => { setSelectedId(item.id); setRotationValues({}); }} className={`w-full rounded-md border p-3 text-start ${selected?.id === item.id ? 'border-cyan-500 bg-zinc-800' : 'border-zinc-800 bg-zinc-950'}`}><span className="flex items-center justify-between gap-2"><span className="font-medium">{item.name}</span><span className={`rounded px-2 py-0.5 text-xs ${statusClass(item.status)}`}>{item.status}</span></span><span className="mt-1 block text-xs text-zinc-500">{item.provider_slug} · {item.environment}{item.is_primary ? ' · primary' : ''}</span></button>)}
            </div>
          </Card>
        </div>

        <Card>
          {!selected ? <p className="text-sm text-zinc-500">Select or create a connection.</p> : <div className="space-y-5">
            <div className="flex flex-wrap items-start justify-between gap-3"><div><h2 className="font-medium">{selected.name}</h2><p className="text-xs text-zinc-500">Revision {selected.revision} · {selected.provider_slug}/{selected.auth_mode}</p></div><Button type="button" disabled={busy || selected.status === 'disabled'} onClick={disableConnection} className="bg-rose-600 text-white hover:bg-rose-500">Disable</Button></div>
            <div className="grid gap-3 sm:grid-cols-3"><div><span className="text-xs text-zinc-500">Runtime</span><p>{selected.is_primary && selected.status === 'active' ? 'Active primary' : 'Not selected'}</p></div><div><span className="text-xs text-zinc-500">Last test</span><p>{selected.last_test_status ?? 'Not tested'}</p></div><div><span className="text-xs text-zinc-500">Category</span><p>{selected.last_test_category ?? '—'}</p></div></div>
            {selected.last_test_summary ? <p className="rounded-md bg-zinc-950 p-3 text-sm text-zinc-300">{selected.last_test_summary}</p> : null}
            <div className="border-t border-zinc-800 pt-4"><h3 className="mb-2 text-sm font-medium">Create rotation version</h3><div className="space-y-2">{rotationFields.map((field) => <label key={field.key} className="block text-sm">{field.label}{field.required ? ' *' : ''}<Input className="mt-1" type="password" value={rotationValues[field.key] ?? ''} onChange={(event) => setRotationValues((current) => ({ ...current, [field.key]: event.target.value }))} placeholder={`${field.label} (write-only)`} autoComplete="new-password" /></label>)}<Button type="button" disabled={busy || !selectedSchema?.runtime_ready || rotationFields.some((field) => field.required && !rotationValues[field.key]?.trim())} onClick={rotateCredential}>Create</Button></div></div>
            <div className="border-t border-zinc-800 pt-4"><h3 className="mb-2 text-sm font-medium">Credential versions</h3><div className="space-y-2">{selected.credential_versions?.map((version) => <div key={version.id} className="rounded-md border border-zinc-800 bg-zinc-950 p-3"><div className="flex flex-wrap items-center justify-between gap-2"><div><span className="font-medium">Version {version.version}</span><span className={`ms-2 rounded px-2 py-0.5 text-xs ${statusClass(version.status)}`}>{version.status}</span><p className="mt-1 text-xs text-zinc-500">{version.display_hint ?? 'configured'} · {new Date(version.created_at).toLocaleString()}</p></div><div className="flex flex-wrap gap-2"><Button type="button" disabled={busy || version.status === 'revoked'} onClick={() => versionAction(version.version, 'test')}>Test</Button>{version.status === 'validated' ? <Button type="button" disabled={busy} onClick={() => versionAction(version.version, 'activate')}>Activate</Button> : null}{version.status === 'retiring' ? <Button type="button" disabled={busy} onClick={() => versionAction(version.version, 'rollback')}>Rollback</Button> : null}{version.status !== 'revoked' ? <Button type="button" disabled={busy} onClick={() => versionAction(version.version, 'revoke')} className="bg-rose-600 text-white hover:bg-rose-500">Revoke</Button> : null}</div></div></div>) ?? <p className="text-sm text-zinc-500">No versions.</p>}</div></div>
          </div>}
        </Card>
      </div>
    </div>
  );
}
