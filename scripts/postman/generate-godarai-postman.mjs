import { mkdir, writeFile } from 'node:fs/promises';
import path from 'node:path';

const OUT_DIR = path.resolve('tests/postman');
const COLLECTIONS_DIR = path.join(OUT_DIR, 'collections');
const PROVIDERS_URL = process.env.MODEL_DISCOVERY_PROVIDERS_URL ?? 'http://localhost:3002/api/v1/providers';

const fallbackProviders = [
  { slug: 'ai21', name: 'Ai21', status: 'active', enabled: true, modelCount: 1 },
  { slug: 'anthropic', name: 'Anthropic', status: 'active', enabled: true, modelCount: 11 },
  { slug: 'aws-bedrock', name: 'Aws Bedrock', status: 'active', enabled: true, modelCount: 137 },
  { slug: 'azure-ai-foundry', name: 'Azure Ai Foundry', status: 'active', enabled: true, modelCount: 68 },
  { slug: 'azure-open-ai', name: 'Azure Open Ai', status: 'active', enabled: true, modelCount: 124 },
  { slug: 'cerebras', name: 'Cerebras', status: 'active', enabled: true, modelCount: 4 },
  { slug: 'cohere', name: 'Cohere', status: 'active', enabled: true, modelCount: 28 },
  { slug: 'databricks', name: 'Databricks', status: 'active', enabled: true, modelCount: 21 },
  { slug: 'deepgram', name: 'Deepgram', status: 'active', enabled: true, modelCount: 143 },
  { slug: 'deepinfra', name: 'Deepinfra', status: 'active', enabled: true, modelCount: 163 },
  { slug: 'elevenlabs', name: 'Elevenlabs', status: 'active', enabled: true, modelCount: 10 },
  { slug: 'google-gemini', name: 'Google Gemini', status: 'active', enabled: true, modelCount: 41 },
  { slug: 'google-vertex', name: 'Google Vertex', status: 'active', enabled: true, modelCount: 301 },
  { slug: 'groq', name: 'Groq', status: 'active', enabled: true, modelCount: 16 },
  { slug: 'mistral-ai', name: 'Mistral Ai', status: 'active', enabled: true, modelCount: 55 },
  { slug: 'openai', name: 'Openai', status: 'active', enabled: true, modelCount: 107 },
  { slug: 'openrouter', name: 'Openrouter', status: 'active', enabled: true, modelCount: 696 },
  { slug: 'perplexity-ai', name: 'Perplexity Ai', status: 'active', enabled: true, modelCount: 5 },
  { slug: 'sambanova', name: 'Sambanova', status: 'active', enabled: true, modelCount: 13 },
  { slug: 'together-ai', name: 'Together Ai', status: 'active', enabled: true, modelCount: 238 },
  { slug: 'xai', name: 'Xai', status: 'active', enabled: true, modelCount: 72 },
  { slug: 'zai', name: 'Z.ai', status: 'active', enabled: true, modelCount: 4 },
];

const endpoints = [
  { key: 'models', method: 'GET', path: '/v1/models', implemented: true },
  { key: 'chat', method: 'POST', path: '/v1/chat/completions', implemented: true },
  { key: 'completion', method: 'POST', path: '/v1/completions', implemented: true },
  { key: 'responses', method: 'POST', path: '/v1/responses', implemented: true },
  { key: 'embeddings', method: 'POST', path: '/v1/embeddings', implemented: true },
  { key: 'moderations', method: 'POST', path: '/v1/moderations', implemented: true },
  { key: 'rerank', method: 'POST', path: '/v1/rerank', implemented: true },
  { key: 'images_generations', method: 'POST', path: '/v1/images/generations', implemented: true },
  { key: 'images_edits', method: 'POST', path: '/v1/images/edits', implemented: true },
  { key: 'images_variations', method: 'POST', path: '/v1/images/variations', implemented: true },
  { key: 'audio_transcriptions', method: 'POST', path: '/v1/audio/transcriptions', implemented: true },
  { key: 'audio_speech', method: 'POST', path: '/v1/audio/speech', implemented: true },
  { key: 'layout_parsing', method: 'POST', path: '/v1/layout_parsing', implemented: true },
  { key: 'batch', method: 'POST', path: '/v1/batches', implemented: false },
  { key: 'files', method: 'POST', path: '/v1/files', implemented: false },
  { key: 'video', method: 'POST', path: '/v1/videos', implemented: true },
];

