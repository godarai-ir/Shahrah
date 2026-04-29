import { NextResponse } from "next/server";

import {
  insertSignupSubmission,
  type SignupSubmission,
} from "@/app/lib/signup-db";

function isValidSubmission(input: Partial<SignupSubmission>): input is SignupSubmission {
  return (
    typeof input.mobile === "string" &&
    input.mobile.trim().length >= 10 &&
    typeof input.teamSize === "string" &&
    input.teamSize.trim().length > 0 &&
    typeof input.usageType === "string" &&
    input.usageType.trim().length > 0 &&
    typeof input.activityField === "string" &&
    Array.isArray(input.features)
  );
}

export async function POST(req: Request) {
  try {
    const body = (await req.json()) as Partial<SignupSubmission>;
    if (!isValidSubmission(body)) {
      return NextResponse.json(
        { ok: false, error: "Invalid payload" },
        { status: 400 },
      );
    }

    const id = await insertSignupSubmission({
      mobile: body.mobile.trim(),
      teamSize: body.teamSize.trim(),
      usageType: body.usageType.trim(),
      activityField: body.activityField.trim(),
      features: body.features.map((v) => String(v)),
    });

    return NextResponse.json({ ok: true, id });
  } catch (error) {
    return NextResponse.json(
      {
        ok: false,
        error: "Failed to save submission",
        detail: error instanceof Error ? error.message : "unknown_error",
      },
      { status: 500 },
    );
  }
}
