"use client";

import { zodResolver } from '@hookform/resolvers/zod';
import type { Resolver } from 'react-hook-form';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

const schema = z.object({
  providerSlug: z.string().min(1),
  canonicalId: z.string().min(1),
  slug: z.string().min(1),
  displayName: z.string().min(1),
  modelSeriesSlug: z.string().optional(),
  contextWindow: z.coerce.number().int().optional(),
  maxOutputTokens: z.coerce.number().int().optional(),
  inputPrice: z.coerce.number().optional(),
  outputPrice: z.coerce.number().optional(),
});

type FormData = z.infer<typeof schema>;

export default function AdminModelCreatePage() {
  const { register, handleSubmit, formState } = useForm<FormData>({
    resolver: zodResolver(schema) as Resolver<FormData>,
  });
  const onSubmit = async (v: FormData) => {
    await fetch('/api/admin/models', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({
        providerSlug: v.providerSlug,
        canonicalId: v.canonicalId,
        slug: v.slug,
        displayName: v.displayName,
        modelSeriesSlug: v.modelSeriesSlug || undefined,
        enabled: true,
        contextWindow: v.contextWindow,
        maxOutputTokens: v.maxOutputTokens,
        pricing: {
          currency: 'USD',
          inputPricePerMillionTokens: v.inputPrice,
          outputPricePerMillionTokens: v.outputPrice,
        },
      }),
    });
    alert('Model upserted');
  };

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Upsert Model</h1>
      <form onSubmit={handleSubmit(onSubmit)} className="grid max-w-2xl gap-3 rounded-xl border border-zinc-800 bg-zinc-900 p-4">
        <Input placeholder="provider slug" {...register('providerSlug')} />
        <Input placeholder="canonicalId (provider/model)" {...register('canonicalId')} />
        <Input placeholder="slug" {...register('slug')} />
        <Input placeholder="display name" {...register('displayName')} />
        <Input placeholder="model series slug (optional)" {...register('modelSeriesSlug')} />
        <Input placeholder="context window" type="number" {...register('contextWindow')} />
        <Input placeholder="max output tokens" type="number" {...register('maxOutputTokens')} />
        <Input placeholder="input price / 1M" type="number" step="0.01" {...register('inputPrice')} />
        <Input placeholder="output price / 1M" type="number" step="0.01" {...register('outputPrice')} />
        <Button disabled={formState.isSubmitting}>Save</Button>
      </form>
    </div>
  );
}
