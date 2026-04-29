import type { ReactNode } from 'react';
import { SiteShell } from '@/app/components/SiteShell';

export default function SiteLayout({ children }: { children: ReactNode }) {
  return <SiteShell>{children}</SiteShell>;
}
