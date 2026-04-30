import crypto from "node:crypto";

import { NextResponse } from "next/server";

import { listSignupSubmissions } from "@/app/lib/signup-db";

function safeEqual(a: string, b: string) {
  const ab = Buffer.from(a);
  const bb = Buffer.from(b);
  if (ab.length !== bb.length) return false;
  return crypto.timingSafeEqual(ab, bb);
}

function escapeHtml(value: string) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

export async function GET(req: Request) {
  const authHeader = req.headers.get("authorization") ?? "";
  const expected = process.env.SIGNUP_ADMIN_PASSWORD ?? "";

  if (!expected) {
    return new NextResponse("SIGNUP_ADMIN_PASSWORD is not set", { status: 500 });
  }

  const bearerPrefix = "Bearer ";
  const token = authHeader.startsWith(bearerPrefix)
    ? authHeader.slice(bearerPrefix.length).trim()
    : "";

  if (!token || !safeEqual(token, expected)) {
    return new NextResponse("Unauthorized", { status: 401 });
  }

  const rows = await listSignupSubmissions();
  const bodyRows = rows
    .map((row) => {
      const features = row.features.map(escapeHtml).join(", ");
      return `<tr>
  <td>${row.id}</td>
  <td>${escapeHtml(row.mobile)}</td>
  <td>${escapeHtml(row.team_size)}</td>
  <td>${escapeHtml(row.usage_type)}</td>
  <td>${escapeHtml(row.activity_field)}</td>
  <td>${features}</td>
  <td>${escapeHtml(row.created_at)}</td>
</tr>`;
    })
    .join("\n");

  const html = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Signup Responses</title>
  <style>
    body { font-family: Arial, sans-serif; padding: 24px; }
    table { border-collapse: collapse; width: 100%; }
    th, td { border: 1px solid #ddd; padding: 8px; vertical-align: top; }
    th { background: #f5f5f5; text-align: left; }
    tr:nth-child(even) { background: #fafafa; }
  </style>
</head>
<body>
  <h1>Signup Responses</h1>
  <p>Total: ${rows.length}</p>
  <table>
    <thead>
      <tr>
        <th>ID</th>
        <th>Mobile</th>
        <th>Team Size</th>
        <th>Usage Type</th>
        <th>Activity Field</th>
        <th>Features</th>
        <th>Created At (UTC)</th>
      </tr>
    </thead>
    <tbody>
      ${bodyRows}
    </tbody>
  </table>
</body>
</html>`;

  return new NextResponse(html, {
    status: 200,
    headers: {
      "cache-control": "no-store, max-age=0",
      "content-type": "text/html; charset=utf-8",
    },
  });
}
