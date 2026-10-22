import { HeroWindow } from './HeroWindow';
import { MediaSlot } from './MediaSlot';
import { PageHero, ProductStage } from './patterns';

// Until the GodarChat screenshot or product tour exists, the first viewport shows a drawn product window.
// It crosses the panel's reading-start edge (right in RTL) so it reads as an object resting on the bench.
function HeroScene() {
  return (
    <ProductStage overhang layout="size-full p-6 sm:p-10">
      <div className="relative w-full max-w-sm lg:-ms-20 lg:me-auto">
        <HeroWindow />
      </div>
    </ProductStage>
  );
}

export function Hero() {
  return (
    <PageHero
      title={
        <>
          هوش مصنوعی برای کار <span className="text-site-accent">هر روز</span> تیم‌ها؛ بی‌گدار به آب نزنید.
        </>
      }
      lead="به هر تیم دستیاری بدهید که با دانش و ابزارهای خودش کار می‌کند. دسترسی، داده و هزینه دست سازمانتان می‌ماند."
      primary={{ href: '/onboarding', label: '۱۵ روز رایگان شروع کنید' }}
      secondary={{ href: '/demo', label: 'درخواست دمو' }}
      visual={
        <MediaSlot
          id="heroDemo"
          priority
          className="h-auto overflow-visible rounded-3xl sm:h-auto lg:h-full"
          fallback={<HeroScene />}
        />
      }
    />
  );
}
