export const RECENT_COMMANDS_STORAGE_KEY = "shell.recentCommands";
export const RECENT_COMMANDS_LIMIT = 5;

/** Stored ids only: titles and permissions are resolved again at read time. */
export function readRecentCommandIDs(): string[] {
  try {
    const raw = window.localStorage.getItem(RECENT_COMMANDS_STORAGE_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((id): id is string => typeof id === "string").slice(0, RECENT_COMMANDS_LIMIT);
  } catch {
    return [];
  }
}

export function pushRecentCommandID(id: string): void {
  const next = [id, ...readRecentCommandIDs().filter((existing) => existing !== id)].slice(0, RECENT_COMMANDS_LIMIT);
  try {
    window.localStorage.setItem(RECENT_COMMANDS_STORAGE_KEY, JSON.stringify(next));
  } catch {
    // Storage can be unavailable (private mode, quota); recents are only a convenience.
  }
}

/** Maps stored ids to commands the user can still run; unknown or no-longer-permitted ids drop silently. */
export function resolveRecentCommands<T>(ids: readonly string[], byId: ReadonlyMap<string, T>): T[] {
  return ids.map((id) => byId.get(id)).filter((command): command is T => Boolean(command));
}
