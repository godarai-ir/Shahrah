import type { Metadata } from 'next';
import { notFound } from 'next/navigation';

import { buildMetadata } from '@/app/lib/seo';
import { getSolution, solutionHref, solutions } from '@/app/lib/solutions';
import { SolutionPage } from '@/app/pages/solutions/SolutionPage';

type Props = { params: Promise<{ slug: string }> };

export const dynamicParams = false;

export function generateStaticParams() {
  return solutions.map((solution) => ({ slug: solution.slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const solution = getSolution(slug);
  if (!solution) return {};
  return buildMetadata({
    title: solution.seo.title,
    description: solution.seo.description,
    keywords: solution.seo.keywords,
    path: solutionHref(solution.slug),
  });
}

export default async function Page({ params }: Props) {
  const { slug } = await params;
  if (!getSolution(slug)) notFound();
  return <SolutionPage slug={slug} />;
}
