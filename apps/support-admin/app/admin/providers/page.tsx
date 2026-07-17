"use client";

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { resolveAssetUrl } from '@/lib/api';
import { prettyPrintJson } from 'pretty-print-json';
import 'pretty-print-json/css/pretty-print-json.css';

type Provider = {
  slug: string;
  name: string;
  description?: string | null;
  websiteUrl?: string | null;
  docsUrl?: string | null;
  status?: string;
  supportsByok?: boolean;
  enabled: boolean;
  logoUrl?: string | null;
  connectionProfile?: {
    providerType: string;
    authMode?: string | null;
    configJson: Record<string, unknown>;
  } | null;
};

type ProviderDraft = {
  name: string;
  description: string;
  websiteUrl: string;
  docsUrl: string;
  status: string;
  supportsByok: boolean;
  providerType: string;
  authMode: string;
  configJson: string;
  logoFile: File | null;
};

type TabKey = 'general' | 'connection' | 'branding' | 'danger';

type ProviderKind = 'api_key_only' | 'aws_bedrock' | 'azure_ai_foundry' | 'azure_openai' | 'google_vertex' | 'databricks';

const PROVIDER_KIND_BY_SLUG: Record<string, ProviderKind> = {
  'aws-bedrock': 'aws_bedrock',
  bedrock: 'aws_bedrock',
  'azure-ai-foundry': 'azure_ai_foundry',
  'azure-foundry': 'azure_ai_foundry',
  'azure-openai': 'azure_openai',
  'google-vertex': 'google_vertex',
  vertex: 'google_vertex',
  databricks: 'databricks',
};

function inferProviderKind(providerSlug: string): ProviderKind {
  return PROVIDER_KIND_BY_SLUG[providerSlug.toLowerCase()] || 'api_key_only';
}

function toDraft(provider: Provider): ProviderDraft {
  return {
    name: provider.name || '',
    description: provider.description || '',
    websiteUrl: provider.websiteUrl || '',
    docsUrl: provider.docsUrl || '',
    status: provider.status || 'active',
    supportsByok: Boolean(provider.supportsByok),
    providerType: provider.connectionProfile?.providerType || provider.slug,
    authMode: provider.connectionProfile?.authMode || 'api_key',
    configJson: JSON.stringify(provider.connectionProfile?.configJson || {}, null, 2),
    logoFile: null,
  };
}

function parseConfigJson(raw: string): Record<string, unknown> {
  try {
    const parsed = JSON.parse(raw || '{}');
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>;
    }
    return {};
  } catch {
    return {};
  }
}

function updateConfigField(
  currentJson: string,
  key: string,
  value: string | boolean,
): string {
  const cfg = parseConfigJson(currentJson);
  cfg[key] = value;
  return JSON.stringify(cfg, null, 2);
}

function validateConnectionConfig(kind: ProviderKind, rawJson: string): string | null {
  const cfg = parseConfigJson(rawJson);
  const required = (value: unknown) =>
    typeof value === 'string' ? value.trim().length > 0 : Boolean(value);

  if (kind === 'api_key_only' && !required(cfg.apiKey)) {
    return 'API key is required for this provider.';
  }
  if (kind === 'azure_ai_foundry' || kind === 'azure_openai') {
    if (!required(cfg.endpoint) || !required(cfg.deployment) || !required(cfg.apiVersion)) {
      return 'Endpoint, deployment and apiVersion are required.';
    }
    if (String(cfg.authMode || 'api_key') === 'api_key' && !required(cfg.apiKey)) {
      return 'API key is required when authMode is api_key.';
    }
  }
  if (kind === 'google_vertex') {
    if (!required(cfg.projectId) || !required(cfg.location) || !required(cfg.model)) {
      return 'projectId, location and model are required.';
    }
    if (String(cfg.authMode || '') === 'service_account_json' && !required(cfg.serviceAccountJson)) {
      return 'serviceAccountJson is required when authMode is service_account_json.';
    }
  }
  if (kind === 'aws_bedrock') {
    if (!required(cfg.region) || !required(cfg.modelId)) {
      return 'region and modelId are required.';
    }
    if (String(cfg.authMode || '') === 'aws_sigv4') {
      if (!required(cfg.accessKeyId) || !required(cfg.secretAccessKey)) {
        return 'accessKeyId and secretAccessKey are required for aws_sigv4.';
      }
    }
    if (String(cfg.authMode || '') === 'bedrock_api_key' && !required(cfg.apiKey)) {
      return 'apiKey is required for bedrock_api_key mode.';
    }
  }
  if (kind === 'databricks') {
    if (!required(cfg.workspaceUrl) || !required(cfg.servingEndpoint)) {
      return 'workspaceUrl and servingEndpoint are required.';
    }
    if (String(cfg.authMode || 'pat') === 'pat' && !required(cfg.apiKey)) {
      return 'PAT/API key is required for pat mode.';
    }
    if (String(cfg.authMode || '') === 'oauth_m2m') {
      if (!required(cfg.clientId) || !required(cfg.clientSecret)) {
        return 'clientId and clientSecret are required for oauth_m2m.';
      }
    }
  }
  return null;
}

