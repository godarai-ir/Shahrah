export function decodeLoginIdentifierParam(value: string | undefined) {
  if (!value) {
    return "";
  }
  try {
    const base64 = value.replace(/-/g, "+").replace(/_/g, "/");
    const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, "=");
    const decoded = Buffer.from(padded, "base64").toString("utf8").trim();
    if (decoded.length > 254 || /[\r\n<>]/.test(decoded)) {
      return "";
    }
    if (/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(decoded.toLowerCase())) {
      return decoded.toLowerCase();
    }
    const mobile = decoded.replace(/[\s-]/g, "").replace(/^00/, "+");
    if (/^\+?[0-9]{10,15}$/.test(mobile)) {
      return mobile;
    }
  } catch {
    return "";
  }
  return "";
}
