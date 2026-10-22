import { Hero } from '../components/Hero';
import { Scenario } from '../components/Scenario';
import { Pillars } from '../components/Pillars';
import { Collab } from '../components/Collab';
import { Enterprise } from '../components/Enterprise';
import { TeamTabs } from '../components/TeamTabs';
import { GetStarted } from '../components/GetStarted';
import { HomeFAQ } from '../components/HomeFAQ';
import { ClosingCta } from '../components/ClosingCta';

// Customer logos and quotes are left out until real material exists: LogoStrip and
// QuoteCard stay in the codebase and return here when there is something true to show.
export function Home() {
  return (
    <main className="bg-site-ground selection:bg-site-panel-deep">
      <Hero />
      <Scenario />
      <Pillars />
      <Collab />
      <Enterprise />
      <TeamTabs />
      <GetStarted />
      <HomeFAQ />
      <ClosingCta />
    </main>
  );
}
