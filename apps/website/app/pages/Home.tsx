import { Hero } from '../components/Hero';
import { LogoStrip } from '../components/LogoStrip';
import { Scenario } from '../components/Scenario';
import { QuoteCard } from '../components/QuoteCard';
import { Pillars } from '../components/Pillars';
import { Collab } from '../components/Collab';
import { Enterprise } from '../components/Enterprise';
import { TeamTabs } from '../components/TeamTabs';
import { GetStarted } from '../components/GetStarted';
import { HomeFAQ } from '../components/HomeFAQ';
import { ClosingCta } from '../components/ClosingCta';

export function Home() {
  return (
    <main>
      <Hero />
      <LogoStrip />
      <Scenario />
      <QuoteCard />
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
