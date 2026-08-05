# APIهای native میسترال در گدارAI

گدارAI مسیرهای رسمی زیر را با قرارداد native میسترال، احراز هویت گدارAI و مسیریابی ثابت از Silk Road ارائه می‌کند:

- `/v1/files` و عملیات resource، content و URL؛
- `/v1/batch/jobs` و عملیات get، cancel و delete؛
- `/v1/agents` و عملیات list/pages/get/update/delete؛
- `/v1/conversations` و عملیات append/history/messages/restart/delete.

این APIها reverse proxy عمومی نیستند. هر درخواست به یک `ProviderOperation` ثابت تبدیل می‌شود و کاربر نمی‌تواند path یا host upstream را تعیین کند.

## فعال‌سازی

هر سطح در حالت پیش‌فرض خاموش است:

```text
SHAHRAH_MISTRAL_FILES_ENABLED=true
SHAHRAH_MISTRAL_BATCH_ENABLED=true
SHAHRAH_MISTRAL_AGENTS_ENABLED=true
SHAHRAH_MISTRAL_CONVERSATIONS_ENABLED=true
SHAHRAH_MISTRAL_CONVERSATIONS_STREAMING_ENABLED=true
```

`SHAHRAH_MISTRAL_ENABLED=false` همهٔ مسیرهای میسترال را خاموش می‌کند. مقدار نامعتبر نیز fail-closed تفسیر می‌شود. فعال‌سازی production باید مدل‌به‌مدل و پس از contract test و smoke واقعی انجام شود.

چون `/v1/files` با API سازگار OpenAI مشترک است، برای create/list میسترال header زیر لازم است:

```http
x-godarai-provider: mistral-ai
```

در عملیات یک resource، پیشوند عمومی `mfile_gdr_` ارائه‌دهنده را بدون header مشخص می‌کند. اگر یک مدل روی چند حساب میسترال تنظیم شده باشد، `x-godarai-provider-account` الزامی است.

## مالکیت و IDها

IDهای upstream هرگز به client برگردانده نمی‌شوند. گدارAI برای منابع IDهای opaque زیر را صادر می‌کند:

- `mfile_gdr_...`
- `mbatch_gdr_...`
- `magent_gdr_...`
- `mconv_gdr_...`

نگاشت شامل tenant، حساب ارائه‌دهنده و نوع resource است. get/update/delete، ارجاع `file_id`، `input_files`، `agent_id` و `handoffs` در صورت tenant یا حساب متفاوت با خطا متوقف می‌شوند. list فقط منابع ساخته‌شده از طریق همان tenant و حساب را نشان می‌دهد.

## Files

آپلود با `multipart/form-data`، دقیقاً یک فیلد `file`، purpose برابر `batch` یا `ocr` و حداکثر اندازهٔ فایل ۵۱۲ MiB پذیرفته می‌شود. content به‌صورت stream از tunnel عبور می‌کند. ارجاع ID عمومی Files در OCR و Audio Transcriptions پیش از ارسال به upstream resolve می‌شود و affinity حساب حفظ می‌شود.

## Batch و صورتحساب

فقط endpointهای allowlist رسمی پذیرفته می‌شوند و eligibility مدل از runtime profile همان revision کنترل می‌شود. هنگام create هزینه به‌صورت محافظه‌کارانه reserve و نگاشت job با pricing revision ذخیره می‌شود.

worker پس‌زمینه jobهای درحال اجرا را با claim اتمیک poll می‌کند. در موفقیت، فایل JSONL خروجی خوانده می‌شود؛ عملیات توکنی با usage تجمیعی، OCR با تعداد صفحه و Transcription با ثانیه تسویه می‌شوند. failure، timeout و cancellation reservation را آزاد می‌کنند. commit صورتحساب و تغییر terminal state قابل retry هستند.

## Agents و Conversations

مدل و endpoint باید در runtime profile مجاز باشند. function tool سفارشی پذیرفته می‌شود؛ ابزار داخلی پولی تا زمانی که پاسخ رسمی usage مقدار قابل اعتماد همان ابزار را ندهد، fail-closed است. Conversationهای sync و stream هر دو reserve/finalize می‌شوند؛ stream تنها پس از usage پایانی commit می‌شود و قطع یا stream ناقص reservation را آزاد می‌کند.

## منابع رسمی

- [Mistral OpenAPI](https://docs.mistral.ai/openapi.yaml)
- [Files API](https://docs.mistral.ai/api/endpoint/files)
- [Batch API](https://docs.mistral.ai/api/endpoint/batch)
- [Agents API](https://docs.mistral.ai/api/endpoint/beta/agents)
- [Conversations API](https://docs.mistral.ai/api/endpoint/beta/conversations)
