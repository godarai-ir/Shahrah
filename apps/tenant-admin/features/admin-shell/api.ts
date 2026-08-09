import { requestJSON } from "@/lib/http/api-client";

export async function logout(): Promise<void> {
  await requestJSON("/api/auth/logout", { method: "POST" });
}
