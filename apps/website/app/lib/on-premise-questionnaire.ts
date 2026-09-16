// Shared definition of the on-premise readiness questionnaire. The client renders
// the wizard from it and the API route validates and formats the email from it,
// so both stay in sync when a question changes.

export type Option = { value: string; label: string };

export const yesNoUnsure: Option[] = [
  { value: 'yes', label: 'بله' },
  { value: 'no', label: 'خیر' },
  { value: 'unsure', label: 'نمی‌دانم' },
];

const resourceOptions: Option[] = [
  { value: 'available', label: 'موجود است' },
  { value: 'needs-provisioning', label: 'نیازمند تأمین' },
  { value: 'unsure', label: 'نمی‌دانم' },
];

export type ChoiceQuestion = {
  id: string;
  kind: 'choice';
  label: string;
  hint?: string;
  options: Option[];
  /** Optional free-text follow-up, shown when the answer is one of `showWhen`. */
  detail?: { label: string; placeholder?: string; showWhen: string[]; dir?: 'ltr' | 'rtl' };
  /** Only ask this question when another question has one of the given answers. */
  showIf?: { id: string; in: string[] };
};

export type RangeQuestion = {
  id: string;
  kind: 'range';
  label: string;
  hint?: string;
  options: Option[];
};

export type Question = ChoiceQuestion | RangeQuestion;

export type QuestionGroup = { title: string; questions: Question[] };

export type Step = {
  id: string;
  title: string;
  shortTitle: string;
  description: string;
  groups: QuestionGroup[];
};

export const contactFields = {
  organization: { label: 'نام سازمان / واحد بهره‌بردار', max: 200, required: true },
  contactName: { label: 'نام و نام خانوادگی مسئول فنی', max: 120, required: true },
  contactRole: { label: 'سمت', max: 120, required: false },
  email: { label: 'ایمیل کاری', max: 254, required: true },
  phone: { label: 'شماره تماس', max: 32, required: true },
} as const;

export type ContactFieldId = keyof typeof contactFields;

export const notesMax = 3000;
export const detailMax = 300;

