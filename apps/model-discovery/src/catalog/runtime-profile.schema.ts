import { z } from 'zod';

const thinkingModeSchema = z.enum(['adaptive', 'manual', 'disabled']);
const endpointSchema = z.enum([
  'messages',
  'count_tokens',
  'message_batches',
  'files',
  'models',
]);

export const modelRuntimeProfileSchema = z
  .object({
    version: z.number().int().positive(),
    endpoints: z.array(endpointSchema).min(1),
    thinking: z.object({
      defaultMode: thinkingModeSchema,
      allowedModes: z.array(thinkingModeSchema).min(1),
      manualDeprecated: z.boolean().default(false),
      effortValues: z
        .array(z.enum(['low', 'medium', 'high', 'max']))
        .default([]),
    }),
    assistantPrefill: z.boolean(),
    sampling: z.object({
      parameters: z.array(z.enum(['temperature', 'top_p', 'top_k'])),
      disallowedWithThinking: z
        .array(z.enum(['temperature', 'top_p', 'top_k']))
        .default([]),
    }),
    tools: z.object({
      custom: z.boolean(),
      parallel: z.boolean(),
      structuredOutput: z.boolean(),
    }),
    modalities: z.object({
      input: z.array(z.enum(['text', 'image', 'pdf'])).min(1),
      output: z.array(z.enum(['text'])).min(1),
    }),
    requestControls: z
      .object({
        serviceTierValues: z
          .array(z.enum(['auto', 'standard_only']))
          .default(['auto', 'standard_only']),
        inferenceGeoValues: z
          .array(z.enum(['global', 'us']))
          .default(['global']),
        speedValues: z.array(z.enum(['standard', 'fast'])).default([]),
      })
      .default({
        serviceTierValues: ['auto', 'standard_only'],
        inferenceGeoValues: ['global'],
        speedValues: [],
      }),
    betaHeaders: z.array(z.string().min(1)).default([]),
    limits: z.object({
      contextWindow: z.number().int().positive(),
      maxOutputTokens: z.number().int().positive(),
      messagesBytes: z
        .number()
        .int()
        .positive()
        .default(32 * 1024 * 1024),
      batchBytes: z
        .number()
        .int()
        .positive()
        .default(256 * 1024 * 1024),
      fileBytes: z
        .number()
        .int()
        .positive()
        .default(500 * 1024 * 1024),
    }),
  })
  .strict()
  .superRefine((profile, context) => {
    if (!profile.thinking.allowedModes.includes(profile.thinking.defaultMode)) {
      context.addIssue({
        code: 'custom',
        path: ['thinking', 'defaultMode'],
        message: 'defaultMode must be included in allowedModes',
      });
    }
    if (
      profile.thinking.defaultMode === 'disabled' &&
      profile.thinking.effortValues.length > 0
    ) {
      context.addIssue({
        code: 'custom',
        path: ['thinking', 'effortValues'],
        message: 'effortValues require an enabled thinking mode',
      });
    }
  });

export type ModelRuntimeProfile = z.infer<typeof modelRuntimeProfileSchema>;