function draftFingerprint(draft: ProviderDraft): string {
  const config = (() => {
    try {
      return JSON.stringify(JSON.parse(draft.configJson || '{}'));
    } catch {
      return draft.configJson;
    }
  })();
  return JSON.stringify({
    name: draft.name,
    description: draft.description,
    websiteUrl: draft.websiteUrl,
    docsUrl: draft.docsUrl,
    status: draft.status,
    supportsByok: draft.supportsByok,
    providerType: draft.providerType,
    authMode: draft.authMode,
    config,
    logoSelected: Boolean(draft.logoFile),
  });
}

function connectionFingerprint(draft: ProviderDraft): string {
  const config = (() => {
    try {
      return JSON.stringify(JSON.parse(draft.configJson || '{}'));
    } catch {
      return draft.configJson;
    }
  })();
  return JSON.stringify({
    providerType: draft.providerType,
    authMode: draft.authMode,
    config,
  });
}

export default function AdminProvidersPage() {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [drafts, setDrafts] = useState<Record<string, ProviderDraft>>({});
  const [selectedSlug, setSelectedSlug] = useState<string>('');
  const [activeTab, setActiveTab] = useState<TabKey>('general');
  const [search, setSearch] = useState('');
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string>('');
  const [rawJsonEditMode, setRawJsonEditMode] = useState(false);

  const load = async () => {
    const res = await fetch('/api/admin/providers');
    if (!res.ok) return;
    const data = (await res.json()) as Provider[];
    setProviders(data);
    setDrafts(Object.fromEntries(data.map((p) => [p.slug, toDraft(p)])));
    setSelectedSlug((prev) => prev || data[0]?.slug || '');
  };

  useEffect(() => {
    let cancelled = false;
    void fetch('/api/admin/providers')
      .then((res) => (res.ok ? res.json() : []))
      .then((data: Provider[]) => {
        if (cancelled) return;
        setProviders(data);
        setDrafts(Object.fromEntries(data.map((p) => [p.slug, toDraft(p)])));
        setSelectedSlug((prev) => prev || data[0]?.slug || '');
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return providers;
    return providers.filter((p) => p.slug.toLowerCase().includes(q) || p.name.toLowerCase().includes(q));
  }, [providers, search]);

  const selectedProvider = providers.find((p) => p.slug === selectedSlug) || null;
  const selectedDraft = selectedSlug ? drafts[selectedSlug] : undefined;
  const selectedProviderKind = selectedProvider ? inferProviderKind(selectedProvider.slug) : 'api_key_only';
  const selectedConfig = selectedDraft ? parseConfigJson(selectedDraft.configJson) : {};
  const selectedInitial = selectedProvider ? toDraft(selectedProvider) : null;
  const hasChanges = Boolean(
    selectedDraft && selectedInitial && draftFingerprint(selectedDraft) !== draftFingerprint(selectedInitial),
  );
  const hasConnectionChanges = Boolean(
    selectedDraft
      && selectedInitial
      && connectionFingerprint(selectedDraft) !== connectionFingerprint(selectedInitial),
  );
  const prettyJsonHtml = useMemo(() => {
    if (!selectedDraft) return '';
    try {
      return prettyPrintJson.toHtml(JSON.parse(selectedDraft.configJson || '{}'));
    } catch {
      return prettyPrintJson.toHtml({ error: 'Invalid JSON' });
    }
  }, [selectedDraft]);

  const setSelectedDraft = (next: Partial<ProviderDraft>) => {
    if (!selectedSlug || !selectedDraft) return;
    setDrafts((prev) => ({ ...prev, [selectedSlug]: { ...selectedDraft, ...next } }));
  };

  const saveProviderSettings = async () => {
    if (!selectedProvider || !selectedDraft) return;
    if (hasConnectionChanges) {
      const parseError = (() => {
        try {
          JSON.parse(selectedDraft.configJson || '{}');
          return null;
        } catch {
          return 'Connection JSON is not valid.';
        }
      })();
      if (parseError) {
        setMessage(parseError);
        return;
      }
      const validationError = validateConnectionConfig(selectedProviderKind, selectedDraft.configJson);
      if (validationError) {
        setMessage(validationError);
        return;
      }
    }
    setSaving(true);
    setMessage('');
    await fetch('/api/admin/providers', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({
        slug: selectedProvider.slug,
        name: selectedDraft.name,
        description: selectedDraft.description || undefined,
        websiteUrl: selectedDraft.websiteUrl || undefined,
        docsUrl: selectedDraft.docsUrl || undefined,
        status: selectedDraft.status || 'active',
        supportsByok: selectedDraft.supportsByok,
        enabled: selectedProvider.enabled,
      }),
    });
    if (hasConnectionChanges) {
      await fetch('/api/admin/providers/connection-profile', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          providerSlug: selectedProvider.slug,
          providerType: selectedDraft.providerType,
          authMode: selectedDraft.authMode || undefined,
          configJson: JSON.parse(selectedDraft.configJson || '{}'),
        }),
      });
    }
    if (selectedDraft.logoFile) {
      const fd = new FormData();
      fd.append('slug', selectedProvider.slug);
      fd.append('file', selectedDraft.logoFile);
      await fetch('/api/admin/providers/logo', { method: 'POST', body: fd });
    }
    await load();
    setSaving(false);
    setMessage('Changes saved successfully.');
  };

  const tabClass = (tab: TabKey) =>
    `rounded-md px-3 py-2 text-sm ${activeTab === tab ? 'bg-cyan-500 text-black' : 'bg-zinc-800 text-zinc-300 hover:bg-zinc-700'}`;

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Provider Management</h1>
      <div className="grid gap-4 lg:grid-cols-[320px_1fr]">
        <aside className="space-y-3 rounded-xl border border-zinc-800 bg-zinc-900 p-3">
          <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search provider..." />
          <div className="max-h-[70vh] space-y-2 overflow-auto pr-1">
            {filtered.map((p) => (
              <button
                key={p.slug}
                type="button"
                className={`w-full rounded-lg border px-3 py-2 text-left ${
                  selectedSlug === p.slug ? 'border-cyan-500 bg-zinc-800' : 'border-zinc-700 bg-zinc-950 hover:border-zinc-500'
                }`}
                onClick={() => setSelectedSlug(p.slug)}
              >
                <div className="flex items-center justify-between">
                  <span className="flex items-center gap-2 font-medium">
                    {resolveAssetUrl(p.logoUrl) ? (
                      <img
                        src={resolveAssetUrl(p.logoUrl) || ''}
                        alt={p.name}
                        className="h-4 w-4 rounded"
                      />
                    ) : null}
                    {p.name}
                  </span>
                  <span className={`rounded px-2 py-0.5 text-xs ${p.enabled ? 'bg-emerald-950 text-emerald-300' : 'bg-rose-950 text-rose-300'}`}>
                    {p.enabled ? 'Active' : 'Disabled'}
                  </span>
                </div>
                <div className="text-xs text-zinc-500">{p.slug}</div>
              </button>
            ))}
          </div>
        </aside>

        <section className="space-y-4 rounded-xl border border-zinc-800 bg-zinc-900 p-4">
          {!selectedProvider || !selectedDraft ? (
            <p className="text-zinc-400">No provider selected.</p>
          ) : (
            <>
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-semibold">{selectedProvider.name}</h2>
                  <p className="text-xs text-zinc-500">{selectedProvider.slug}</p>
                </div>
                <Button
                  onClick={async () => {
                    await fetch('/api/admin/providers/enabled', {
                      method: 'POST',
                      headers: { 'content-type': 'application/json' },
                      body: JSON.stringify({ slug: selectedProvider.slug, enabled: !selectedProvider.enabled }),
                    });
                    await load();
                  }}
                  className={selectedProvider.enabled ? 'bg-rose-500 text-black hover:bg-rose-400' : 'bg-emerald-500 text-black hover:bg-emerald-400'}
                >
                  {selectedProvider.enabled ? 'Disable Provider' : 'Enable Provider'}
                </Button>
              </div>

              <div className="flex flex-wrap gap-2">
                <button type="button" className={tabClass('general')} onClick={() => setActiveTab('general')}>General</button>
                <Link className="rounded-md bg-zinc-800 px-3 py-2 text-sm text-zinc-300 hover:bg-zinc-700" href="/admin/provider-connections">Credentials</Link>
                <button type="button" className={tabClass('branding')} onClick={() => setActiveTab('branding')}>Branding</button>
                <button type="button" className={tabClass('danger')} onClick={() => setActiveTab('danger')}>Danger Zone</button>
              </div>

              {activeTab === 'general' ? (
                <div className="grid gap-3 md:grid-cols-2">
                  <Input value={selectedDraft.name} onChange={(e) => setSelectedDraft({ name: e.target.value })} placeholder="Name" />
                  <Input value={selectedDraft.status} onChange={(e) => setSelectedDraft({ status: e.target.value })} placeholder="Status" />
                  <Input value={selectedDraft.websiteUrl} onChange={(e) => setSelectedDraft({ websiteUrl: e.target.value })} placeholder="Website URL" />
                  <Input value={selectedDraft.docsUrl} onChange={(e) => setSelectedDraft({ docsUrl: e.target.value })} placeholder="Docs URL" />
                  <div className="md:col-span-2">
                    <Input value={selectedDraft.description} onChange={(e) => setSelectedDraft({ description: e.target.value })} placeholder="Description" />
                  </div>
                  <label className="flex items-center gap-2 text-sm text-zinc-300">
                    <input
                      type="checkbox"
                      checked={selectedDraft.supportsByok}
                      onChange={(e) => setSelectedDraft({ supportsByok: e.target.checked })}
                    />
                    Supports BYOK
                  </label>
                </div>
              ) : null}

              {activeTab === 'connection' ? (
                <div className="space-y-3">
                  <div className="grid gap-3 md:grid-cols-2">
                    <Input value={selectedDraft.providerType} onChange={(e) => setSelectedDraft({ providerType: e.target.value })} placeholder="Provider Type" />
                    <Input value={selectedDraft.authMode} onChange={(e) => setSelectedDraft({ authMode: e.target.value })} placeholder="Auth Mode" />
                  </div>
                  {selectedProviderKind === 'api_key_only' ? (
                    <div className="grid gap-3 md:grid-cols-2">
                      <Input
                        placeholder="API Key"
                        value={String(selectedConfig.apiKey || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'apiKey', e.target.value) })}
                      />
                      <Input
                        placeholder="Base URL (optional)"
                        value={String(selectedConfig.baseUrl || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'baseUrl', e.target.value) })}
                      />
                      <Input
                        placeholder="Default model (optional)"
                        value={String(selectedConfig.defaultModel || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'defaultModel', e.target.value) })}
                      />
                      <Input
                        placeholder="Organization / project (optional)"
                        value={String(selectedConfig.organization || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'organization', e.target.value) })}
                      />
                    </div>
                  ) : null}

                  {selectedProviderKind === 'azure_ai_foundry' || selectedProviderKind === 'azure_openai' ? (
                    <div className="grid gap-3 md:grid-cols-2">
                      <Input
                        placeholder="Endpoint"
                        value={String(selectedConfig.endpoint || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'endpoint', e.target.value) })}
                      />
                      <Input
                        placeholder="Deployment"
                        value={String(selectedConfig.deployment || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'deployment', e.target.value) })}
                      />
                      <Input
                        placeholder="API Version"
                        value={String(selectedConfig.apiVersion || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'apiVersion', e.target.value) })}
                      />
                      <Input
                        placeholder="Auth Mode (api_key / entra_id)"
                        value={String(selectedConfig.authMode || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'authMode', e.target.value) })}
                      />
                      <Input
                        placeholder="API Key (for api_key mode)"
                        value={String(selectedConfig.apiKey || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'apiKey', e.target.value) })}
                      />
                      <Input
                        placeholder="Tenant ID (for entra_id)"
                        value={String(selectedConfig.tenantId || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'tenantId', e.target.value) })}
                      />
                      <Input
                        placeholder="Client ID (for entra_id)"
                        value={String(selectedConfig.clientId || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'clientId', e.target.value) })}
                      />
                      <Input
                        placeholder="Client Secret (for entra_id)"
                        value={String(selectedConfig.clientSecret || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'clientSecret', e.target.value) })}
                      />
                    </div>
                  ) : null}

                  {selectedProviderKind === 'google_vertex' ? (
                    <div className="grid gap-3 md:grid-cols-2">
                      <Input
                        placeholder="Project ID"
                        value={String(selectedConfig.projectId || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'projectId', e.target.value) })}
                      />
                      <Input
                        placeholder="Location"
                        value={String(selectedConfig.location || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'location', e.target.value) })}
                      />
                      <Input
                        placeholder="Model"
                        value={String(selectedConfig.model || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'model', e.target.value) })}
                      />
                      <Input
                        placeholder="Auth Mode (service_account_json / adc)"
                        value={String(selectedConfig.authMode || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'authMode', e.target.value) })}
                      />
                      <div className="md:col-span-2">
                        <textarea
                          className="h-32 w-full rounded-md border border-zinc-700 bg-zinc-950 p-3 font-mono text-xs"
                          placeholder="Service account JSON"
                          value={String(selectedConfig.serviceAccountJson || '')}
                          onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'serviceAccountJson', e.target.value) })}
                        />
                      </div>
                    </div>
                  ) : null}

                  {selectedProviderKind === 'aws_bedrock' ? (
                    <div className="grid gap-3 md:grid-cols-2">
                      <Input
                        placeholder="Region"
                        value={String(selectedConfig.region || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'region', e.target.value) })}
                      />
                      <Input
                        placeholder="Model ID"
                        value={String(selectedConfig.modelId || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'modelId', e.target.value) })}
                      />
                      <Input
                        placeholder="Auth Mode (aws_sigv4 / bedrock_api_key)"
                        value={String(selectedConfig.authMode || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'authMode', e.target.value) })}
                      />
                      <Input
                        placeholder="Access Key ID"
                        value={String(selectedConfig.accessKeyId || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'accessKeyId', e.target.value) })}
                      />
                      <Input
                        placeholder="Secret Access Key"
                        value={String(selectedConfig.secretAccessKey || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'secretAccessKey', e.target.value) })}
                      />
                      <Input
                        placeholder="Session Token (optional)"
                        value={String(selectedConfig.sessionToken || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'sessionToken', e.target.value) })}
                      />
                      <Input
                        placeholder="API Key (if bedrock_api_key)"
                        value={String(selectedConfig.apiKey || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'apiKey', e.target.value) })}
                      />
                    </div>
                  ) : null}

                  {selectedProviderKind === 'databricks' ? (
                    <div className="grid gap-3 md:grid-cols-2">
                      <Input
                        placeholder="Workspace URL"
                        value={String(selectedConfig.workspaceUrl || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'workspaceUrl', e.target.value) })}
                      />
                      <Input
                        placeholder="Serving endpoint"
                        value={String(selectedConfig.servingEndpoint || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'servingEndpoint', e.target.value) })}
                      />
                      <Input
                        placeholder="Auth Mode (pat / oauth_m2m)"
                        value={String(selectedConfig.authMode || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'authMode', e.target.value) })}
                      />
                      <Input
                        placeholder="PAT / API Key"
                        value={String(selectedConfig.apiKey || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'apiKey', e.target.value) })}
                      />
                      <Input
                        placeholder="Client ID (oauth)"
                        value={String(selectedConfig.clientId || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'clientId', e.target.value) })}
                      />
                      <Input
                        placeholder="Client Secret (oauth)"
                        value={String(selectedConfig.clientSecret || '')}
                        onChange={(e) => setSelectedDraft({ configJson: updateConfigField(selectedDraft.configJson, 'clientSecret', e.target.value) })}
                      />
                    </div>
                  ) : null}

                  <details className="rounded-md border border-zinc-700 bg-zinc-950 p-2">
                    <summary className="cursor-pointer text-sm text-zinc-300">Advanced: Raw JSON</summary>
                    <div className="mt-2 flex justify-end">
                      <Button
                        type="button"
                        className="h-8 bg-zinc-700 px-2 text-xs text-zinc-100 hover:bg-zinc-600"
                        onClick={() => setRawJsonEditMode((prev) => !prev)}
                      >
                        {rawJsonEditMode ? 'Preview' : 'Edit JSON'}
                      </Button>
                    </div>
                    {rawJsonEditMode ? (
                      <textarea
                        className="mt-2 h-48 w-full rounded-md border border-zinc-700 bg-zinc-950 p-3 font-mono text-xs"
                        value={selectedDraft.configJson}
                        onChange={(e) => setSelectedDraft({ configJson: e.target.value })}
                      />
                    ) : (
                      <div
                        className="mt-2 overflow-auto rounded-md border border-zinc-700 bg-zinc-950 p-3 text-xs leading-6"
                        dangerouslySetInnerHTML={{ __html: prettyJsonHtml }}
                      />
                    )}
                  </details>
                  <p className="text-xs text-zinc-500">Structured fields are mapped into configJson. Save Changes persists all connection settings.</p>
                </div>
              ) : null}

              {activeTab === 'branding' ? (
                <div className="space-y-3">
                  {resolveAssetUrl(selectedProvider.logoUrl) ? (
                    <img
                      src={resolveAssetUrl(selectedProvider.logoUrl) || ''}
                      alt={selectedProvider.name}
                      className="h-16 w-16 rounded"
                    />
                  ) : (
                    <div className="text-sm text-zinc-500">No logo uploaded yet.</div>
                  )}
                  <label className="inline-flex cursor-pointer items-center rounded-md border border-zinc-600 bg-zinc-800 px-3 py-2 text-sm text-zinc-100 hover:bg-zinc-700">
                    Choose Logo File
                    <input
                      type="file"
                      accept="image/*"
                      className="sr-only"
                      onChange={(e) => setSelectedDraft({ logoFile: e.target.files?.[0] || null })}
                    />
                  </label>
                  <div className="text-sm text-zinc-400">
                    {selectedDraft.logoFile ? `Selected: ${selectedDraft.logoFile.name}` : 'No file selected'}
                  </div>
                  <p className="text-xs text-zinc-500">Allowed: image files up to 2MB.</p>
                </div>
              ) : null}

              {activeTab === 'danger' ? (
                <div className="rounded-lg border border-rose-900 bg-rose-950/30 p-4">
                  <p className="text-sm text-rose-200">Disabling this provider hides all its models from discovery.</p>
                  <Button
                    onClick={async () => {
                      await fetch('/api/admin/providers/enabled', {
                        method: 'POST',
                        headers: { 'content-type': 'application/json' },
                        body: JSON.stringify({ slug: selectedProvider.slug, enabled: !selectedProvider.enabled }),
                      });
                      await load();
                    }}
                    className="mt-3 bg-rose-500 text-black hover:bg-rose-400"
                  >
                    {selectedProvider.enabled ? 'Disable Now' : 'Enable Now'}
                  </Button>
                </div>
              ) : null}

              <div className="sticky bottom-0 flex items-center justify-between rounded-lg border border-zinc-700 bg-zinc-950/95 p-3">
                <div className="text-sm text-zinc-400">{message || (hasChanges ? 'Unsaved changes' : 'No pending changes')}</div>
                <div className="flex gap-2">
                  <Button
                    type="button"
                    className="bg-zinc-700 text-zinc-100 hover:bg-zinc-600"
                    disabled={!hasChanges || saving}
                    onClick={() => {
                      if (!selectedProvider) return;
                      setDrafts((prev) => ({ ...prev, [selectedProvider.slug]: toDraft(selectedProvider) }));
                      setMessage('Changes discarded.');
                    }}
                  >
                    Discard
                  </Button>
                  <Button type="button" disabled={!hasChanges || saving} onClick={saveProviderSettings}>
                    {saving ? 'Saving...' : 'Save Changes'}
                  </Button>
                </div>
              </div>
            </>
          )}
        </section>
      </div>
    </div>
  );
}
