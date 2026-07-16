type DebugOTPEnvironment = Readonly<{
  NODE_ENV?: string;
  SHAHRAH_EXPOSE_DEBUG_OTP?: string;
}>;

export function shouldExposeDebugOTP(
  environment: DebugOTPEnvironment = process.env,
): boolean {
  const nodeEnv = environment.NODE_ENV?.trim().toLowerCase() ?? "";
  if (nodeEnv !== "development" && nodeEnv !== "test") {
    return false;
  }

  return environment.SHAHRAH_EXPOSE_DEBUG_OTP?.trim().toLowerCase() === "true";
}

export function debugOTPResponseField(
  debugOTP: string | undefined,
  environment: DebugOTPEnvironment = process.env,
): { debug_otp?: string } {
  if (!debugOTP || !shouldExposeDebugOTP(environment)) {
    return {};
  }

  return { debug_otp: debugOTP };
}
