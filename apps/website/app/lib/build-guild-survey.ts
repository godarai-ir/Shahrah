export type SurveyOption = { value: string; label: string };

export type BuildGuildAnswers = {
  fullName: string;
  mobile: string;
  role: string;
  roleOther: string;
  orgSize: string;
  challenges: string[];
  challengesOther: string;
  aiManagement: string;
  aiManagementOther: string;
};

export const surveyStorageKey = 'godarai-build-guild-survey-v1';

export const otherValue = 'other';
export const maxChallenges = 3;
export const textMax = 120;

export const emptyAnswers: BuildGuildAnswers = {
  fullName: '',
  mobile: '',
  role: '',
  roleOther: '',
  orgSize: '',
  challenges: [],
  challengesOther: '',
  aiManagement: '',
  aiManagementOther: '',
};

export const surveyQuestions = {
  contact: {
    label: 'نام و شماره همراه',
    fullName: 'نام و نام خانوادگی',
    mobile: 'شماره همراه',
  },
  role: {
    label: 'کدام گزینه به نقش فعلی شما در سازمان نزدیک‌تر است؟',
    options: [
      { value: 'cto', label: 'CTO / مدیر فنی' },
      { value: 'engineer', label: 'AI Engineer / Software Engineer / Data Engineer' },
      { value: 'product', label: 'Product Manager / AI Product Manager' },
      { value: 'devops', label: 'DevOps Engineer' },
      { value: 'ceo', label: 'مدیر عامل / بنیان‌گذار' },
      { value: 'qa', label: 'QA Engineer / Compliance Engineer' },
      { value: otherValue, label: 'سایر' },
    ],
  },
  orgSize: {
    label: 'جمعیت سازمان شما چند نفر است؟',
    options: [
      { value: '1-10', label: '۱ تا ۱۰ نفر' },
      { value: '11-50', label: '۱۱ تا ۵۰ نفر' },
      { value: '51-200', label: '۵۱ تا ۲۰۰ نفر' },
      { value: '201-1000', label: '۲۰۱ تا ۱۰۰۰ نفر' },
      { value: '1000+', label: 'بیشتر از ۱۰۰۰ نفر' },
    ],
  },
  challenges: {
    label: 'بزرگ‌ترین چالش‌های استفاده از AI در سازمان شما چیست؟',
    options: [
      { value: 'cost-control', label: 'کنترل و پیش‌بینی هزینه‌ها' },
      { value: 'usage-limits', label: 'تعیین سقف و کنترل مصرف کاربران/تیم‌ها' },
      { value: 'usage-reporting', label: 'شفافیت و گزارش‌گیری از نحوهٔ مصرف' },
      { value: 'data-security', label: 'امنیت و محرمانگی داده‌ها' },
      { value: 'access-keys', label: 'مدیریت دسترسی‌ها و کلیدهای API' },
      { value: 'model-selection', label: 'انتخاب مدل مناسب برای هر کاربرد' },
      { value: 'procurement', label: 'فرآیند خرید، پرداخت و تأمین سرویس‌های خارجی' },
      { value: 'non-technical-ui', label: 'نبود رابط کاربری ساده برای کاربران غیرفنی' },
      { value: otherValue, label: 'سایر' },
    ],
  },
  aiManagement: {
    label: 'به‌طور کلی، سازمان شما امروز چطور استفاده از هوش مصنوعی را مدیریت می‌کند؟',
    options: [
      { value: 'none', label: 'هنوز رویکرد یا فرایند مشخصی نداریم' },
      { value: 'individual', label: 'هر فرد/تیم به‌صورت مستقل، با ابزارها یا حساب‌های شخصی، مدیریت می‌کند' },
      { value: 'direct-api', label: 'اتصال مستقیم و متمرکز به API سرویس‌دهندگان داریم' },
      { value: 'internal-platform', label: 'پلتفرم یا ابزار داخلی اختصاصی ساخته‌ایم یا در حال ساخت آن هستیم' },
      { value: 'commercial-platform', label: 'از یک پلتفرم سازمانی/تجاری آماده استفاده می‌کنیم' },
      { value: 'manual-policy', label: 'صرفاً از طریق سیاست‌گذاری و فرایندهای دستی کنترل می‌کنیم (بدون ابزار فنی مشخص)' },
      { value: otherValue, label: 'سایر' },
    ],
  },
};

