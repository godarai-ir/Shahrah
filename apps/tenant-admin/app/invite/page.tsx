import InviteAcceptPanel from "@/app/components/invite-accept-panel";

type InvitePageProps = {
  searchParams?: Promise<{
    token?: string;
    invitetoken?: string;
  }>;
};

export const dynamic = "force-dynamic";

export default async function InvitePage({ searchParams }: InvitePageProps) {
  const params = (await searchParams) ?? {};
  const token =
    (typeof params.token === "string" ? params.token.trim() : "") ||
    (typeof params.invitetoken === "string" ? params.invitetoken.trim() : "");

  return <InviteAcceptPanel token={token || undefined} />;
}
