import { Film, Image as ImageIcon, Play } from 'lucide-react';
import type { ReactNode } from 'react';
import type { LucideIcon } from 'lucide-react';
import Image from 'next/image';
import { homeMedia } from '@/app/lib/home-media';
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
  const frame = `relative w-full overflow-hidden rounded-2xl ${aspectClass[media.aspect]} ${className}`;

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
      className={`${frame} flex flex-col items-center justify-center gap-3 border-2 border-dashed border-slate-300 bg-white/70 p-4 text-center dark:border-slate-700 dark:bg-slate-900/50`}
    >
      <span className="flex h-12 w-12 items-center justify-center rounded-full bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400">
        <Icon className="h-5 w-5" />
      </span>
      <div className="min-w-0 max-w-md">
        <p className="mb-1 text-sm font-bold text-slate-700 dark:text-gray-200">
          جای {kindLabel[media.kind]}: {media.title}
        </p>
        {media.aspect === 'strip' ? null : (
          <p className="text-xs leading-6 text-slate-500 dark:text-gray-400">{media.spec}</p>
        )}
      </div>
    </div>
  );
}
