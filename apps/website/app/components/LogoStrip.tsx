import { Reveal } from './Reveal';
import { MediaSlot } from './MediaSlot';

export function LogoStrip() {
  return (
    <section className="flex w-full items-center justify-center bg-white pb-20 pt-12 dark:bg-slate-950">
      <div className="mx-auto flex w-full max-w-6xl flex-col items-center justify-center gap-12 px-6 text-center">
        <Reveal className="w-full max-w-4xl">
          <MediaSlot id="logos" />
        </Reveal>
      </div>
    </section>
  );
}
