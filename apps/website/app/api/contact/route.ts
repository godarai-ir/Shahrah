import { sendContactEmail } from "@/app/lib/email";
import { deliverFormEmail, formError, formOk, isBotSubmission, normalizeField, readFormBody } from "@/app/lib/form-submission";
import { isValidEmail } from "@/app/lib/submit-form";

const limits = {
  name: 120,
  email: 254,
  subject: 200,
  message: 5000,
} as const;

export async function POST(request: Request) {
  const body = await readFormBody(request);
  if (!body) {
    return formError("درخواست نامعتبر است.");
  }

  if (isBotSubmission(body)) {
    return formOk();
  }

  const name = normalizeField(body.name);
  const email = normalizeField(body.email).toLowerCase();
  const subject = normalizeField(body.subject);
  const message = normalizeField(body.message);

  if (
    !name ||
    !email ||
    !subject ||
    !message ||
    name.length > limits.name ||
    email.length > limits.email ||
    subject.length > limits.subject ||
    message.length > limits.message ||
    !isValidEmail(email)
  ) {
    return formError("لطفاً همه فیلدها را با اطلاعات معتبر تکمیل کنید.");
  }

  return deliverFormEmail(
    "contact",
    () => sendContactEmail({ name, email, subject, message }),
    "ارسال پیام انجام نشد. لطفاً دوباره تلاش کنید.",
  );
}
