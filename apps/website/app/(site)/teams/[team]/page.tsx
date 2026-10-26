import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { TeamPage } from '@/app/pages/teams/TeamPage';
import { buildMetadata } from '@/app/lib/seo';
import { getTeam, teams } from '@/app/lib/teams';

type Params = { params: Promise<{ team: string }> };

export const dynamicParams = false;

export function generateStaticParams() {
  return teams.map((team) => ({ team: team.id }));
}

export async function generateMetadata({ params }: Params): Promise<Metadata> {
  const team = getTeam((await params).team);

  if (!team) {
    return {};
  }

  return buildMetadata({
    title: `هوش مصنوعی برای تیم ${team.label}`,
    description: team.seo,
    path: `/teams/${team.id}`,
  });
}

export default async function Page({ params }: Params) {
  const team = getTeam((await params).team);

  if (!team) {
    notFound();
  }

  return <TeamPage team={team} />;
}
