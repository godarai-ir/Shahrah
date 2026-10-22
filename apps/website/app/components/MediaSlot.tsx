import { Film, Image as ImageIcon, Play } from 'lucide-react';
import type { ReactNode } from 'react';
import type { LucideIcon } from 'lucide-react';
import Image from 'next/image';
import { homeMedia } from '@/app/lib/home-media';
import { cn } from '@/app/lib/utils';
import type { HomeMediaId, MediaAspect, MediaKind } from '@/app/lib/home-media';

const aspectClass: Record<MediaAspect, string> = {
  video: 'aspect-video',
  '4/3': 'aspect-4/3',
  '3/2': 'aspect-3/2',
  '5/6': 'aspect-5/6',
  square: 'aspect-square',
  strip: 'h-24',
  fill: 'h-80 sm:h-96 lg:h-full lg:min-h-96',
};

const kindLabel: Record<MediaKind, string> = {
  video: 'ویدیو',
  gif: 'گیف',
  image: 'تصویر',
};

const kindIcon: Record<MediaKind, LucideIcon> = {
  video: Play,
  gif: Film,
  image: ImageIcon,
};

type MediaSlotProps = {
  id: HomeMediaId;
  className?: string;
  priority?: boolean;
  /** Drawn instead of the placeholder while the slot has no `src`. */
  fallback?: ReactNode;
};

export function MediaSlot({ id, className = '', priority = false, fallback }: MediaSlotProps) {
  const media: {
    kind: MediaKind;
    aspect: MediaAspect;
    title: string;
    spec: string;
    alt: string;
    src?: string;
    poster?: string;
  } = homeMedia[id];
  const frame = cn('relative w-full overflow-hidden rounded-2xl', aspectClass[media.aspect], className);

  if (media.src && media.kind === 'video') {
    return (
      <div className={frame}>
        <video
          className="h-full w-full object-cover"
          src={media.src}
          poster={media.poster}
          aria-label={media.alt}
          autoPlay
          muted
          loop
          playsInline
          preload={priority ? 'auto' : 'metadata'}
        />
      </div>
    );
  }

  if (media.src) {
    return (
      <div className={frame}>
        <Image
          src={media.src}
          alt={media.alt}
          fill
          sizes="(min-width: 1024px) 50vw, 100vw"
          className="object-cover"
          priority={priority}
          unoptimized={media.kind === 'gif'}
        />
      </div>
    );
  }

  if (fallback) {
    return <div className={frame}>{fallback}</div>;
  }

  const Icon = kindIcon[media.kind];

  return (
    <div
      className={cn(frame, 'flex flex-col items-center justify-center gap-3 border-2 border-dashed border-site-line bg-site-panel p-4 text-center')}
    >
      <span className="flex size-12 items-center justify-center rounded-full bg-site-panel-deep text-site-ink-soft">
        <Icon className="size-5" />
      </span>
      <div className="min-w-0 max-w-md">
        <p className="mb-1 text-sm font-semibold text-site-ink">
          جای {kindLabel[media.kind]}: {media.title}
        </p>
        {media.aspect === 'strip' ? null : (
          <p className="text-xs leading-6 text-site-ink-soft">{media.spec}</p>
        )}
      </div>
    </div>
  );
}
