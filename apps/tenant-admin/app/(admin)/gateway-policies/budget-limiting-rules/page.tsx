import { redirect } from "next/navigation";

/** Rate and budget rules share one list now; kept so old links and bookmarks still land. */
export default function Page() {
  redirect("/gateway-policies?type=budget");
}
