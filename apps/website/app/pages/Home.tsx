import dynamic from "next/dynamic";
import { Hero } from '../components/Hero';
import { Problem } from '../components/Problem';
import { Solution } from '../components/Solution';
import { Trust } from '../components/Trust';

const Features = dynamic(() =>
  import('../components/Features').then((mod) => mod.Features),
);
const Developer = dynamic(() =>
  import('../components/Developer').then((mod) => mod.Developer),
);
const HowItWorks = dynamic(() =>
  import('../components/HowItWorks').then((mod) => mod.HowItWorks),
);
const UseCases = dynamic(() =>
  import('../components/UseCases').then((mod) => mod.UseCases),
);
const DashboardPreview = dynamic(() =>
  import('../components/DashboardPreview').then((mod) => mod.DashboardPreview),
);

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
