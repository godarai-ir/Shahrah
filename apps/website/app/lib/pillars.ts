import { Activity, ShieldCheck, Users, Wallet } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

export type PillarId = 'guard' | 'govern' | 'observe' | 'cost';

export type Pillar = {
  id: PillarId;
  title: string;
  summary: string;
  href: string;
  icon: LucideIcon;
};

// Each differentiator keeps its tone on every page it appears (DESIGN.md, The Color Means a Promise Rule).
// Class names are spelled out in full so Tailwind can find them.
export const toneSurface: Record<PillarId, string> = {
  guard: 'bg-tone-guard',
  govern: 'bg-tone-govern',
  observe: 'bg-tone-observe',
  cost: 'bg-tone-cost',
};

export const pillars: Pillar[] = [
  {
    id: 'guard',
    title: 'گاردریل معنایی',
    summary: 'داده‌ی محرمانه را از معنایش می‌شناسد.',
    href: '/features/guardrails',
    icon: ShieldCheck,
  },
  {
    id: 'govern',
    title: 'حاکمیت',
    summary: 'هر تیم به همان چیزی می‌رسد که باید.',
    href: '/features/governance',
    icon: Users,
  },
  {
    id: 'observe',
    title: 'مشاهده‌پذیری',
    summary: 'مصرف و هزینه‌ی هر تیم، و رد هر درخواست.',
    href: '/features/observability',
    icon: Activity,
  },
  {
    id: 'cost',
    title: 'مدیریت هزینه',
    summary: 'هزینه زیر سقفی که خودتان می‌گذارید.',
    href: '/features/cost-control',
    icon: Wallet,
  },
];

export function getPillar(id: PillarId): Pillar {
  const pillar = pillars.find((item) => item.id === id);

  if (!pillar) {
    throw new Error(`Unknown pillar: ${id}`);
  }

  return pillar;
}
