import { NextResponse } from "next/server";

import { sendContactEmail } from "@/app/lib/email";

type ContactBody = {
  name?: unknown;
  email?: unknown;
  subject?: unknown;
  message?: unknown;
  website?: unknown;
};

const limits = {
  name: 120,
  email: 254,
  subject: 200,
  message: 5000,
} as const;

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as ContactBody | null;
  if (!body) {
    return NextResponse.json({ ok: false, error: "درخواست نامعتبر است." }, { status: 400 });
  }

  // A hidden field catches basic form bots without revealing whether an email was sent.
  if (typeof body.website === "string" && body.website.trim()) {
    return NextResponse.json({ ok: true });
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
    return NextResponse.json(
      { ok: false, error: "لطفاً همه فیلدها را با اطلاعات معتبر تکمیل کنید." },
      { status: 400 },
    );
  }

  try {
    await sendContactEmail({ name, email, subject, message });
    return NextResponse.json({ ok: true });
  } catch (error) {
    console.error("[website][contact] failed to send contact email", error);
    return NextResponse.json(
      { ok: false, error: "ارسال پیام انجام نشد. لطفاً دوباره تلاش کنید." },
      { status: 502 },
    );
  }
}

function normalizeField(value: unknown) {
  return typeof value === "string" ? value.trim() : "";
}

function isValidEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(value);
}
