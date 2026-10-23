import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  tutorialSidebar: [
    {
      type: 'category',
      label: 'شروع و نمای کلی',
      items: [
        'intro',
        'feature-overview',
        'quick-start',
      ],
    },
    {
      type: 'category',
      label: 'درگاه، مدل‌ها و اتصال',
      items: [
        'intro-to-gateway',
        'virtual-model',
        'model-identifiers',
        'making-llm-requests-via-gateway',
        'authentication',
        'guides/byok',
        'guides/byok-security',
        'request-headers',
        'native-sdk-support',
      ],
    },
    {
      type: 'category',
      label: 'ارائه‌دهندگان و مدل‌ها',
      link: {
        type: 'doc',
        id: 'provider-models-overview',
      },
      items: [
        'provider-models-catalog',
        'provider-models-concepts',
        'provider-models-compare',
        'providers/deepseek',
        'providers/moonshot',
        'providers/zai',
      ],
    },
    {
      type: 'category',
      label: 'APIهای متنی و داده',
      items: [
        'supported-apis/index',
        'supported-apis/chat-completions/overview',
        'supported-apis/chat-completions/multimodal',
        'supported-apis/chat-completions/tools',
        'supported-apis/chat-completions/structured-outputs',
        'supported-apis/chat-completions/caching-and-reasoning',
        'supported-apis/chat-completions/extended-thinking',
        'supported-apis/responses-api',
        'supported-apis/messages',
        'supported-apis/embeddings',
        'supported-apis/rerank',
        'supported-apis/moderation',
        'supported-apis/batch-api',
      ],
    },
    {
      type: 'category',
      label: 'APIهای تصویر و صوت',
      items: [
        'supported-apis/image-generation',
        'supported-apis/image-edit',
        'supported-apis/image-variation',
        'supported-apis/speech-to-text',
        'supported-apis/text-to-speech',
        'supported-apis/audio-translation',
      ],
    },
    {
      type: 'category',
      label: 'رجیستری پرامپت',
      link: {
        type: 'doc',
        id: 'prompt-registry',
      },
      items: [
        'prompt-registry-create',
        'prompt-registry-variables',
        'prompt-registry-run',
        'prompt-registry-advanced-settings',
        'prompt-registry-troubleshooting',
      ],
    },
    {
      type: 'category',
      label: 'کنترل مصرف و هزینه',
      items: [
        'guides/caching',
        'guides/rate-limiting',
        'guides/budget-limiting',
      ],
    },
    {
      type: 'category',
      label: 'مالی',
      items: [
        'financial/overview',
        'financial/rate-card',
        'budget-me',
        'cost-tracking',
      ],
    },
    {
      type: 'category',
      label: 'گاردریل‌های محتوا و امنیت',
      link: {
        type: 'doc',
        id: 'guardrails',
      },
      items: [
        'guardrails-capabilities',
        'guardrails-keyword-blocklist',
        'guardrails-regex',
        'guardrails-secrets-detection',
        'guardrails-pii-redaction',
        'guardrails-regex-replace',
        'guardrails-jwt-validation',
      ],
    },
    {
      type: 'category',
      label: 'گاردریل‌های قالب و اعتبارسنجی',
      items: [
        'guardrails-json-schema',
        'guardrails-json-keys',
        'guardrails-contains-code',
        'guardrails-not-null',
        'guardrails-inline-image-urls',
        'guardrails-sentence-count',
        'guardrails-word-count',
        'guardrails-character-count',
      ],
    },
    {
      type: 'category',
      label: 'مشاهده‌پذیری و پایش',
      items: [
        'observability-overview',
        'observability-logs',
        'observability-traces',
        'observability-model-metrics',
        'observability-logs-export',
        'observability-opentelemetry',
      ],
    },
    {
      type: 'category',
      label: 'کاربران و دسترسی',
      items: [
        'guides/users-access-overview',
        'guides/tenant-team-account-model',
        'guides/user-management',
        'guides/team-management',
        'guides/virtual-account-management',
        'guides/api-keys',
        'guides/api-key-rotation',
        'guides/roles-and-permissions',
      ],
    },
  ],
};

export default sidebars;
