import { cn } from '@/lib/utils';

export function Card({ children, className }: { children: React.ReactNode; className?: string }) {
  return <div className={cn('rounded-xl border border-zinc-800 bg-zinc-900/70 p-4', className)}>{children}</div>;
}

