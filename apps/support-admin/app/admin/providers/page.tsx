"use client";

import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { apiPost } from '@/lib/api';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

const schema = z.object({ token: z.string().min(1), slug: z.string().min(1), name: z.string().min(1) });

type FormData = z.infer<typeof schema>;

export default function AdminProvidersPage() {
  const { register, handleSubmit, formState, reset } = useForm<FormData>({ resolver: zodResolver(schema) });

  const onSubmit = async (values: FormData) => {
    await apiPost('/providers/admin/upsert', { slug: values.slug, name: values.name, status: 'active' }, values.token);
    reset({ ...values, slug: '', name: '' });
    alert('Provider upserted');
  };

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Admin Providers</h1>
      <form onSubmit={handleSubmit(onSubmit)} className="grid max-w-xl gap-3 rounded-xl border border-zinc-800 bg-zinc-900 p-4">
        <Input placeholder="Admin JWT" {...register('token')} />
        <Input placeholder="provider slug" {...register('slug')} />
        <Input placeholder="provider name" {...register('name')} />
        <Button disabled={formState.isSubmitting}>Upsert Provider</Button>
      </form>
    </div>
  );
}