export const questionnaireSteps: Step[] = [
  {
    id: 'models',
    title: 'مدل‌ها و مقیاس استفاده',
    shortTitle: 'مقیاس',
    description: 'یک تخمین تقریبی کافی است؛ بعداً همراه تیم ما دقیق‌ترش می‌کنیم.',
    groups: [
      {
        title: 'ظرفیت موردنیاز',
        questions: [
          {
            id: 'expectedUsers',
            kind: 'range',
            label: 'تعداد کاربران مورد انتظار',
            options: [
              { value: 'lt-50', label: 'کمتر از ۵۰' },
              { value: '50-200', label: '۵۰ تا ۲۰۰' },
              { value: '200-1000', label: '۲۰۰ تا ۱٬۰۰۰' },
              { value: '1000-5000', label: '۱٬۰۰۰ تا ۵٬۰۰۰' },
              { value: 'gt-5000', label: 'بیش از ۵٬۰۰۰' },
            ],
          },
        ],
      },
      {
        title: 'ارائه‌دهندگان مدل هوش مصنوعی',
        questions: [
          {
            id: 'providersDefined',
            kind: 'choice',
            label: 'Providerهای مدل هوش مصنوعی مشخص شده‌اند؟',
            options: yesNoUnsure,
            detail: { label: 'نام Providerها', placeholder: 'مثلاً OpenAI، Anthropic، مدل‌های داخلی', showWhen: ['yes'] },
          },
          {
            id: 'providerKeysReady',
            kind: 'choice',
            label: 'API Keyهای Providerها آماده است؟',
            options: yesNoUnsure,
          },
          {
            id: 'providerNetworkTested',
            kind: 'choice',
            label: 'دسترسی شبکه‌ای به Providerها تست شده است؟',
            options: yesNoUnsure,
          },
        ],
      },
    ],
  },
  {
    id: 'platform',
    title: 'Kubernetes و منابع سخت‌افزاری',
    shortTitle: 'زیرساخت',
    description: 'گدارAI روی Kubernetes اجرا می‌شود. اگر هنوز کلاستر ندارید، مشکلی نیست.',
    groups: [
      {
        title: 'Kubernetes',
        questions: [
          {
            id: 'k8sAvailable',
            kind: 'choice',
            label: 'Kubernetes در سازمان موجود است؟',
            options: yesNoUnsure,
            detail: { label: 'نسخه', placeholder: 'مثلاً 1.30', showWhen: ['yes'], dir: 'ltr' },
          },
          {
            id: 'k8sNamespace',
            kind: 'choice',
            label: 'امکان ایجاد Namespace اختصاصی وجود دارد؟',
            options: yesNoUnsure,
            showIf: { id: 'k8sAvailable', in: ['yes'] },
          },
          {
            id: 'k8sRegistry',
            kind: 'choice',
            label: 'امکان Deploy کردن Container Image از Registry وجود دارد؟',
            options: yesNoUnsure,
            showIf: { id: 'k8sAvailable', in: ['yes'] },
          },
          {
            id: 'k8sIngress',
            kind: 'choice',
            label: 'Ingress Controller در اختیار است؟',
            options: yesNoUnsure,
            detail: { label: 'نوع', placeholder: 'مثلاً NGINX، Traefik', showWhen: ['yes'], dir: 'ltr' },
            showIf: { id: 'k8sAvailable', in: ['yes'] },
          },
        ],
      },
      {
        title: 'منابع قابل تخصیص به گدارAI',
        questions: [
          {
            id: 'resourceCpu',
            kind: 'choice',
            label: 'CPU',
            options: resourceOptions,
            detail: { label: 'مقدار', placeholder: 'مثلاً 16 vCPU', showWhen: ['available', 'needs-provisioning'] },
          },
          {
            id: 'resourceRam',
            kind: 'choice',
            label: 'RAM',
            options: resourceOptions,
            detail: { label: 'مقدار', placeholder: 'مثلاً 64 GB', showWhen: ['available', 'needs-provisioning'] },
          },
          {
            id: 'resourceStorage',
            kind: 'choice',
            label: 'فضای ذخیره‌سازی',
            options: resourceOptions,
            detail: { label: 'مقدار', placeholder: 'مثلاً 500 GB SSD', showWhen: ['available', 'needs-provisioning'] },
          },
        ],
      },
    ],
  },
  {
    id: 'network',
    title: 'شبکه و امنیت',
    shortTitle: 'شبکه',
    description: 'دسترسی‌ها و الزامات امنیتی محیط استقرار.',
    groups: [
      {
        title: 'شبکه',
        questions: [
          { id: 'egressToProviders', kind: 'choice', label: 'دسترسی خروجی از محیط استقرار به API Providerها وجود دارد؟', options: yesNoUnsure },
          { id: 'firewallRules', kind: 'choice', label: 'امکان تعریف Firewall Rule موردنیاز وجود دارد؟', options: yesNoUnsure },
          { id: 'dns', kind: 'choice', label: 'رکورد DNS موردنیاز قابل ایجاد است؟', options: yesNoUnsure },
          {
            id: 'loadBalancer',
            kind: 'choice',
            label: 'Load Balancer / Reverse Proxy در دسترس است؟',
            options: yesNoUnsure,
            detail: { label: 'نوع', placeholder: 'مثلاً HAProxy، F5', showWhen: ['yes'], dir: 'ltr' },
          },
        ],
      },
      {
        title: 'امنیت',
        questions: [
          { id: 'tlsCertificate', kind: 'choice', label: 'امکان تهیه و نصب TLS Certificate وجود دارد؟', options: yesNoUnsure },
          {
            id: 'secretManagement',
            kind: 'choice',
            label: 'سیستم مدیریت Secret (مانند Vault یا K8s Secret) وجود دارد؟',
            options: yesNoUnsure,
          },
        ],
      },
    ],
  },
  {
    id: 'data',
    title: 'داده و احراز هویت',
    shortTitle: 'داده',
    description: 'سرویس‌های داده و سیستم هویت سازمانی که گدارAI به آن‌ها متصل می‌شود.',
    groups: [
      {
        title: 'پایگاه داده و Cache',
        questions: [
          {
            id: 'postgres',
            kind: 'choice',
            label: 'PostgreSQL در سازمان موجود است؟',
            options: yesNoUnsure,
            detail: { label: 'نسخه / نوع', placeholder: 'مثلاً 16، Patroni', showWhen: ['yes'], dir: 'ltr' },
          },
          {
            id: 'redis',
            kind: 'choice',
            label: 'Redis در سازمان موجود است؟',
            options: yesNoUnsure,
            detail: { label: 'نسخه / نوع', placeholder: 'مثلاً 7، Sentinel', showWhen: ['yes'], dir: 'ltr' },
          },
        ],
      },
      {
        title: 'احراز هویت',
        questions: [
          {
            id: 'identityProvider',
            kind: 'choice',
            label: 'سیستم احراز هویت سازمانی (AD / LDAP / Keycloak / SSO) وجود دارد؟',
            options: yesNoUnsure,
            detail: { label: 'نام سیستم', placeholder: 'مثلاً Active Directory', showWhen: ['yes'], dir: 'ltr' },
          },
          {
            id: 'ssoIntegration',
            kind: 'choice',
            label: 'در صورت نیاز، امکان اتصال گدارAI به SSO وجود دارد؟',
            options: [
              { value: 'yes', label: 'بله' },
              { value: 'no', label: 'خیر' },
              { value: 'needs-review', label: 'نیازمند بررسی' },
            ],
            showIf: { id: 'identityProvider', in: ['yes'] },
          },
        ],
      },
    ],
  },
  {
    id: 'operations',
    title: 'پایش و عملیات',
    shortTitle: 'عملیات',
    description: 'آخرین گروه پرسش‌ها؛ درباره ابزارها و تیم‌هایی که از سرویس نگهداری می‌کنند.',
    groups: [
      {
        title: 'پایش، لاگ و هشدار',
        questions: [
          {
            id: 'monitoring',
            kind: 'choice',
            label: 'سیستم مانیتورینگ Kubernetes / Application موجود است؟',
            options: yesNoUnsure,
            detail: { label: 'ابزار', placeholder: 'مثلاً Prometheus، Grafana', showWhen: ['yes'], dir: 'ltr' },
          },
          {
            id: 'centralLogging',
            kind: 'choice',
            label: 'سیستم مدیریت متمرکز Log موجود است؟',
            options: yesNoUnsure,
            detail: { label: 'ابزار', placeholder: 'مثلاً ELK، Loki', showWhen: ['yes'], dir: 'ltr' },
          },
          {
            id: 'alerting',
            kind: 'choice',
            label: 'سیستم Alert Management موجود است؟',
            options: yesNoUnsure,
            detail: { label: 'ابزار', placeholder: 'مثلاً Alertmanager، PagerDuty', showWhen: ['yes'], dir: 'ltr' },
          },
        ],
      },
      {
        title: 'تیم‌ها و فرآیندها',
        questions: [
          {
            id: 'infraTeam',
            kind: 'choice',
            label: 'تیم مسئول زیرساخت مشخص شده است؟',
            options: yesNoUnsure,
            detail: { label: 'واحد / مسئول', showWhen: ['yes'] },
          },
          {
            id: 'appTeam',
            kind: 'choice',
            label: 'تیم مسئول Application مشخص شده است؟',
            options: yesNoUnsure,
            detail: { label: 'واحد / مسئول', showWhen: ['yes'] },
          },
          { id: 'incidentProcess', kind: 'choice', label: 'فرآیند Incident Management وجود دارد؟', options: yesNoUnsure },
          { id: 'backupRestoreProcess', kind: 'choice', label: 'فرآیند Backup / Restore مشخص است؟', options: yesNoUnsure },
        ],
      },
    ],
  },
];

export type Answers = Record<string, string>;

export function detailKey(questionId: string) {
  return `${questionId}__detail`;
}

export function isQuestionVisible(question: Question, answers: Answers) {
  if (question.kind !== 'choice' || !question.showIf) {
    return true;
  }
  return question.showIf.in.includes(answers[question.showIf.id] ?? '');
}

export function isDetailVisible(question: Question, answers: Answers) {
  return question.kind === 'choice' && !!question.detail && question.detail.showWhen.includes(answers[question.id] ?? '');
}

export function stepQuestions(step: Step) {
  return step.groups.flatMap((group) => group.questions);
}

export const allQuestions: Question[] = questionnaireSteps.flatMap(stepQuestions);

export function isValidEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(value);
}

export function isValidPhone(value: string) {
  const digits = value.replace(/[۰-۹]/g, (d) => String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d))).replace(/[\s\-()+]/g, '');
  return /^\d{8,15}$/.test(digits);
}
