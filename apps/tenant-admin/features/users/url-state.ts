export type MemberStatusFilter = "all" | "active" | "deactivated";

export type UsersURLState = {
  /** Free-text search, matched against identity and display name. */
  q: string;
  /** Role name from the tenant role catalog, or "" for every role. */
  role: string;
  status: MemberStatusFilter;
  /** Member whose detail drawer is open. */
  member: string;
};

const MEMBER_STATUSES = new Set<MemberStatusFilter>(["all", "active", "deactivated"]);

export function parseUsersURLState(params: URLSearchParams): UsersURLState {
  const rawStatus = params.get("status") as MemberStatusFilter | null;

  return {
    q: params.get("q")?.trim() ?? "",
    role: params.get("role")?.trim() ?? "",
    status: rawStatus && MEMBER_STATUSES.has(rawStatus) ? rawStatus : "all",
    member: params.get("member")?.trim() ?? "",
  };
}

export function patchUsersURL(
  current: URLSearchParams,
  patch: Record<string, string | null | undefined>,
): string {
  const next = new URLSearchParams(current.toString());
  for (const [key, value] of Object.entries(patch)) {
    const serialized = value?.trim();
    // "all" is the default status, so it stays out of the URL.
    if (serialized && serialized !== "all") next.set(key, serialized);
    else next.delete(key);
  }
  return next.toString();
}

export function hasActiveUserFilters(state: UsersURLState): boolean {
  return state.q !== "" || state.role !== "" || state.status !== "all";
}
