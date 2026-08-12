import { afterEach, describe, expect, it, vi } from "vitest";
import { fetchJSON } from "./api";

describe("API error compatibility", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("understands the standard error envelope", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ success: false, error: { code: "MODULE_DISABLED", message: "ماژول غیرفعال است", details: null, requestId: "req-1" } }), { status: 409, headers: { "Content-Type": "application/json" } })));
    await expect(fetchJSON("/api/v1/test")).rejects.toMatchObject({ code: "MODULE_DISABLED", message: "ماژول غیرفعال است", requestId: "req-1" });
  });

  it("still understands the legacy string error", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: "legacy error" }), { status: 400, headers: { "Content-Type": "application/json", "X-Request-ID": "req-2" } })));
    await expect(fetchJSON("/api/v1/test")).rejects.toMatchObject({ message: "legacy error", requestId: "req-2" });
  });
});