function uid(seed) {
  let h = 2166136261;
  for (const ch of seed) {
    h ^= ch.charCodeAt(0);
    h = Math.imul(h, 16777619);
  }
  const hex = (h >>> 0).toString(16).padStart(8, '0');
  return `${hex}-${hex.slice(0, 4)}-${hex.slice(4, 8)}-${hex.slice(0, 4)}-${hex}${hex.slice(0, 4)}`;
}

function jsonBody(value) {
  return {
    mode: 'raw',
    raw: JSON.stringify(value, null, 2),
    options: { raw: { language: 'json' } },
  };
}

function rawJsonRequest(name, method, urlPath, body, testScript, prerequestScript = []) {
  return {
    name,
    event: [
      ...(prerequestScript.length ? [{ listen: 'prerequest', script: { type: 'text/javascript', exec: prerequestScript } }] : []),
      { listen: 'test', script: { type: 'text/javascript', exec: testScript } },
    ],
    request: {
      auth: { type: 'bearer', bearer: [{ key: 'token', value: '{{godarai_api_key}}', type: 'string' }] },
      method,
      header: [
        { key: 'Content-Type', value: 'application/json' },
        { key: 'Accept', value: 'application/json' },
        { key: 'x-godarai-tenant', value: '{{tenant_slug}}' },
      ],
      url: {
        raw: `{{base_url}}${urlPath}`,
        host: ['{{base_url}}'],
        path: urlPath.replace(/^\//, '').split('/'),
      },
      ...(body ? { body: jsonBody(body) } : {}),
    },
  };
}

function multipartRequest(name, urlPath, formdata, testScript, prerequestScript = []) {
  return {
    name,
    event: [
      ...(prerequestScript.length ? [{ listen: 'prerequest', script: { type: 'text/javascript', exec: prerequestScript } }] : []),
      { listen: 'test', script: { type: 'text/javascript', exec: testScript } },
    ],
    request: {
      auth: { type: 'bearer', bearer: [{ key: 'token', value: '{{godarai_api_key}}', type: 'string' }] },
      method: 'POST',
      header: [
        { key: 'Accept', value: 'application/json' },
        { key: 'x-godarai-tenant', value: '{{tenant_slug}}' },
      ],
      body: { mode: 'formdata', formdata },
      url: {
        raw: `{{base_url}}${urlPath}`,
        host: ['{{base_url}}'],
        path: urlPath.replace(/^\//, '').split('/'),
      },
    },
  };
}

const skipUnlessRunnable = [
  "const required = ['base_url', 'tenant_slug', 'godarai_api_key'];",
  "const missing = required.filter((key) => !pm.variables.get(key));",
  "if (missing.length) {",
  "  pm.collectionVariables.set('last_skip_reason', `Missing required variable(s): ${missing.join(', ')}`);",
  "  pm.execution.skipRequest();",
  "}",
];

function skipUnlessCapability(flagName, modelVarName) {
  return [
    ...skipUnlessRunnable,
    `if (String(pm.variables.get('${flagName}') || '').toLowerCase() !== 'true') {`,
    `  pm.collectionVariables.set('last_skip_reason', 'Capability ${flagName} is disabled for this provider/account.');`,
    '  pm.execution.skipRequest();',
    '}',
    `if (!pm.variables.get('${modelVarName}')) {`,
    `  pm.collectionVariables.set('last_skip_reason', 'Missing model variable: ${modelVarName}');`,
    '  pm.execution.skipRequest();',
    '}',
  ];
}

function basicJsonTests(expectedStatuses = [200]) {
  return [
    "const skip = pm.collectionVariables.get('last_skip_reason');",
    "if (skip) { pm.test(`Skipped: ${skip}`, () => pm.expect(true).to.eql(true)); pm.collectionVariables.unset('last_skip_reason'); }",
    `pm.test('status is expected', () => pm.expect(${JSON.stringify(expectedStatuses)}).to.include(pm.response.code));`,
    "pm.test('response is JSON when body is present', () => { if (pm.response.text()) pm.response.to.be.json; });",
    "const requestId = pm.response.headers.get('x-request-id') || pm.response.headers.get('x-godarai-request-id') || pm.response.headers.get('x-trace-id');",
    "if (requestId) pm.collectionVariables.set('last_request_id', requestId);",
    "try { const body = pm.response.json(); if (body.id) pm.collectionVariables.set('last_response_id', body.id); } catch (_) {}",
  ];
}

const chatTests = [
  ...basicJsonTests([200]),
  "const body = pm.response.json();",
  "pm.test('chat completion shape', () => { pm.expect(body).to.have.property('choices'); pm.expect(body.choices).to.be.an('array').that.is.not.empty; });",
  "if (body.id) pm.collectionVariables.set('last_chat_completion_id', body.id);",
];

const responsesTests = [
  ...basicJsonTests([200]),
  "const body = pm.response.json();",
  "pm.test('responses shape', () => { pm.expect(body.id || body.output || body.output_text).to.exist; });",
  "if (body.id) pm.collectionVariables.set('last_response_api_id', body.id);",
];

const completionTests = [
  ...basicJsonTests([200]),
  "const body = pm.response.json();",
  "pm.test('completion shape', () => { pm.expect(body).to.have.property('choices'); });",
];

const embeddingTests = [
  ...basicJsonTests([200]),
  "const body = pm.response.json();",
  "pm.test('embedding shape', () => { pm.expect(body.object).to.eql('list'); pm.expect(body.data).to.be.an('array').that.is.not.empty; });",
];

const moderationTests = [
  ...basicJsonTests([200]),
  "const body = pm.response.json();",
  "pm.test('moderation shape', () => { pm.expect(body.results).to.be.an('array').that.is.not.empty; });",
];

const imageTests = [
  ...basicJsonTests([200]),
  "const body = pm.response.json();",
  "pm.test('image shape', () => { pm.expect(body.data || body.output || body.images).to.exist; });",
];

const audioBinaryTests = [
  "const skip = pm.collectionVariables.get('last_skip_reason');",
  "if (skip) { pm.test(`Skipped: ${skip}`, () => pm.expect(true).to.eql(true)); pm.collectionVariables.unset('last_skip_reason'); }",
  "pm.test('status is expected', () => pm.expect([200]).to.include(pm.response.code));",
  "pm.test('audio response has content', () => pm.expect(pm.response.responseSize).to.be.above(0));",
];

const cacheTests = [
  ...basicJsonTests([200]),
  "const status = pm.response.headers.get('x-godarai-cache-status');",
  "if (status) pm.collectionVariables.set('last_cache_status', status);",
  "const cachedTrace = pm.response.headers.get('x-godarai-cached-trace-id');",
  "if (cachedTrace) pm.collectionVariables.set('last_cached_trace_id', cachedTrace);",
  "const similarity = pm.response.headers.get('x-godarai-cache-similarity-score');",
  "if (similarity) pm.collectionVariables.set('last_cache_similarity', similarity);",
  "pm.test('cache header is diagnostic when present', () => { if (status) pm.expect(['hit', 'miss', 'bypass', 'error']).to.include(status); else pm.expect(true).to.eql(true); });",
];

const negativeTests = [
  ...basicJsonTests([400, 401, 403, 404]),
  "pm.test('negative case returns error object when JSON', () => { try { const body = pm.response.json(); pm.expect(body.error || body.message).to.exist; } catch (_) { pm.expect(true).to.eql(true); } });",
];

function addCacheHeader(request, headerValue) {
  request.request.header.push({ key: 'x-godarai-cache-config', value: headerValue });
  return request;
}

function openAICollection(provider) {
  const variables = [
    ['provider_name', provider.name],
    ['provider_slug', provider.slug],
    ['provider_account_id', '{{openai_account_provider_id}}'],
    ['provider_account_name', '{{openai_account_provider_name}}'],
    ['model_fqn_default', '{{openai_chat_model_fqn}}'],
    ['model_fqn_chat', '{{openai_chat_model_fqn}}'],
    ['model_fqn_completion', '{{openai_chat_model_fqn}}'],
    ['model_fqn_responses', '{{openai_responses_model_fqn}}'],
    ['model_fqn_embeddings', '{{openai_embedding_model_fqn}}'],
    ['model_fqn_images', '{{openai_image_model_fqn}}'],
    ['model_fqn_moderations', '{{openai_moderation_model_fqn}}'],
    ['model_fqn_audio_speech', '{{openai_audio_speech_model_fqn}}'],
    ['model_fqn_audio_transcription', '{{openai_audio_transcription_model_fqn}}'],
    ['supports_chat', 'true'],
    ['supports_completion', 'true'],
    ['supports_responses', 'true'],
    ['supports_embeddings', 'true'],
    ['supports_images', 'false'],
    ['supports_audio_speech', 'true'],
    ['supports_audio_transcription', 'false'],
    ['supports_moderations', 'true'],
    ['supports_rerank', 'false'],
    ['supports_batch', 'false'],
    ['supports_files', 'false'],
    ['supports_video', 'false'],
  ].map(([key, value]) => ({ key, value, type: 'string' }));

  const chatBody = {
    model: '{{model_fqn_chat}}',
    messages: [{ role: 'user', content: 'Say "GodarAI smoke OK" in one short sentence.' }],
    temperature: 0,
    max_tokens: 32,
  };

  const collection = {
    info: {
      _postman_id: uid('godarai-openai'),
      name: 'GodarAI - OpenAI',
      description: 'Executable OpenAI-focused GodarAI/Shahrah gateway checks. Secrets are supplied only through environment variables.',
      schema: 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json',
    },
    variable: variables,
    event: [
      {
        listen: 'prerequest',
        script: {
          type: 'text/javascript',
          exec: [
            "if (!pm.collectionVariables.get('cache_namespace')) {",
            "  pm.collectionVariables.set('cache_namespace', `postman-${Date.now()}`);",
            "}",
          ],
        },
      },
    ],
    item: [
      {
        name: 'Gateway Health',
        item: [
          rawJsonRequest('Health', 'GET', '/health', null, basicJsonTests([200])),
          rawJsonRequest('Readiness', 'GET', '/ready', null, basicJsonTests([200])),
        ],
      },
      {
        name: 'Discovery',
        item: [
          rawJsonRequest('List Gateway Models', 'GET', '/v1/models', null, [
            ...basicJsonTests([200]),
            "const body = pm.response.json();",
            "pm.test('model list shape', () => { pm.expect(body.object).to.eql('list'); pm.expect(body.data).to.be.an('array'); });",
          ], skipUnlessRunnable),
          rawJsonRequest('Invalid Model FQN', 'POST', '/v1/chat/completions', {
            model: 'openai:missing-account:gpt-4o-mini',
            messages: [{ role: 'user', content: 'This should fail before provider call.' }],
            max_tokens: 8,
          }, negativeTests, skipUnlessRunnable),
        ],
      },
      {
        name: 'Text',
        item: [
          rawJsonRequest('Chat Completions - Smoke', 'POST', '/v1/chat/completions', chatBody, chatTests, skipUnlessCapability('supports_chat', 'model_fqn_chat')),
          rawJsonRequest('Completions - Legacy Smoke', 'POST', '/v1/completions', {
            model: '{{model_fqn_completion}}',
            prompt: 'Complete this phrase with two words: GodarAI debug',
            temperature: 0,
            max_tokens: 16,
          }, completionTests, skipUnlessCapability('supports_completion', 'model_fqn_completion')),
          rawJsonRequest('Responses - Smoke', 'POST', '/v1/responses', {
            model: '{{model_fqn_responses}}',
            input: 'Return the exact text: GodarAI responses OK',
            temperature: 0,
            max_output_tokens: 32,
          }, responsesTests, skipUnlessCapability('supports_responses', 'model_fqn_responses')),
        ],
      },
      {
        name: 'Embeddings',
        item: [
          rawJsonRequest('Embeddings - Single String', 'POST', '/v1/embeddings', {
            model: '{{model_fqn_embeddings}}',
            input: 'GodarAI embedding smoke test',
          }, embeddingTests, skipUnlessCapability('supports_embeddings', 'model_fqn_embeddings')),
          rawJsonRequest('Embeddings - Batch', 'POST', '/v1/embeddings', {
            model: '{{model_fqn_embeddings}}',
            input: ['GodarAI first vector', 'GodarAI second vector'],
          }, embeddingTests, skipUnlessCapability('supports_embeddings', 'model_fqn_embeddings')),
        ],
      },
      {
        name: 'Moderation',
        item: [
          rawJsonRequest('Moderations - Smoke', 'POST', '/v1/moderations', {
            model: '{{model_fqn_moderations}}',
            input: 'This is a harmless API smoke test.',
          }, moderationTests, skipUnlessCapability('supports_moderations', 'model_fqn_moderations')),
        ],
      },
      {
        name: 'Images',
        item: [
          rawJsonRequest('Images - Generations', 'POST', '/v1/images/generations', {
            model: '{{model_fqn_images}}',
            prompt: 'A small clean product demo icon for GodarAI API debugging',
            n: 1,
            size: '1024x1024',
          }, imageTests, skipUnlessCapability('supports_images', 'model_fqn_images')),
          multipartRequest('Images - Edits Placeholder', '/v1/images/edits', [
            { key: 'model', value: '{{model_fqn_images}}', type: 'text' },
            { key: 'prompt', value: 'Add a subtle API dashboard background.', type: 'text' },
            { key: 'image', type: 'file', src: '{{sample_image_path}}' },
          ], imageTests, skipUnlessCapability('supports_images', 'model_fqn_images')),
          multipartRequest('Images - Variations Placeholder', '/v1/images/variations', [
            { key: 'model', value: '{{model_fqn_images}}', type: 'text' },
            { key: 'image', type: 'file', src: '{{sample_image_path}}' },
          ], imageTests, skipUnlessCapability('supports_images', 'model_fqn_images')),
        ],
      },
      {
        name: 'Audio',
        item: [
          rawJsonRequest('Audio Speech - Smoke', 'POST', '/v1/audio/speech', {
            model: '{{model_fqn_audio_speech}}',
            voice: 'alloy',
            input: 'GodarAI audio smoke test',
            response_format: 'mp3',
          }, audioBinaryTests, skipUnlessCapability('supports_audio_speech', 'model_fqn_audio_speech')),
          multipartRequest('Audio Transcriptions Placeholder', '/v1/audio/transcriptions', [
            { key: 'model', value: '{{model_fqn_audio_transcription}}', type: 'text' },
            { key: 'file', type: 'file', src: '{{sample_audio_path}}' },
          ], basicJsonTests([200]), skipUnlessCapability('supports_audio_transcription', 'model_fqn_audio_transcription')),
        ],
      },
      {
        name: 'Cache',
        item: [
          addCacheHeader(rawJsonRequest('Simple Cache - Prime', 'POST', '/v1/chat/completions', chatBody, cacheTests, skipUnlessCapability('supports_chat', 'model_fqn_chat')), '{"type":"simple","ttl":600,"namespace":"{{cache_namespace}}"}'),
          addCacheHeader(rawJsonRequest('Simple Cache - Repeat', 'POST', '/v1/chat/completions', chatBody, cacheTests, skipUnlessCapability('supports_chat', 'model_fqn_chat')), '{"type":"simple","ttl":600,"namespace":"{{cache_namespace}}"}'),
          addCacheHeader(rawJsonRequest('Smart Cache - Prime', 'POST', '/v1/chat/completions', {
            model: '{{model_fqn_chat}}',
            messages: [{ role: 'user', content: 'Explain GodarAI cache diagnostics in one compact sentence.' }],
            temperature: 0,
            max_tokens: 32,
          }, cacheTests, skipUnlessCapability('supports_chat', 'model_fqn_chat')), '{"type":"smart","ttl":600,"namespace":"{{cache_namespace}}","similarity_threshold":0.85}'),
          addCacheHeader(rawJsonRequest('Smart Cache - Similar Prompt', 'POST', '/v1/chat/completions', {
            model: '{{model_fqn_chat}}',
            messages: [{ role: 'user', content: 'Briefly explain GodarAI response-cache diagnostics in a single sentence.' }],
            temperature: 0,
            max_tokens: 32,
          }, cacheTests, skipUnlessCapability('supports_chat', 'model_fqn_chat')), '{"type":"smart","ttl":600,"namespace":"{{cache_namespace}}","similarity_threshold":0.85}'),
        ],
      },
      {
        name: 'Not Implemented Yet',
        item: [
          rawJsonRequest('Batch Placeholder - Route Not Registered', 'POST', '/v1/batches', { input_file_id: '{{batch_input_file_id}}', endpoint: '/v1/chat/completions' }, basicJsonTests([404, 501]), skipUnlessRunnable),
          rawJsonRequest('Files Placeholder - Route Not Registered', 'POST', '/v1/files', { purpose: 'batch' }, basicJsonTests([404, 501]), skipUnlessRunnable),
        ],
      },
    ],
  };

  return collection;
}

function placeholderCollection(provider) {
  const title = `GodarAI - ${provider.name}`;
  const modelPrefix = provider.slug;
  const variables = [
    ['provider_name', provider.name],
    ['provider_slug', provider.slug],
    ['provider_model_count', String(provider.modelCount ?? '')],
    ['provider_enabled', 'false'],
    ['provider_api_key', ''],
    ['provider_account_id', ''],
    ['provider_account_name', ''],
    ['model_fqn_default', `${modelPrefix}:{{provider_account_name}}:{{provider_default_model_slug}}`],
    ['model_fqn_chat', `${modelPrefix}:{{provider_account_name}}:{{provider_chat_model_slug}}`],
    ['model_fqn_completion', `${modelPrefix}:{{provider_account_name}}:{{provider_completion_model_slug}}`],
    ['model_fqn_responses', `${modelPrefix}:{{provider_account_name}}:{{provider_responses_model_slug}}`],
    ['model_fqn_embeddings', `${modelPrefix}:{{provider_account_name}}:{{provider_embedding_model_slug}}`],
    ['model_fqn_images', `${modelPrefix}:{{provider_account_name}}:{{provider_image_model_slug}}`],
    ['model_fqn_video', `${modelPrefix}:{{provider_account_name}}:{{provider_video_model_slug}}`],
    ['provider_video_model_slug', provider.slug === 'xai' ? 'grok-imagine-video-1.5-lite' : ''],
    ['model_fqn_ocr', `${modelPrefix}:{{provider_account_name}}:{{provider_ocr_model_slug}}`],
    ['provider_ocr_model_slug', provider.slug === 'zai' ? 'glm-ocr' : ''],
    ['ocr_file_url', 'https://example.com/document.pdf'],
    ['supports_chat', 'false'],
    ['supports_completion', 'false'],
    ['supports_responses', 'false'],
    ['supports_embeddings', 'false'],
    ['supports_images', 'false'],
    ['supports_video', provider.slug === 'xai' ? 'true' : 'false'],
    ['supports_ocr', provider.slug === 'zai' ? 'true' : 'false'],
    ['supports_batch', 'false'],
  ].map(([key, value]) => ({ key, value, type: 'string' }));

  const placeholderSkip = [
    ...skipUnlessRunnable,
    "if (String(pm.collectionVariables.get('provider_enabled') || '').toLowerCase() !== 'true') {",
    "  pm.collectionVariables.set('last_skip_reason', 'Provider placeholder is disabled until credentials/account-provider/FQNs are configured.');",
    '  pm.execution.skipRequest();',
    '}',
  ];

  return {
    info: {
      _postman_id: uid(`godarai-${provider.slug}`),
      name: title,
      description: `Placeholder collection for ${provider.name}. Enable provider_enabled and set account/model variables after valid credentials are available.`,
      schema: 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json',
    },
    variable: variables,
    item: [
      {
        name: 'Discovery',
        item: [
          rawJsonRequest('List Gateway Models', 'GET', '/v1/models', null, basicJsonTests([200]), placeholderSkip),
        ],
      },
      {
        name: 'Text',
        item: [
          rawJsonRequest('Chat Completions Placeholder', 'POST', '/v1/chat/completions', {
            model: '{{model_fqn_chat}}',
            messages: [{ role: 'user', content: `Smoke test for ${provider.name}` }],
            max_tokens: 32,
          }, chatTests, placeholderSkip),
          rawJsonRequest('Responses Placeholder', 'POST', '/v1/responses', {
            model: '{{model_fqn_responses}}',
            input: `Smoke test for ${provider.name}`,
            max_output_tokens: 32,
          }, responsesTests, placeholderSkip),
          rawJsonRequest('Completions Placeholder', 'POST', '/v1/completions', {
            model: '{{model_fqn_completion}}',
            prompt: `Smoke test for ${provider.name}`,
            max_tokens: 32,
          }, completionTests, placeholderSkip),
        ],
      },
      {
        name: 'Specialized',
        item: [
          rawJsonRequest('Embeddings Placeholder', 'POST', '/v1/embeddings', {
            model: '{{model_fqn_embeddings}}',
            input: `Embedding smoke test for ${provider.name}`,
          }, embeddingTests, placeholderSkip),
          rawJsonRequest('Images Placeholder', 'POST', '/v1/images/generations', {
            model: '{{model_fqn_images}}',
            prompt: `API diagnostic image for ${provider.name}`,
            n: 1,
          }, imageTests, placeholderSkip),
          ...(provider.slug === 'zai' ? [rawJsonRequest('GLM-OCR Layout Parsing', 'POST', '/v1/layout_parsing', {
            model: '{{model_fqn_ocr}}',
            file: '{{ocr_file_url}}',
            start_page_id: 1,
            end_page_id: 1,
            return_crop_images: false,
            need_layout_visualization: false,
          }, [
            ...basicJsonTests([200]),
            "const body = pm.response.json();",
            "pm.test('layout parsing shape', () => { pm.expect(body.md_results || body.layout_details).to.exist; });",
          ], placeholderSkip)] : []),
          rawJsonRequest('Batch Placeholder - Route Not Registered', 'POST', '/v1/batches', {
            input_file_id: '{{batch_input_file_id}}',
            endpoint: '/v1/chat/completions',
          }, basicJsonTests([404, 501]), placeholderSkip),
          // Video jobs are async: this creates one 1 s 480p clip; poll GET /v1/videos/{id}.
          rawJsonRequest('Video - Create', 'POST', '/v1/videos', {
            model: '{{model_fqn_video}}',
            prompt: `Video diagnostic for ${provider.name}`,
            seconds: '1',
            size: '854x480',
          }, basicJsonTests([200]), skipUnlessCapability('supports_video', 'model_fqn_video')),
        ],
      },
    ],
  };
}

function environment() {
  const values = [
    ['base_url', 'http://localhost:4000', 'default'],
    ['tenant_slug', 'tenant1', 'default'],
    ['tenant_id', '3843f47a-d87d-40ac-9823-9620aae83584', 'default'],
    ['godarai_api_key', '', 'secret'],
    ['master_key', '', 'secret'],
    ['openai_api_key', '', 'secret'],
    ['openai_account_provider_id', 'bccca7ec-6ede-4268-8bcd-70ee701ce121', 'default'],
    ['openai_account_provider_name', 'openai-daring-voyager', 'default'],
    ['openai_chat_model_fqn', 'openai:openai-daring-voyager:gpt-5-nano', 'default'],
    ['openai_responses_model_fqn', 'openai:openai-daring-voyager:gpt-5-nano', 'default'],
    ['openai_embedding_model_fqn', 'openai:openai-daring-voyager:text-embedding-3-small', 'default'],
    ['openai_moderation_model_fqn', 'openai:openai-daring-voyager:omni-moderation-latest', 'default'],
    ['openai_image_model_fqn', 'openai:openai-daring-voyager:gpt-image-1', 'default'],
    ['openai_audio_speech_model_fqn', 'openai:openai-daring-voyager:gpt-4o-mini-tts', 'default'],
    ['openai_audio_transcription_model_fqn', 'openai:openai-daring-voyager:gpt-4o-mini-transcribe', 'default'],
    ['request_timeout_ms', '60000', 'default'],
    ['cache_enabled', 'true', 'default'],
    ['debug_trace_enabled', 'true', 'default'],
    ['sample_image_path', '', 'default'],
    ['sample_audio_path', '', 'default'],
    ['batch_input_file_id', '', 'default'],
  ].map(([key, value, type]) => ({ key, value, type, enabled: true }));

  return {
    id: uid('godarai-local-env'),
    name: 'GodarAI.local',
    values,
    _postman_variable_scope: 'environment',
    _postman_exported_using: 'Codex',
  };
}

function readme(providers) {
  const providerLines = providers.map((p) => `| ${p.slug} | ${p.name} | ${p.status ?? ''} | ${p.enabled} | ${p.modelCount ?? 0} |`).join('\n');
  const collectionLines = providers.map((p) => `- \`collections/GodarAI - ${p.slug}.postman_collection.json\``).join('\n');

  return `# GodarAI Postman Collections

این پوشه برای تست API layer، دیباگ و دموی GodarAI/Shahrah gateway ساخته شده است. کالکشن‌ها از providerهای زنده \`model-discovery\` در \`${PROVIDERS_URL}\` تولید شده‌اند و هیچ secret یا credential داخل فایل‌ها ذخیره نشده است.

## فایل‌ها

- \`GodarAI.local.postman_environment.json\`: template محیط محلی بدون secret.
- \`provider-manifest.json\`: snapshot providerها، endpointها و منبع تولید.
${collectionLines}

## اجرای سریع OpenAI

1. در Postman، environment و کالکشن \`GodarAI - OpenAI\` را import کنید.
2. مقدار \`godarai_api_key\` را با PAT tenant یا credential معتبر gateway پر کنید. برای local bootstrap می‌توانید مقدار \`MASTER_KEY\` را فقط در محیط شخصی خودتان به عنوان bearer token بگذارید، اما آن را در فایل‌ها ذخیره نکنید.
3. \`tenant_slug\` به صورت پیش‌فرض \`tenant1\` است.
4. اگر از tenant UI نیاز به ساخت Account-Provider یا خواندن FQNها داشتید، از \`http://tenant1.localhost:3000/\` استفاده کنید. در صورت نیاز به OTP، صندوق محلی روی \`http://localhost:8025/\` است.
5. درخواست \`Discovery / List Gateway Models\` را اجرا کنید و در صورت تفاوت، متغیرهای \`openai_*_model_fqn\` را با مدل‌های expose شده جایگزین کنید.

## متغیرهای مهم

- \`base_url\`: آدرس control-plane، پیش‌فرض \`http://localhost:4000\`.
- \`godarai_api_key\`: bearer token برای gateway.
- \`tenant_slug\`: tenant resolver header.
- \`openai_account_provider_id\` و \`openai_account_provider_name\`: account-provider فعلی OpenAI در local DB.
- \`openai_chat_model_fqn\`, \`openai_responses_model_fqn\`, \`openai_embedding_model_fqn\`: FQNهای مدل‌های قابل اجرا.
- \`sample_image_path\` و \`sample_audio_path\`: فقط برای requestهای multipart.

## Cache

کالکشن OpenAI دو سناریو دارد:

- Simple cache: دو request کاملا یکسان با header \`x-godarai-cache-config\` و namespace مشترک.
- Smart cache: دو prompt مشابه با \`type=smart\` و \`similarity_threshold=0.85\`.

assertionها عمدا diagnostic هستند: اگر headerهای \`x-godarai-cache-status\`, \`x-godarai-cached-trace-id\` یا \`x-godarai-cache-similarity-score\` وجود داشته باشند اعتبارسنجی می‌شوند؛ اگر gateway cache را bypass کند تست فقط وضعیت را ثبت می‌کند.

## Providerهای بدون credential

برای providerهای غیر OpenAI کالکشن placeholder ساخته شده است. تا وقتی \`provider_enabled=true\`، \`provider_account_name\` و FQNهای مدل پر نشوند، requestها با pre-request script skip می‌شوند و secretی hard-code نشده است.

## Provider Snapshot

| slug | name | status | enabled | modelCount |
| --- | --- | --- | --- | --- |
${providerLines}

## بازتولید

\`\`\`powershell
node scripts/postman/generate-godarai-postman.mjs
\`\`\`

اگر \`model-discovery\` در دسترس نباشد، generator از snapshot fallback همین فایل استفاده می‌کند.
`;
}

function manifest(providers) {
  return {
    generatedAt: new Date().toISOString(),
    providerSource: PROVIDERS_URL,
    providerCount: providers.length,
    gatewayBaseUrlVariable: 'base_url',
    auth: {
      bearerVariable: 'godarai_api_key',
      tenantHeader: 'x-godarai-tenant',
      tenantVariable: 'tenant_slug',
    },
    cacheHeaders: [
      'x-godarai-cache-config',
      'x-godarai-cache-status',
      'x-godarai-cached-trace-id',
      'x-godarai-cache-similarity-score',
    ],
    endpoints,
    providers: providers.map(({ slug, name, status, enabled, modelCount, docsUrl }) => ({
      slug,
      name,
      status,
      enabled,
      modelCount,
      docsUrl,
      collection: `collections/GodarAI - ${slug}.postman_collection.json`,
      executionStatus: slug === 'openai' ? 'executable-template' : 'placeholder-no-credentials',
    })),
  };
}

async function loadProviders() {
  try {
    const response = await fetch(PROVIDERS_URL);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const providers = await response.json();
    if (!Array.isArray(providers) || providers.length === 0) throw new Error('empty provider list');
    return providers;
  } catch (error) {
    console.warn(`[postman] Provider API unavailable (${error.message}); using fallback snapshot.`);
    return fallbackProviders;
  }
}

async function main() {
  const providers = (await loadProviders()).sort((a, b) => a.slug.localeCompare(b.slug));
  await mkdir(COLLECTIONS_DIR, { recursive: true });

  for (const provider of providers) {
    const collection = provider.slug === 'openai' ? openAICollection(provider) : placeholderCollection(provider);
    await writeFile(path.join(COLLECTIONS_DIR, `GodarAI - ${provider.slug}.postman_collection.json`), `${JSON.stringify(collection, null, 2)}\n`);
  }

  await writeFile(path.join(OUT_DIR, 'GodarAI.local.postman_environment.json'), `${JSON.stringify(environment(), null, 2)}\n`);
  await writeFile(path.join(OUT_DIR, 'provider-manifest.json'), `${JSON.stringify(manifest(providers), null, 2)}\n`);
  await writeFile(path.join(OUT_DIR, 'README.md'), readme(providers));

  console.log(`[postman] Generated ${providers.length} collections in ${COLLECTIONS_DIR}`);
}

await main();
