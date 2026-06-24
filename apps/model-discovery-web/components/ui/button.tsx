import { cn } from '@/lib/utils';

export function Button(props: React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button {...props} className={cn('rounded-md bg-cyan-500 px-3 py-2 text-sm font-medium text-black hover:bg-cyan-400 disabled:opacity-60', props.className)} />;
}

