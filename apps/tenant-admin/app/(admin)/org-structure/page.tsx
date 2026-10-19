import { redirect } from "next/navigation";

export const dynamic = "force-dynamic";

// Legacy URL kept so old bookmarks and docs links land on the teams page.
export default function LegacyOrgStructurePage() {
  redirect("/teams");
}
