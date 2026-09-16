import { sanitizeAnswers } from "@/app/lib/build-guild-survey";
import { sendBuildGuildSurveyEmail } from "@/app/lib/email";
import { deliverFormEmail, formError, formOk, isBotSubmission, readFormBody } from "@/app/lib/form-submission";

export async function POST(request: Request) {
  const body = await readFormBody(request);
  if (!body) {
    return formError("درخواست نامعتبر است.");
  }

  if (isBotSubmission(body)) {
    return formOk();
  }

  return deliverFormEmail(
    "build-guild-survey",
    () => sendBuildGuildSurveyEmail({ answers: sanitizeAnswers(body.answers) }),
    "ارسال پاسخ‌ها انجام نشد.",
  );
}
