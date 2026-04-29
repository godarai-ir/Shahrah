'use client';

import { Hero } from '../components/Hero';
import { Problem } from '../components/Problem';
import { Solution } from '../components/Solution';
import { Features } from '../components/Features';
import { Developer } from '../components/Developer';
import { HowItWorks } from '../components/HowItWorks';
import { UseCases } from '../components/UseCases';
import { DashboardPreview } from '../components/DashboardPreview';
import { Trust } from '../components/Trust';

export function Home() {
  return (
    <main>
      <Hero />
      <Problem />
      <Solution />
      <Features />
      <Developer />
      <HowItWorks />
      <UseCases />
      <DashboardPreview />
      <Trust />
    </main>
  );
}

