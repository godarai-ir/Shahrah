// Browser-side counterpart of form-submission.ts: posts a website form and surfaces the API error message.
const defaultFailureMessage = 'ارسال انجام نشد. لطفاً دوباره تلاش کنید.';

export async function submitForm(url: string, body: unknown, failureMessage = defaultFailureMessage) {
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
  const payload = (await response.json().catch(() => null)) as { error?: string } | null;
  if (!response.ok) {
    throw new Error(payload?.error || failureMessage);
  }
}

export function isValidEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(value);
}
