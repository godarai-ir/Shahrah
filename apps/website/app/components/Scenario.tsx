import { Pebbles, WaterLines } from './FordArt';
import { FeatureRow, ProductStage } from './patterns';
import { WorkflowCard } from './WorkflowCard';

export function Scenario() {
  return (
    <FeatureRow
      id="scenario"
      reverse
      title="از سابقه‌ی مشتری تا پیش‌نویس جواب"
      copy="مشتری می‌نویسد فاکتورش اشتباه است. دستیار می‌فهمد درباره‌ی کدام سفارش حرف می‌زند، فاکتور و سابقه‌اش را از سیستم‌های شما بیرون می‌کشد و جواب را می‌نویسد. همکار پشتیبانی نگاهی می‌اندازد و می‌فرستد. همین."
      visual={
        <ProductStage
          layout="w-full max-w-lg p-8 pb-24 sm:p-12 sm:pb-28"
          backdrop={<WaterLines className="text-site-accent/20" />}
        >
          <div className="relative w-full">
            <WorkflowCard />
          </div>
          <Pebbles className="absolute bottom-3 start-6 w-56 sm:w-64" />
        </ProductStage>
      }
    />
  );
}
