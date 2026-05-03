import { NextRequest, NextResponse } from "next/server";

const CONTROL_PLANE_URL = process.env.CONTROL_PLANE_URL || "http://localhost:8080";

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const path = searchParams.get("path") || "/wallet/transactions";
  
  try {
    const response = await fetch(`${CONTROL_PLANE_URL}${path}`, {
      headers: {
        "Content-Type": "application/json",
        // Forward all relevant headers for authentication
        "Cookie": request.headers.get("Cookie") || "",
        "X-Tenant-ID": request.headers.get("X-Tenant-ID") || "",
      },
      credentials: "include",
    });

    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    console.error("Proxy error:", error);
    return NextResponse.json(
      { error: "Failed to connect to backend" },
      { status: 502 }
    );
  }
}

export async function POST(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const path = searchParams.get("path") || "/wallet";
  
  try {
    const body = await request.json();
    
    const response = await fetch(`${CONTROL_PLANE_URL}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Cookie": request.headers.get("Cookie") || "",
        "X-Tenant-ID": request.headers.get("X-Tenant-ID") || "",
      },
      credentials: "include",
      body: JSON.stringify(body),
    });

    const data = await response.json();
    return NextResponse.json(data, { status: response.status });
  } catch (error) {
    console.error("Proxy error:", error);
    return NextResponse.json(
      { error: "Failed to connect to backend" },
      { status: 502 }
    );
  }
}