export function isValidMobile(value: string) {
  const digits = value.replace(/[۰-۹]/g, (d) => String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d))).replace(/[\s\-()+]/g, '');
  return /^\d{10,13}$/.test(digits);
}

function text(value: unknown) {
  return typeof value === 'string' ? value.trim().replace(/\s+/g, ' ').slice(0, textMax) : '';
}

function option(value: unknown, options: SurveyOption[]) {
  const normalized = text(value);
  return options.some((item) => item.value === normalized) ? normalized : '';
}

/** Keeps only known option values and bounded free text, so untrusted input never reaches the email as-is. */
export function sanitizeAnswers(raw: unknown): BuildGuildAnswers {
  const input = raw && typeof raw === 'object' && !Array.isArray(raw) ? (raw as Record<string, unknown>) : {};
  const role = option(input.role, surveyQuestions.role.options);
  const aiManagement = option(input.aiManagement, surveyQuestions.aiManagement.options);
  const challenges = Array.isArray(input.challenges)
    ? [...new Set(input.challenges.map((value) => option(value, surveyQuestions.challenges.options)).filter(Boolean))].slice(0, maxChallenges)
    : [];
  const mobile = text(input.mobile);

  return {
    fullName: text(input.fullName),
    mobile: isValidMobile(mobile) ? mobile : '',
    role,
    roleOther: role === otherValue ? text(input.roleOther) : '',
    orgSize: option(input.orgSize, surveyQuestions.orgSize.options),
    challenges,
    challengesOther: challenges.includes(otherValue) ? text(input.challengesOther) : '',
    aiManagement,
    aiManagementOther: aiManagement === otherValue ? text(input.aiManagementOther) : '',
  };
}

function labelOf(value: string, options: SurveyOption[]) {
  return options.find((item) => item.value === value)?.label ?? '';
}

/** Human-readable rows shared by the email body and any summary UI. */
export function describeAnswers(answers: BuildGuildAnswers) {
  const withOther = (label: string, other: string) => (label && other ? `${label}: ${other}` : label);
  const q = surveyQuestions;
  return [
    { label: q.contact.fullName, value: answers.fullName },
    { label: q.contact.mobile, value: answers.mobile },
    { label: q.role.label, value: withOther(labelOf(answers.role, q.role.options), answers.roleOther) },
    { label: q.orgSize.label, value: labelOf(answers.orgSize, q.orgSize.options) },
    {
      label: q.challenges.label,
      value: answers.challenges
        .map((value) => withOther(labelOf(value, q.challenges.options), value === otherValue ? answers.challengesOther : ''))
        .join('، '),
    },
    { label: q.aiManagement.label, value: withOther(labelOf(answers.aiManagement, q.aiManagement.options), answers.aiManagementOther) },
  ];
}

export function readStoredSurveyRaw(): string | null {
  try {
    return window.sessionStorage.getItem(surveyStorageKey);
  } catch {
    return null;
  }
}

export function parseStoredSurvey(raw: string | null): BuildGuildAnswers | null {
  try {
    return raw ? sanitizeAnswers(JSON.parse(raw)) : null;
  } catch {
    return null;
  }
}

export function writeStoredSurvey(answers: BuildGuildAnswers | null) {
  try {
    if (answers) window.sessionStorage.setItem(surveyStorageKey, JSON.stringify(answers));
    else window.sessionStorage.removeItem(surveyStorageKey);
  } catch {
    // Storage can be unavailable (private mode); the survey is still emailed.
  }
}
