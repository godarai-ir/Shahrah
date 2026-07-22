import { cn } from '@/lib/utils';

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'default' | 'secondary' | 'outline' | 'ghost' | 'destructive';
  size?: 'default' | 'sm' | 'icon';
};

const variantClassNames: Record<NonNullable<ButtonProps['variant']>, string> = {
  default: 'bg-primary text-primary-foreground hover:bg-primary/90',
  secondary: 'border border-primary-soft bg-primary-soft text-primary-soft-foreground hover:bg-primary-soft/80',
  outline: 'border border-border bg-background text-foreground hover:bg-accent hover:text-accent-foreground',
  ghost: 'bg-transparent text-foreground hover:bg-accent hover:text-accent-foreground',
  destructive: 'bg-destructive text-foreground hover:bg-destructive/90',
};

const sizeClassNames: Record<NonNullable<ButtonProps['size']>, string> = {
  default: 'px-3 py-2',
  sm: 'h-8 px-2.5 text-xs',
  icon: 'size-9 p-0',
};

export function Button({
  className,
  variant = 'default',
  size = 'default',
  ...props
}: ButtonProps) {
  return (
    <button
      {...props}
      data-slot="button"
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-60 [&_svg]:size-4 [&_svg]:shrink-0',
        variantClassNames[variant],
        sizeClassNames[size],
        className,
      )}
    />
  );
}

