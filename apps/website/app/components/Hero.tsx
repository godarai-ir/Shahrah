import { MediaSlot } from './MediaSlot';
import { Ripples, WaterLines } from './FordArt';
import { HeroWindow } from './HeroWindow';
import { PageHero, ProductStage } from './patterns';

// Until the product tour exists, the first viewport shows a drawn product window.
// A real video in `heroDemo` replaces it automatically.
function HeroScene() {
  return (
    <ProductStage
      layout="size-full p-6 sm:p-10"
      backdrop={
        <>
          <WaterLines className="text-site-accent/25" />
          <Ripples />
        </>
      }
    >
      <div className="relative w-full max-w-sm">
        <HeroWindow />
      </div>
    </ProductStage>
  );
}

export function Hero() {
  return (
    <PageHero
      title="دستیاری که شرکتتان را می‌شناسد"
      lead="گدارAI هوش مصنوعی را به ابزارهای روزمره‌ی تیم‌تان وصل می‌کند، از CRM تا بله و ایتا. کارهای تکراری با دستیار؛ تصمیم‌های مهم با خودتان."
      primary={{ href: '/demo', label: 'با ما حرف بزنید' }}
      secondary={{ href: '/onboarding', label: '۱۵ روز رایگان امتحان کنید' }}
      visual={<MediaSlot id="heroDemo" priority className="rounded-3xl" fallback={<HeroScene />} />}
    />
  );
}
