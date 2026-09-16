import { NextResponse } from "next/server";

import { sanitizeAnswers } from "@/app/lib/build-guild-survey";
import { sendBuildGuildSurveyEmail } from "@/app/lib/email";

type RequestBody = {
  answers?: unknown;
  website?: unknown;
};

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as RequestBody | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json({ ok: false, error: "درخواست نامعتبر است." }, { status: 400 });
  }

  // A hidden field catches basic form bots without revealing whether an email was sent.
  if (typeof body.website === "string" && body.website.trim()) {
    return NextResponse.json({ ok: true });
  }

  try {
    await sendBuildGuildSurveyEmail({ answers: sanitizeAnswers(body.answers) });
    return NextResponse.json({ ok: true });
  } catch (error) {
    console.error("[website][build-guild-survey] failed to send email", error);
    return NextResponse.json({ ok: false, error: "ارسال پاسخ‌ها انجام نشد." }, { status: 502 });
  }
}
