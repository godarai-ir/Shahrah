"use client";

import Link from 'next/link';

export default function AdminModelsPage() {
  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Admin Models</h1>
      <Link href="/admin/models/new" className="inline-block rounded-md bg-cyan-500 px-3 py-2 text-sm font-medium text-black">Create / Upsert Model</Link>
    </div>
  );
}
