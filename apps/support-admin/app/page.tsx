import Link from 'next/link';

export default function Home() {
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-semibold">Model Catalog Platform</h1>
      <p className="text-zinc-400">Production-focused catalog experience for discovering, comparing, and managing LLM models.</p>
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <Link href="/models" className="rounded-xl border border-zinc-800 bg-zinc-900 p-4 hover:border-cyan-500">Browse Models</Link>
        <Link href="/providers" className="rounded-xl border border-zinc-800 bg-zinc-900 p-4 hover:border-cyan-500">Provider Pages</Link>
        <Link href="/compare" className="rounded-xl border border-zinc-800 bg-zinc-900 p-4 hover:border-cyan-500">Model Compare</Link>
        <Link href="/admin" className="rounded-xl border border-zinc-800 bg-zinc-900 p-4 hover:border-cyan-500">Admin Panel</Link>
      </div>
    </div>
  );
}
