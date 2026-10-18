import { WorkplaceCta } from '../components/workplace/WorkplaceCta';
import { WorkplaceGuide } from '../components/workplace/WorkplaceGuide';
import { AnnounceBar, WorkplaceHero } from '../components/workplace/WorkplaceHero';
import { WorkplaceProof } from '../components/workplace/WorkplaceProof';
import { WorkplaceSuite } from '../components/workplace/WorkplaceSuite';
import { WorkplaceValues } from '../components/workplace/WorkplaceValues';

export function Workplace() {
  return (
    <main className="bg-zinc-100 text-zinc-900 dark:bg-slate-950 dark:text-white">
      <AnnounceBar />
      <WorkplaceHero />
      <WorkplaceProof />
      <WorkplaceSuite />
      <WorkplaceValues />
      <WorkplaceGuide />
      <WorkplaceCta />
    </main>
  );
}
