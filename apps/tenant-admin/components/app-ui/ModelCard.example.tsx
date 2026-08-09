"use client";

import { ModelCard } from "@/components/app-ui/ModelCard";
import { mockModelCards } from "@/components/app-ui/ModelCard.mock";

export function ModelCardExampleGrid() {
  return (
    <section className="grid max-w-6xl grid-cols-1 gap-4 md:grid-cols-2">
      {mockModelCards.map((model) => (
        <ModelCard
          key={model.id}
          model={model}
          onViewDetails={(item) => console.log("view-details", item.id)}
          onCompare={(item) => console.log("compare", item.id)}
        />
      ))}
    </section>
  );
}
