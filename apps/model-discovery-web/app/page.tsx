import { ModelsExplorer } from '@/components/catalog/models-explorer';

export default function HomePage() {
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Model Catalog</h1>
      <ModelsExplorer />
    </div>
  );
}
