import { ClosingCta } from '../components/ClosingCta';
import { Enterprise } from '../components/Enterprise';
import { GetStarted } from '../components/GetStarted';
import { Hero } from '../components/Hero';
import { HomeFAQ } from '../components/HomeFAQ';
import { Pillars } from '../components/Pillars';
import { Scenario } from '../components/Scenario';
import { TeamTabs } from '../components/TeamTabs';
import { TwoWays } from '../components/TwoWays';

// Customer logos and quotes are left out until real material exists: LogoStrip and
// QuoteCard stay in the codebase and return here when there is something true to show.
export function Home() {
  return (
    <main className="bg-site-ground selection:bg-site-panel-deep">
      <Hero />
      <Scenario />
      <TwoWays />
      <Pillars />
      <TeamTabs />
      <Enterprise />
      <GetStarted />
      <HomeFAQ />
      <ClosingCta />
    </main>
  );
}
