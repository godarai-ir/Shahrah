import { HeroWindow } from './HeroWindow';
import { MediaSlot } from './MediaSlot';
import { PageHero, ProductStage } from './patterns';

// Until the GodarChat screenshot or product tour exists, the first viewport shows a drawn product window
// centred on a sand tile.
function HeroScene() {
  return (
    <ProductStage layout="size-full p-6 sm:p-10">
      <div className="relative w-full max-w-sm">
        <HeroWindow />
      </div>
    </ProductStage>
  );
}

// The page's one use of the tagline, on the brand's navy field.
export function Hero() {
  return (
    <PageHero
      tone="night"
      title="بی‌گدار به آب و AI نزنید!"
      lead="به هر تیم دستیاری بدهید که با دانش و ابزارهای خودش کار می‌کند. دسترسی، داده و هزینه دست شما می‌ماند."
      primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
      secondary={{ href: '/demo', label: 'درخواست دمو' }}
      visual={
        <MediaSlot
          id="heroDemo"
          priority
          className="h-auto overflow-visible rounded-card sm:h-auto lg:h-full"
          fallback={<HeroScene />}
        />
      }
    />
  );
}
