"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useFieldArray, useForm, type Resolver } from "react-hook-form";
import { z } from "zod";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

const rateSchema = z.object({
  usage: z.enum(["input", "output", "cache_read", "cache_write"]),
  unit: z.string().min(1),
  amount: z.string().regex(/^\d+(?:\.\d+)?$/),
  quantity: z.coerce.number().int().positive(),
  serviceTier: z.enum(["standard", "batch", "priority"]),
  resolution: z.string().optional(),
  tierFrom: z.coerce.number().int().nonnegative().optional(),
  tierTo: z.coerce.number().int().nonnegative().optional(),
});

const schema = z
  .object({
    providerSlug: z.string().min(1),
    canonicalId: z.string().min(1),
    slug: z.string().min(1),
    displayName: z.string().min(1),
    modelSeriesSlug: z.string().optional(),
    contextWindow: z.coerce.number().int().optional(),
    maxOutputTokens: z.coerce.number().int().optional(),
    region: z.string().min(1).default("global"),
    currency: z.string().min(3).default("USD"),
    rates: z.array(rateSchema).min(1),
    confirmZero: z.boolean().default(false),
  })
  .superRefine((value, context) => {
    if (
      value.rates.some((rate) => Number(rate.amount) === 0) &&
      !value.confirmZero
    ) {
      context.addIssue({
        code: "custom",
        path: ["confirmZero"],
        message: "Zero-priced rates require explicit confirmation.",
      });
    }
  });

type FormData = z.infer<typeof schema>;

const defaultRate: FormData["rates"][number] = {
  usage: "output",
  unit: "second",
  amount: "",
  quantity: 1,
  serviceTier: "standard",
  resolution: "",
};

export default function AdminModelCreatePage() {
  const { register, control, handleSubmit, formState, watch } =
    useForm<FormData>({
      resolver: zodResolver(schema as never) as Resolver<FormData>,
      defaultValues: {
        region: "global",
        currency: "USD",
        rates: [{ ...defaultRate }],
        confirmZero: false,
      },
    });
  const rates = useFieldArray({ control, name: "rates" });
  const watchedRates = watch("rates");

  const onSubmit = async (value: FormData) => {
    const response = await fetch("/api/admin/models", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        providerSlug: value.providerSlug,
        canonicalId: value.canonicalId,
        slug: value.slug,
        displayName: value.displayName,
        modelSeriesSlug: value.modelSeriesSlug || undefined,
        enabled: true,
        contextWindow: value.contextWindow,
        maxOutputTokens: value.maxOutputTokens,
        pricingStatus: value.rates.some((rate) => Number(rate.amount) > 0)
          ? "priced"
          : "unknown",
        pricingProfiles: [
          {
            region: value.region,
            currency: value.currency,
            rates: value.rates.map((rate) => ({
              ...rate,
              key: [
                rate.usage,
                rate.unit,
                rate.serviceTier,
                `resolution=${rate.resolution || "default"}`,
                `tier=${rate.tierFrom ?? "base"}`,
              ].join(":"),
              resolution: rate.resolution || undefined,
            })),
          },
        ],
      }),
    });
    if (!response.ok) throw new Error(await response.text());
    alert("Model upserted");
  };

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Upsert model</h1>
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="grid max-w-4xl gap-4 rounded-xl border bg-card p-4 text-card-foreground"
      >
        <div className="grid gap-3 md:grid-cols-2">
          <Input placeholder="provider slug" {...register("providerSlug")} />
          <Input
            placeholder="canonicalId (provider/model)"
            {...register("canonicalId")}
          />
          <Input placeholder="slug" {...register("slug")} />
          <Input placeholder="display name" {...register("displayName")} />
          <Input
            placeholder="model series slug (optional)"
            {...register("modelSeriesSlug")}
          />
          <Input
            placeholder="context window"
            type="number"
            {...register("contextWindow")}
          />
          <Input
            placeholder="max output tokens"
            type="number"
            {...register("maxOutputTokens")}
          />
        </div>

        <fieldset className="grid gap-3 rounded-lg border p-3">
          <legend className="px-1 font-medium">Pricing profile</legend>
          <div className="grid gap-3 md:grid-cols-2">
            <Input placeholder="region" {...register("region")} />
            <Input placeholder="currency" {...register("currency")} />
          </div>

          {rates.fields.map((field, index) => {
            const rate = watchedRates?.[index];
            const key = [
              rate?.usage,
              rate?.unit,
              rate?.serviceTier,
              `resolution=${rate?.resolution || "default"}`,
              `tier=${rate?.tierFrom ?? "base"}`,
            ].join(":");
            return (
              <div key={field.id} className="grid gap-2 rounded-lg bg-muted p-3">
                <div className="grid gap-2 md:grid-cols-4">
                  <select
                    className="h-9 rounded-md border bg-background px-3 text-sm"
                    {...register(`rates.${index}.usage`)}
                  >
                    <option value="input">input</option>
                    <option value="output">output</option>
                    <option value="cache_read">cache_read</option>
                    <option value="cache_write">cache_write</option>
                  </select>
                  <Input
                    placeholder="unit (second, token, request…)"
                    {...register(`rates.${index}.unit`)}
                  />
                  <Input
                    placeholder="amount"
                    inputMode="decimal"
                    {...register(`rates.${index}.amount`)}
                  />
                  <Input
                    placeholder="quantity"
                    type="number"
                    {...register(`rates.${index}.quantity`)}
                  />
                  <select
                    className="h-9 rounded-md border bg-background px-3 text-sm"
                    {...register(`rates.${index}.serviceTier`)}
                  >
                    <option value="standard">standard</option>
                    <option value="batch">batch</option>
                    <option value="priority">priority</option>
                  </select>
                  <Input
                    placeholder="resolution (optional)"
                    {...register(`rates.${index}.resolution`)}
                  />
                  <Input
                    placeholder="tier from"
                    type="number"
                    {...register(`rates.${index}.tierFrom`)}
                  />
                  <Input
                    placeholder="tier to"
                    type="number"
                    {...register(`rates.${index}.tierTo`)}
                  />
                </div>
                <div className="flex items-center justify-between gap-3">
                  <code className="text-xs text-muted-foreground" dir="ltr">
                    {key}
                  </code>
                  <Button
                    type="button"
                    className="border bg-background text-foreground hover:bg-muted"
                    onClick={() => rates.remove(index)}
                    disabled={rates.fields.length === 1}
                  >
                    Remove rate
                  </Button>
                </div>
              </div>
            );
          })}

          <Button
            type="button"
            className="border bg-background text-foreground hover:bg-muted"
            onClick={() => rates.append({ ...defaultRate })}
          >
            Add rate
          </Button>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" {...register("confirmZero")} />
            I confirm that any zero amount is intentional and does not mark the
            model free.
          </label>
          {formState.errors.confirmZero ? (
            <p className="text-sm text-destructive">
              {formState.errors.confirmZero.message}
            </p>
          ) : null}
        </fieldset>

        <Button disabled={formState.isSubmitting}>Save</Button>
      </form>
    </div>
  );
}
