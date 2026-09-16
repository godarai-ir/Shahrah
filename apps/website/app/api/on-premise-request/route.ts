import { sendOnPremiseRequestEmail } from "@/app/lib/email";
import { deliverFormEmail, formError, formOk, isBotSubmission, normalizeField, readFormBody } from "@/app/lib/form-submission";
import {
  allQuestions,
  contactFields,
  detailKey,
  detailMax,
  isDetailVisible,
  isQuestionVisible,
  isValidEmail,
  isValidPhone,
  notesMax,
  type Answers,
  type ContactFieldId,
} from "@/app/lib/on-premise-questionnaire";

export async function POST(request: Request) {
  const body = await readFormBody(request);
  if (!body) {
    return formError("درخواست نامعتبر است.");
  }

  if (isBotSubmission(body)) {
    return formOk();
  }

  const rawContact = asRecord(body.contact);
  const contact = {} as Record<ContactFieldId, string>;
  for (const [id, field] of Object.entries(contactFields) as Array<[ContactFieldId, (typeof contactFields)[ContactFieldId]]>) {
    // Contact values end up in the email subject, so collapse any line breaks.
    const value = normalizeField(rawContact[id]).replace(/\s+/g, " ");
    if ((field.required && !value) || value.length > field.max) {
      return formError("لطفاً اطلاعات تماس را کامل و معتبر وارد کنید.");
    }
    contact[id] = value;
  }
  contact.email = contact.email.toLowerCase();
  if (!isValidEmail(contact.email) || !isValidPhone(contact.phone)) {
    return formError("ایمیل یا شماره تماس معتبر نیست.");
  }

  // Keep only known questions with known option values, so the email never carries arbitrary keys.
  const rawAnswers = asRecord(body.answers);
  const answers: Answers = {};
  for (const question of allQuestions) {
    const value = normalizeField(rawAnswers[question.id]);
    if (value && question.options.some((option) => option.value === value)) {
      answers[question.id] = value;
    }
  }
  for (const question of allQuestions) {
    if (!isQuestionVisible(question, answers)) {
      delete answers[question.id];
      continue;
    }
    const detail = normalizeField(rawAnswers[detailKey(question.id)]).slice(0, detailMax);
    if (detail && isDetailVisible(question, answers)) {
      answers[detailKey(question.id)] = detail;
    }
  }

  const notes = normalizeField(body.notes).slice(0, notesMax);

  return deliverFormEmail(
    "on-premise-request",
    () => sendOnPremiseRequestEmail({ contact, answers, notes }),
    "ارسال درخواست انجام نشد. لطفاً دوباره تلاش کنید.",
  );
}

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? (value as Record<string, unknown>) : {};
}
