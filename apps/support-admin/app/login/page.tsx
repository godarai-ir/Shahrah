"use client";

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    if (!res.ok) {
      setError('Invalid credentials');
      setLoading(false);
      return;
    }
    const nextPath =
      typeof window !== 'undefined'
        ? new URLSearchParams(window.location.search).get('next')
        : null;
    router.replace(nextPath || '/admin');
  };

  return (
    <div className="mx-auto max-w-md space-y-4 rounded-xl border border-zinc-800 bg-zinc-900 p-6">
      <h1 className="text-xl font-semibold">Admin Login</h1>
      <form className="space-y-3" onSubmit={onSubmit}>
        <Input placeholder="username" value={username} onChange={(e) => setUsername(e.target.value)} />
        <Input placeholder="password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
        {error ? <p className="text-sm text-red-400">{error}</p> : null}
        <Button disabled={loading}>{loading ? 'Signing in...' : 'Sign in'}</Button>
      </form>
    </div>
  );
}
