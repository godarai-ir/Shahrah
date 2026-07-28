import { NextRequest, NextResponse } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { DELETE as deleteVideo, GET as getVideo } from "./[id]/route";
import { POST as cancelVideo } from "./[id]/cancel/route";
import { GET as getVideoContent } from "./[id]/content/route";
import { GET as listVideos, POST as createVideo } from "./route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());
const getCurrentMePermissionsMock = vi.hoisted(() => vi.fn());
const getNormalizedPermissionSetMock = vi.hoisted(() => vi.fn());
const fetchControlPlaneJSONMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneAssetMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));
vi.mock("@/lib/auth/me-permissions", () => ({
  getCurrentMePermissions: getCurrentMePermissionsMock,
  getNormalizedPermissionSet: getNormalizedPermissionSetMock,
}));
vi.mock("@/lib/control-plane", () => ({
  fetchControlPlaneJSON: fetchControlPlaneJSONMock,
  proxyToControlPlane: proxyToControlPlaneMock,
  proxyToControlPlaneAsset: proxyToControlPlaneAssetMock,
}));

const videoContext = { params: Promise.resolve({ id: "job-1" }) };

describe("playground video route authorization", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    readAuthFromRequestMock.mockResolvedValue({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    getCurrentMePermissionsMock.mockResolvedValue({ permissions: ["playground:run"] });
    getNormalizedPermissionSetMock.mockReturnValue(new Set(["playground:run"]));
    fetchControlPlaneJSONMock.mockResolvedValue({
      ok: true,
      status: 200,
      payload: {
        data: [{ id: "video-model", supported_modes: ["video_generation"] }],
      },
      responseHeaders: {},
    });
    proxyToControlPlaneMock.mockResolvedValue(NextResponse.json({ ok: true }));
    proxyToControlPlaneAssetMock.mockResolvedValue(new NextResponse("video"));
  });

  it("fails closed before proxying when no read or mutation permission is present", async () => {
    getNormalizedPermissionSetMock.mockReturnValue(new Set(["settings:list"]));

    const responses = await Promise.all([
      listVideos(requestFor("GET", "/api/playground/videos")),
      getVideo(requestFor("GET", "/api/playground/videos/job-1"), videoContext),
      getVideoContent(requestFor("GET", "/api/playground/videos/job-1/content"), videoContext),
      createVideo(videoCreateRequest()),
      cancelVideo(requestFor("POST", "/api/playground/videos/job-1/cancel"), videoContext),
      deleteVideo(requestFor("DELETE", "/api/playground/videos/job-1"), videoContext),
    ]);

    expect(responses.map((response) => response.status)).toEqual([403, 403, 403, 403, 403, 403]);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
    expect(proxyToControlPlaneAssetMock).not.toHaveBeenCalled();
  });

  it("allows read permissions to list, inspect, and fetch content but not mutate", async () => {
    getNormalizedPermissionSetMock.mockReturnValue(new Set(["playground:read"]));

    const listResponse = await listVideos(requestFor("GET", "/api/playground/videos"));
    const getResponse = await getVideo(requestFor("GET", "/api/playground/videos/job-1"), videoContext);
    const contentResponse = await getVideoContent(
      requestFor("GET", "/api/playground/videos/job-1/content"),
      videoContext,
    );
    const cancelResponse = await cancelVideo(
      requestFor("POST", "/api/playground/videos/job-1/cancel"),
      videoContext,
    );
    const deleteResponse = await deleteVideo(
      requestFor("DELETE", "/api/playground/videos/job-1"),
      videoContext,
    );

    expect([listResponse.status, getResponse.status, contentResponse.status]).toEqual([200, 200, 200]);
    expect([cancelResponse.status, deleteResponse.status]).toEqual([403, 403]);
    expect(proxyToControlPlaneMock).toHaveBeenCalledTimes(2);
    expect(proxyToControlPlaneAssetMock).toHaveBeenCalledOnce();
  });

  it("allows run permission to create, poll, cancel, and delete", async () => {
    const createResponse = await createVideo(videoCreateRequest());
    const getResponse = await getVideo(requestFor("GET", "/api/playground/videos/job-1"), videoContext);
    const cancelResponse = await cancelVideo(
      requestFor("POST", "/api/playground/videos/job-1/cancel"),
      videoContext,
    );
    const deleteResponse = await deleteVideo(
      requestFor("DELETE", "/api/playground/videos/job-1"),
      videoContext,
    );

    expect([
      createResponse.status,
      getResponse.status,
      cancelResponse.status,
      deleteResponse.status,
    ]).toEqual([200, 200, 200, 200]);
    expect(proxyToControlPlaneMock).toHaveBeenCalledTimes(4);
  });

  it("preserves policy permission fallbacks", async () => {
    getNormalizedPermissionSetMock.mockReturnValue(new Set(["policy:list"]));
    const readResponse = await listVideos(requestFor("GET", "/api/playground/videos"));
    const blockedMutation = await cancelVideo(
      requestFor("POST", "/api/playground/videos/job-1/cancel"),
      videoContext,
    );
    expect(readResponse.status).toBe(200);
    expect(blockedMutation.status).toBe(403);

    getNormalizedPermissionSetMock.mockReturnValue(new Set(["policy:manage"]));
    const allowedMutation = await cancelVideo(
      requestFor("POST", "/api/playground/videos/job-1/cancel"),
      videoContext,
    );
    expect(allowedMutation.status).toBe(200);
  });

  it("preserves authentication failures and does not proxy", async () => {
    readAuthFromRequestMock.mockResolvedValue({
      ok: false,
      status: 401,
      message: "Authentication required.",
    });

    const response = await listVideos(requestFor("GET", "/api/playground/videos"));

    expect(response.status).toBe(401);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });

  it("fails closed when permission lookup is unavailable", async () => {
    getCurrentMePermissionsMock.mockRejectedValue(new Error("unavailable"));

    const response = await listVideos(requestFor("GET", "/api/playground/videos"));

    expect(response.status).toBe(403);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });
});

function requestFor(method: string, path: string) {
  return new NextRequest(`http://acme.localhost${path}`, {
    method,
    headers: { host: "acme.localhost" },
  });
}

function videoCreateRequest() {
  const form = new FormData();
  form.set("model", "video-model");
  form.set("prompt", "A safe prompt");
  form.set("seconds", "5");
  form.set("resolution", "720p");
  form.set("aspect_ratio", "16:9");
  form.set("n", "1");
  form.set("service_tier", "standard");
  return new NextRequest("http://acme.localhost/api/playground/videos", {
    method: "POST",
    headers: { host: "acme.localhost" },
    body: form,
  });
}
