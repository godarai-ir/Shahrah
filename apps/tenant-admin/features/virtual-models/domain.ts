import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";

import type {
  TenantAdminUserLookup,
  TenantProviderAccountModel,
  VirtualModel,
  VirtualModelPrincipal,
  VirtualModelTarget,
} from "./api";

export const LIST_PAGE_SIZE = 50;
export const MAX_TARGETS = 3;

export type VirtualModelFormState = {
  id?: string;
  alias: string;
  routing_strategy: "priority" | "weighted";
  enabled: boolean;
  exposed_modes: string[];
  targets: VirtualModelTarget[];
  access: {
    managers: VirtualModelPrincipal[];
    users: VirtualModelPrincipal[];
  };
};

export const DEFAULT_TARGET: VirtualModelTarget = {
  order: 1,
  weight: 100,
  provider: "",
  model: "",
  policy: {
    timeout_sec: 60,
    retry: {
      attempts: 2,
      delay_ms: 100,
      status_codes: [408, 425, 429, 500, 502, 503, 504],
    },
    fallback: {
      status_codes: [429, 500, 502, 503, 504],
      candidate: true,
    },
  },
};

export const MODE_ENDPOINTS: Record<string, string> = {
  chat: "/v1/chat/completions, /v1/responses",
  completion: "/v1/completions",
  embedding: "/v1/embeddings",
  moderation: "/v1/moderations",
  image_generation: "/v1/images/generations",
  text_to_speech: "/v1/audio/speech",
  audio_transcription: "/v1/audio/transcriptions",
  audio_translation: "/v1/audio/translations",
  realtime: "/v1/realtime",
  rerank: "/v1/rerank",
  search: "/v1/search",
};

export function makeDefaultForm(everyoneLabel: string): VirtualModelFormState {
  return {
    alias: "",
    routing_strategy: "priority",
    enabled: true,
    exposed_modes: [],
    targets: [{ ...DEFAULT_TARGET }],
    access: {
      managers: [],
      users: [{ type: "everyone", display_name: everyoneLabel }],
    },
  };
}

export function parseStatusCodes(input: string): number[] {
  return input
    .split(/[,\s]+/)
    .map((value) => Number(value.trim()))
    .filter((value) => Number.isInteger(value) && value >= 100 && value <= 599);
}

export async function mapWithConcurrency<T, R>(
  items: readonly T[],
  concurrency: number,
  mapper: (item: T, index: number) => Promise<R>,
): Promise<R[]> {
  const results = new Array<R>(items.length);
  let nextIndex = 0;
  const workerCount = Math.min(Math.max(1, concurrency), items.length);

  await Promise.all(Array.from({ length: workerCount }, async () => {
    while (nextIndex < items.length) {
      const currentIndex = nextIndex;
      nextIndex += 1;
      results[currentIndex] = await mapper(items[currentIndex], currentIndex);
    }
  }));

  return results;
}

export function createCopyAlias(alias: string, existingAliases: Set<string>) {
  const normalized = alias.trim() || "virtual-model";
  const base = `copy_of_${normalized}`;
  if (!existingAliases.has(base.toLowerCase())) return base;

  let suffix = 2;
  while (existingAliases.has(`copy_of_${suffix}_${normalized}`.toLowerCase())) suffix += 1;
  return `copy_of_${suffix}_${normalized}`;
}

export function resolveUserLabel(user: TenantAdminUserLookup | undefined, fallbackID: string) {
  if (!user) return fallbackID;
  return user.display_name?.trim()
    || user.email?.trim()
    || (mobileAuthUIEnabled ? user.mobile?.trim() : "")
    || user.id
    || fallbackID;
}

export function normalizeTargets(targets: VirtualModelTarget[] | undefined): VirtualModelTarget[] {
  if (!targets || targets.length === 0) return [{ ...DEFAULT_TARGET }];
  return targets
    .map((target, index) => ({
      ...target,
      order: target.order || index + 1,
      weight: target.weight ?? 0,
      policy: {
        timeout_sec: target.policy?.timeout_sec ?? 60,
        retry: {
          attempts: target.policy?.retry?.attempts ?? 2,
          delay_ms: target.policy?.retry?.delay_ms ?? 100,
          status_codes: target.policy?.retry?.status_codes ?? [408, 425, 429, 500, 502, 503, 504],
        },
        fallback: {
          candidate: target.policy?.fallback?.candidate ?? true,
          status_codes: target.policy?.fallback?.status_codes ?? [429, 500, 502, 503, 504],
        },
      },
    }))
    .sort((left, right) => left.order - right.order);
}

