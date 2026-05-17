import { redirect } from 'next/navigation';

export default async function CategoryPage({ params }: { params: Promise<{ categorySlug: string }> }) {
  const { categorySlug } = await params;
  redirect(`/models?category=${categorySlug}`);
}
