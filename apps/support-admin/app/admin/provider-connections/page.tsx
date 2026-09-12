'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  AlertTriangle,
  Ban,
  CheckCircle2,
  FlaskConical,
  KeyRound,
  PlugZap,
  Plus,
  RefreshCw,
  RotateCw,
  ShieldCheck,
} from 'lucide-react';
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

type CredentialTestResult = {
  success: boolean;
  authenticated: boolean;
  ready: boolean;
  category: string;
  summary: string;
  version: number;
  validated: boolean;
  details?: CredentialProbeDetails;
};

type CredentialProbeDetails = {
  authenticated: boolean;
  ready: boolean;
  permissions?: string[];
  credential_disabled?: boolean;
  credential_blocked?: boolean;
  account_blocked?: boolean;
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

const controlClass = 'mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 outline-none focus:border-cyan-500';
const secondaryButtonClass = 'bg-zinc-800 text-zinc-100 hover:bg-zinc-700';
const dangerButtonClass = 'bg-rose-600 text-white hover:bg-rose-500';

// Providers with no non-billable validation endpoint. Their credential probe
// falls back to a minimal real request (e.g. a one-token chat completion)
// through Silk Road, which may create a small upstream charge. Keep this in
// sync with the tunnel prober registrations in
// services/control-plane/internal/platformcredentials/tunnel_prober.go.
const BILLABLE_PROBE_PROVIDERS = new Set(['zai']);

function hasBillableProbe(providerSlug?: string) {
  return Boolean(providerSlug && BILLABLE_PROBE_PROVIDERS.has(providerSlug));
}

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

async function loadConnectionDetails(): Promise<ProviderConnection[]> {
  const data = await platformRequest<ProviderConnection[]>('provider-connections');
  return Promise.all(
    data.map((item) => platformRequest<ProviderConnection>(`provider-connections/${item.id}`).catch(() => item)),
  );
}

function statusClass(status: string) {
  if (status === 'active' || status === 'validated' || status === 'success') return 'border-emerald-800 bg-emerald-950 text-emerald-200';
  if (status === 'disabled' || status === 'revoked' || status === 'failed' || status === 'degraded') return 'border-rose-800 bg-rose-950 text-rose-200';
  return 'border-amber-800 bg-amber-950 text-amber-200';
}

function defaultFieldValues(schema: ConnectionSchema | null): Record<string, string> {
  return Object.fromEntries(
    (schema?.fields ?? []).flatMap((field) => field.validation?.default ? [[field.key, field.validation.default]] : []),
  );
}

function requiredMissing(fields: SchemaField[], values: Record<string, string>) {
  return fields.some((field) => field.required && !values[field.key]?.trim());
}

function collectValues(fields: SchemaField[], values: Record<string, string>) {
  return Object.fromEntries(
    fields
      .map((field) => [field.key, values[field.key]?.trim()] as const)
      .filter(([, value]) => Boolean(value)),
  );
}

function lifecycleState(connection: ProviderConnection | null) {
  const versions = connection?.credential_versions ?? [];
  const hasVersion = versions.length > 0;
  const hasValidated = versions.some((version) => ['validated', 'active'].includes(version.status));
  const isActive = Boolean(connection?.is_primary && connection.status === 'active');
  return { hasVersion, hasValidated, isActive };
}

function statusLabel(connection: ProviderConnection) {
  const versions = connection.credential_versions ?? [];
  if (connection.status === 'active' && connection.is_primary) return 'active primary';
  if (connection.status === 'disabled') return 'disabled';
  if (versions.length === 0) return 'needs key';
  if (versions.some((version) => version.status === 'validated')) return 'ready to activate';
  if (connection.last_test_status === 'failed') return 'test failed';
  return connection.status;
}

function formatDate(value?: string) {
  if (!value) return '-';
  return new Date(value).toLocaleString();
}

function FieldShell({
  field,
  children,
  hint,
}: {
  field: SchemaField;
  children: React.ReactNode;
  hint?: string;
}) {
  return (
    <label className="block text-sm font-medium text-zinc-100">
      <span>{field.label}{field.required ? ' *' : ''}</span>
      {children}
      {hint || field.help_text ? <span className="mt-1 block text-xs leading-5 text-zinc-500">{hint ?? field.help_text}</span> : null}
    </label>
  );
}

function StepItem({ done, label, description }: { done: boolean; label: string; description: string }) {
  return (
    <li className="flex gap-3">
      <span className={`mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full border ${done ? 'border-emerald-700 bg-emerald-950 text-emerald-300' : 'border-zinc-700 bg-zinc-950 text-zinc-500'}`}>
        {done ? <CheckCircle2 className="size-3.5" aria-hidden="true" /> : <span className="size-1.5 rounded-full bg-current" />}
      </span>
      <span>
        <span className="block text-sm font-medium text-zinc-100">{label}</span>
        <span className="block text-xs leading-5 text-zinc-500">{description}</span>
      </span>
    </li>
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
  const [stage, setStage] = useState('');

  const selected = connections.find((item) => item.id === selectedId) ?? connections[0] ?? null;
  const schema = useMemo(() => schemas.find((item) => item.auth_mode === authMode) ?? null, [schemas, authMode]);
  const connectionFields = useMemo(() => schema?.fields.filter((field) => field.scope === 'connection') ?? [], [schema]);
  const credentialFields = useMemo(() => schema?.fields.filter((field) => field.scope === 'credential') ?? [], [schema]);
  const selectedSchema = selectedSchemas.find((item) => item.auth_mode === selected?.auth_mode && item.version === selected?.schema_version) ?? null;
  const rotationFields = selectedSchema?.fields.filter((field) => field.scope === 'credential') ?? [];
  const selectedVersions = selected?.credential_versions ?? [];
  const lifecycle = lifecycleState(selected);
  const createMissing = !schema || !name.trim() || requiredMissing([...(schema?.fields ?? [])], values);
  const rotationMissing = requiredMissing(rotationFields, rotationValues);

  const load = useCallback(async () => {
    const data = await loadConnectionDetails();
    setConnections(data);
    setSelectedId((current) => current || data[0]?.id || '');
  }, []);

  useEffect(() => {
    let cancelled = false;
    loadConnectionDetails()
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
        const ordered = [...active].sort((left, right) => Number(right.slug === 'openai') - Number(left.slug === 'openai'));
        setProviders(ordered);
        setProvider((current) => current || ordered.find((item) => item.slug === 'openai')?.slug || ordered[0]?.slug || '');
      })
      .catch((error) => setMessage(error.message));
  }, []);

  useEffect(() => {
    if (!provider) return;
    let cancelled = false;
    platformRequest<ConnectionSchema[]>(`provider-connection-schemas/${provider}`)
      .then((items) => {
        if (cancelled) return;
        setSchemas(items);
        const next = items.find((item) => item.runtime_ready) ?? items[0] ?? null;
        setAuthMode(next?.auth_mode ?? '');
        setValues(defaultFieldValues(next));
      })
      .catch((error) => {
        if (!cancelled) setMessage(error.message);
      });
    return () => { cancelled = true; };
  }, [provider]);

  useEffect(() => {
    if (!selected) return;
    let cancelled = false;
    platformRequest<ConnectionSchema[]>(`provider-connection-schemas/${selected.provider_slug}`)
      .then((items) => {
        if (!cancelled) setSelectedSchemas(items);
      })
      .catch((error) => {
        if (!cancelled) setMessage(error.message);
      });
    return () => { cancelled = true; };
  }, [selected]);

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
      setStage('');
    }
  }

  async function createConnection() {
    if (!schema || createMissing) {
      setMessage('Connection name and every required provider field are required.');
      return;
    }

    let created: ProviderConnection | null = null;
    setBusy(true);
    setMessage('');
    try {
      setStage('Creating draft connection...');
      const config = collectValues(connectionFields, values);
      created = await platformRequest<ProviderConnection>('provider-connections', {
        method: 'POST',
        headers: { 'idempotency-key': crypto.randomUUID() },
        body: JSON.stringify({
          provider_slug: provider,
          name: name.trim(),
          environment: 'production',
          auth_mode: schema.auth_mode,
          schema_version: schema.version,
          config,
          weight: 1,
        }),
      });
      setSelectedId(created.id);

      setStage('Encrypting and saving credential version...');
      await platformRequest<CredentialVersion>(`provider-connections/${created.id}/credential-versions`, {
        method: 'POST',
        headers: { 'idempotency-key': crypto.randomUUID() },
        body: JSON.stringify({ secrets: collectValues(credentialFields, values) }),
      });

      setStage('Refreshing connection...');
      await load();
      setSelectedId(created.id);
      setName('');
      setValues(defaultFieldValues(schema));
      setMessage('Draft and credential version created. Run Test, then Activate after validation succeeds.');
    } catch (error) {
      if (created) {
        await load();
        setSelectedId(created.id);
        setMessage(`Draft was created, but the credential version was not saved: ${error instanceof Error ? error.message : 'operation failed'}. Add the key in Complete draft.`);
      } else {
        setMessage(error instanceof Error ? error.message : 'Connection could not be created.');
      }
    } finally {
      setBusy(false);
      setStage('');
    }
  }

  async function rotateCredential() {
    if (!selected || rotationMissing) return;
    await run(() => platformRequest(`provider-connections/${selected.id}/credential-versions`, {
      method: 'POST',
      headers: { 'idempotency-key': crypto.randomUUID() },
      body: JSON.stringify({ secrets: collectValues(rotationFields, rotationValues) }),
    }), lifecycle.hasVersion ? 'A new write-only credential version was created.' : 'Credential version saved. Run Test, then Activate after validation succeeds.');
    setRotationValues({});
  }

  async function versionAction(version: number, action: 'test' | 'activate' | 'rollback' | 'revoke') {
    if (!selected) return;
    if (action === 'revoke' && !window.confirm(`Revoke credential version ${version}? This cannot be undone.`)) return;

    setBusy(true);
    setMessage('');
    try {
      if (action === 'test') {
        const result = await platformRequest<CredentialTestResult>(`provider-connections/${selected.id}/credential-versions/${version}/test`, {
          method: 'POST',
        });
        await load();
        setMessage(result.success ? 'Credential test passed. This version can now be activated.' : `Credential test failed: ${result.summary}`);
        return;
      }

      await platformRequest(`provider-connections/${selected.id}/credential-versions/${version}/${action}`, {
        method: 'POST',
        headers: action === 'activate' ? { 'idempotency-key': crypto.randomUUID() } : undefined,
        body: JSON.stringify({ revision: selected.revision, reason: `${action} from support-admin` }),
      });
      await load();
      setMessage(action === 'activate' ? 'Connection is active and selected for runtime traffic.' : `Credential ${action} completed.`);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Operation failed.');
    } finally {
      setBusy(false);
    }
  }

  async function disableConnection() {
    if (!selected || !window.confirm(`Disable ${selected.name}? New requests will fail closed.`)) return;
    await run(() => platformRequest(`provider-connections/${selected.id}/disable`, {
      method: 'POST',
      body: JSON.stringify({ revision: selected.revision, reason: 'disabled from support-admin' }),
    }), 'Connection disabled and removed from runtime selection.');
  }

  return (
    <div className="flex flex-col gap-6" dir="auto">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-50 md:text-3xl">Platform provider credentials</h1>
          <p className="mt-2 max-w-3xl text-sm leading-6 text-zinc-400">
            Create encrypted provider credentials in a controlled flow: draft, credential version, test, then activate for wallet-billed traffic.
          </p>
        </div>
        <Button type="button" className={secondaryButtonClass} disabled={busy} onClick={() => void load()}>
          <RefreshCw className="size-4" aria-hidden="true" />
          Refresh
        </Button>
      </div>

      {message ? (
        <div role="status" className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm leading-6 text-zinc-100">
          {message}
        </div>
      ) : null}

      <div className="grid gap-5 xl:grid-cols-[minmax(22rem,0.85fr)_minmax(34rem,1.15fr)]">
        <div className="flex flex-col gap-5">
          <Card className="space-y-5">
            <div>
              <h2 className="flex items-center gap-2 text-base font-semibold text-zinc-50">
                <Plus className="size-4 text-cyan-300" aria-hidden="true" />
                Create provider connection
              </h2>
              <p className="mt-1 text-sm leading-6 text-zinc-400">
                Saves the connection draft and its first write-only credential version in one guided action.
              </p>
            </div>

            <div className="grid gap-3">
              <label className="block text-sm font-medium text-zinc-100">
                Provider
                <select
                  className={controlClass}
                  value={provider}
                  disabled={busy}
                  onChange={(event) => {
                    setProvider(event.target.value);
                    setSchemas([]);
                    setAuthMode('');
                    setValues({});
                  }}
                >
                  {providers.map((item) => <option key={item.slug} value={item.slug}>{item.name} ({item.slug})</option>)}
                </select>
              </label>

              {schemas.length > 1 ? (
                <label className="block text-sm font-medium text-zinc-100">
                  Authentication mode
                  <select
                    className={controlClass}
                    value={authMode}
                    disabled={busy}
                    onChange={(event) => {
                      const nextMode = event.target.value;
                      const nextSchema = schemas.find((item) => item.auth_mode === nextMode) ?? null;
                      setAuthMode(nextMode);
                      setValues(defaultFieldValues(nextSchema));
                    }}
                  >
                    {schemas.map((item) => (
                      <option key={`${item.auth_mode}:${item.version}`} value={item.auth_mode}>
                        {item.auth_mode}{item.runtime_ready ? '' : ' - runtime not implemented'}
                      </option>
                    ))}
                  </select>
                </label>
              ) : null}

              <label className="block text-sm font-medium text-zinc-100">
                Connection name
                <Input className="mt-1" value={name} onChange={(event) => setName(event.target.value)} disabled={busy} autoComplete="off" />
              </label>

              {connectionFields.map((field) => (
                <FieldShell key={field.key} field={field}>
                  <Input
                    className="mt-1"
                    type={field.type === 'url' ? 'url' : 'text'}
                    value={values[field.key] ?? ''}
                    onChange={(event) => setValues((current) => ({ ...current, [field.key]: event.target.value }))}
                    disabled={busy}
                    required={field.required}
                  />
                </FieldShell>
              ))}

              {credentialFields.map((field) => (
                <FieldShell key={field.key} field={field} hint="Write-only. It is encrypted immediately and cannot be read back.">
                  <Input
                    className="mt-1 font-mono"
                    type="password"
                    value={values[field.key] ?? ''}
                    onChange={(event) => setValues((current) => ({ ...current, [field.key]: event.target.value }))}
                    disabled={busy}
                    autoComplete="new-password"
                    required={field.required}
                  />
                </FieldShell>
              ))}

              {schema && !schema.runtime_ready ? (
                <div className="flex gap-3 rounded-md border border-amber-800 bg-amber-950/70 p-3 text-sm leading-6 text-amber-100">
                  <AlertTriangle className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
                  <p>This authentication schema is documented, but runtime support is not implemented yet. The credential can still be created and stored, but it cannot be tested or activated until runtime support ships.</p>
                </div>
              ) : null}

              {schema && hasBillableProbe(schema.provider_slug) ? (
                <div className="flex gap-3 rounded-md border border-amber-800 bg-amber-950/70 p-3 text-sm leading-6 text-amber-100">
                  <AlertTriangle className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
                  <p>This provider has no non-billable validation endpoint. Testing or activating this credential sends a minimal real request upstream and may incur a small charge.</p>
                </div>
              ) : null}

              <Button type="button" className="w-full" disabled={busy || createMissing} onClick={createConnection}>
                <KeyRound className="size-4" aria-hidden="true" />
                Create draft and save key
              </Button>
              {stage ? <p className="text-xs text-zinc-500">{stage}</p> : null}
            </div>
          </Card>

          <Card>
            <div className="mb-3 flex items-center justify-between gap-3">
              <h2 className="text-base font-semibold text-zinc-50">Connections</h2>
              <span className="text-xs text-zinc-500">{connections.length} total</span>
            </div>
            <div className="space-y-2">
              {connections.length === 0 ? (
                <div className="rounded-md border border-dashed border-zinc-800 bg-zinc-950 p-4 text-sm text-zinc-500">No platform connections yet.</div>
              ) : connections.map((item) => {
                const versionCount = item.credential_versions?.length ?? 0;
                const current = selected?.id === item.id;
                const label = statusLabel(item);
                return (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => { setSelectedId(item.id); setRotationValues({}); }}
                    className={`w-full rounded-md border p-3 text-start transition ${current ? 'border-cyan-500 bg-cyan-950/30' : 'border-zinc-800 bg-zinc-950 hover:border-zinc-700'}`}
                  >
                    <span className="flex items-center justify-between gap-2">
                      <span className="min-w-0 truncate font-medium text-zinc-100">{item.name}</span>
                      <span className={`shrink-0 rounded border px-2 py-0.5 text-xs ${statusClass(label)}`}>{label}</span>
                    </span>
                    <span className="mt-1 flex flex-wrap gap-x-2 gap-y-1 text-xs text-zinc-500">
                      <span>{item.provider_slug} / {item.environment}</span>
                      <span>{versionCount === 1 ? '1 version' : `${versionCount} versions`}</span>
                    </span>
                  </button>
                );
              })}
            </div>
          </Card>
        </div>

        <Card>
          {!selected ? (
            <div className="rounded-md border border-dashed border-zinc-800 bg-zinc-950 p-6 text-sm text-zinc-500">Select or create a connection.</div>
          ) : (
            <div className="space-y-6">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="text-lg font-semibold text-zinc-50">{selected.name}</h2>
                    <span className={`rounded border px-2 py-0.5 text-xs ${statusClass(statusLabel(selected))}`}>{statusLabel(selected)}</span>
                  </div>
                  <p className="mt-1 text-xs text-zinc-500">Revision {selected.revision} - {selected.provider_slug}/{selected.auth_mode} - {selected.environment}</p>
                </div>
                <Button type="button" disabled={busy || selected.status === 'disabled'} onClick={disableConnection} className={dangerButtonClass}>
                  <Ban className="size-4" aria-hidden="true" />
                  Disable
                </Button>
              </div>

              <ol className="grid gap-3 rounded-md border border-zinc-800 bg-zinc-950 p-4 md:grid-cols-2">
                <StepItem done label="Draft saved" description="Connection metadata exists." />
                <StepItem done={lifecycle.hasVersion} label="Credential version" description={lifecycle.hasVersion ? 'A write-only key is stored.' : 'Add the provider API key.'} />
                <StepItem done={lifecycle.hasValidated} label="Provider test" description={lifecycle.hasValidated ? 'Provider accepted a credential.' : 'Run Test before activation.'} />
                <StepItem done={lifecycle.isActive} label="Runtime active" description={lifecycle.isActive ? 'Selected for production traffic.' : 'Activate a validated version.'} />
              </ol>

              <div className="grid gap-3 sm:grid-cols-3">
                <div className="rounded-md border border-zinc-800 bg-zinc-950 p-3">
                  <span className="text-xs text-zinc-500">Runtime</span>
                  <p className="mt-1 text-sm font-medium text-zinc-100">{lifecycle.isActive ? 'Active primary' : 'Not selected'}</p>
                </div>
                <div className="rounded-md border border-zinc-800 bg-zinc-950 p-3">
                  <span className="text-xs text-zinc-500">Last test</span>
                  <p className="mt-1 text-sm font-medium text-zinc-100">{selected.last_test_status ?? 'Not tested'}</p>
                </div>
                <div className="rounded-md border border-zinc-800 bg-zinc-950 p-3">
                  <span className="text-xs text-zinc-500">Category</span>
                  <p className="mt-1 text-sm font-medium text-zinc-100">{selected.last_test_category ?? '-'}</p>
                </div>
              </div>

              {selected.last_test_summary ? (
                <div className={`rounded-md border p-3 text-sm leading-6 ${selected.last_test_status === 'failed' ? 'border-rose-800 bg-rose-950/60 text-rose-100' : 'border-zinc-800 bg-zinc-950 text-zinc-300'}`}>
                  {selected.last_test_summary}
                </div>
              ) : null}

              <div className="border-t border-zinc-800 pt-5">
                <h3 className="flex items-center gap-2 text-sm font-semibold text-zinc-50">
                  <RotateCw className="size-4 text-cyan-300" aria-hidden="true" />
                  {lifecycle.hasVersion ? 'Create rotation version' : 'Complete draft'}
                </h3>
                <p className="mt-1 text-sm leading-6 text-zinc-400">
                  {lifecycle.hasVersion ? 'Store a replacement key as a new draft version before testing it.' : 'This connection has no key version yet. Add the provider API key to continue.'}
                </p>
                <div className="mt-3 grid gap-3">
                  {rotationFields.map((field) => (
                    <FieldShell key={field.key} field={field} hint="Write-only. Existing secret values are never returned.">
                      <Input
                        className="mt-1 font-mono"
                        type="password"
                        value={rotationValues[field.key] ?? ''}
                        onChange={(event) => setRotationValues((current) => ({ ...current, [field.key]: event.target.value }))}
                        placeholder={`${field.label} (write-only)`}
                        disabled={busy}
                        autoComplete="new-password"
                      />
                    </FieldShell>
                  ))}
                  <Button type="button" className="w-full" disabled={busy || rotationMissing} onClick={rotateCredential}>
                    <KeyRound className="size-4" aria-hidden="true" />
                    {lifecycle.hasVersion ? 'Create version' : 'Save key version'}
                  </Button>
                </div>
              </div>

              <div className="border-t border-zinc-800 pt-5">
                <h3 className="flex items-center gap-2 text-sm font-semibold text-zinc-50">
                  <ShieldCheck className="size-4 text-cyan-300" aria-hidden="true" />
                  Credential versions
                </h3>
                {hasBillableProbe(selected.provider_slug) ? (
                  <div className="mt-3 flex gap-3 rounded-md border border-amber-800 bg-amber-950/70 p-3 text-sm leading-6 text-amber-100">
                    <AlertTriangle className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
                    <p>{selected.provider_slug} has no non-billable validation endpoint. Pressing Test sends a minimal real request upstream and may incur a small charge.</p>
                  </div>
                ) : null}
                <div className="mt-3 space-y-3">
                  {selectedVersions.length === 0 ? (
                    <div className="rounded-md border border-dashed border-amber-800 bg-amber-950/30 p-4 text-sm leading-6 text-amber-100">
                      No credential version exists for this draft. Save a key version before testing or activating.
                    </div>
                  ) : selectedVersions.map((version) => (
                    <div key={version.id} className="rounded-md border border-zinc-800 bg-zinc-950 p-3">
                      <div className="flex flex-wrap items-center justify-between gap-3">
                        <div>
                          <div className="flex flex-wrap items-center gap-2">
                            <span className="font-medium text-zinc-100">Version {version.version}</span>
                            <span className={`rounded border px-2 py-0.5 text-xs ${statusClass(version.status)}`}>{version.status}</span>
                          </div>
                          <p className="mt-1 text-xs text-zinc-500">{version.display_hint ?? 'configured'} - {formatDate(version.created_at)}</p>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          <Button type="button" className={secondaryButtonClass} disabled={busy || version.status === 'revoked'} onClick={() => void versionAction(version.version, 'test')}>
                            <FlaskConical className="size-4" aria-hidden="true" />
                            Test
                          </Button>
                          {version.status === 'validated' ? (
                            <Button type="button" disabled={busy} onClick={() => void versionAction(version.version, 'activate')}>
                              <PlugZap className="size-4" aria-hidden="true" />
                              Activate
                            </Button>
                          ) : null}
                          {version.status === 'retiring' ? (
                            <Button type="button" className={secondaryButtonClass} disabled={busy} onClick={() => void versionAction(version.version, 'rollback')}>Rollback</Button>
                          ) : null}
                          {version.status !== 'revoked' ? (
                            <Button type="button" disabled={busy} onClick={() => void versionAction(version.version, 'revoke')} className={dangerButtonClass}>Revoke</Button>
                          ) : null}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