export function normalizeAccess(
  access: VirtualModel["access"] | undefined,
  everyoneLabel: string,
): VirtualModelFormState["access"] {
  const managers = (access?.managers ?? []).filter(Boolean);
  const users = (access?.users ?? []).filter(Boolean);
  if (managers.length === 0 && users.length === 0) {
    return { managers: [], users: [{ type: "everyone", display_name: everyoneLabel }] };
  }
  return { managers, users };
}

type ValidationMessageKey =
  | "messages.aliasRequired"
  | "messages.aliasPattern"
  | "messages.targetsRequired"
  | "messages.targetsLimit"
  | "messages.providerRequired"
  | "messages.modelRequired"
  | "messages.duplicateOrder"
  | "messages.retryAttemptsRange"
  | "messages.retryDelayRange"
  | "messages.timeoutRange"
  | "messages.weightRange"
  | "messages.weightSum";

export function validateForm(
  form: VirtualModelFormState,
  translate: (key: ValidationMessageKey) => string,
) {
  if (!form.alias.trim()) return translate("messages.aliasRequired");
  if (!/^[a-zA-Z0-9_-]+$/.test(form.alias.trim())) return translate("messages.aliasPattern");
  if (form.targets.length === 0) return translate("messages.targetsRequired");
  if (form.targets.length > MAX_TARGETS) return translate("messages.targetsLimit");

  const seenOrders = new Set<number>();
  for (const target of form.targets) {
    if (!target.provider.trim()) return translate("messages.providerRequired");
    if (!target.model.trim()) return translate("messages.modelRequired");
    if (seenOrders.has(target.order)) return translate("messages.duplicateOrder");
    seenOrders.add(target.order);
    if (target.policy.retry.attempts < 0 || target.policy.retry.attempts > 10) return translate("messages.retryAttemptsRange");
    if (target.policy.retry.delay_ms < 0 || target.policy.retry.delay_ms > 10000) return translate("messages.retryDelayRange");
    if ((target.policy.timeout_sec ?? 60) < 1 || (target.policy.timeout_sec ?? 60) > 600) return translate("messages.timeoutRange");
    if (form.routing_strategy === "weighted" && (target.weight ?? 0) <= 0) return translate("messages.weightRange");
  }

  if (form.routing_strategy === "weighted") {
    const totalWeight = form.targets.reduce(
      (sum, target) => sum + (Number.isFinite(target.weight) ? Number(target.weight) : 0),
      0,
    );
    if (Math.abs(totalWeight - 100) > 0.000001) return translate("messages.weightSum");
  }
  return null;
}

export function rebalanceTargetsForWeighted(targets: VirtualModelTarget[]): VirtualModelTarget[] {
  if (targets.length === 0) return [{ ...DEFAULT_TARGET }];

  const base = Math.floor(100 / targets.length);
  let remainder = 100 - base * targets.length;
  return targets.map((target) => {
    const weight = base + (remainder > 0 ? 1 : 0);
    if (remainder > 0) remainder -= 1;
    return { ...target, weight };
  });
}

export function modesForProviderModel(model: TenantProviderAccountModel): string[] {
  return normalizeModeList(model.supported_modes ?? model.effective_capabilities ?? model.types ?? []);
}

export function normalizeModeList(values: string[] | undefined): string[] {
  const seen = new Set<string>();
  const normalized: string[] = [];
  for (const value of values ?? []) {
    const mode = value.trim();
    if (!mode || seen.has(mode)) continue;
    seen.add(mode);
    normalized.push(mode);
  }
  return normalized;
}

export function buildModeSummary(targets: VirtualModelTarget[], modeMap: Record<string, string[]>) {
  const selectedTargets = targets.filter((target) => target.model.trim());
  if (selectedTargets.length === 0) return { intersection: [], partialModes: [], ready: false };

  const targetModes = selectedTargets.map((target) => normalizeModeList(
    modeMap[target.model] ?? modeMap[canonicalModelFQN(target.model)],
  ));
  if (targetModes.some((modes) => modes.length === 0)) return { intersection: [], partialModes: [], ready: true };

  const intersection = targetModes.reduce((current, modes) => current.filter((mode) => modes.includes(mode)));
  const union = normalizeModeList(targetModes.flat());
  return { intersection, partialModes: union.filter((mode) => !intersection.includes(mode)), ready: true };
}

export function sameModelFQN(left: string, right: string): boolean {
  return canonicalModelFQN(left) === canonicalModelFQN(right);
}

export function canonicalModelFQN(value: string): string {
  const parts = value.trim().split(":");
  if (parts.length !== 3) return value.trim();
  const provider = parts[0].trim();
  const account = parts[1].trim();
  let model = parts[2].trim();
  const prefix = `${provider.toLowerCase()}/`;
  if (model.toLowerCase().startsWith(prefix)) model = model.slice(prefix.length).trim();
  if (!provider || !account || !model) return value.trim();
  return `${provider}:${account}:${model}`;
}
