import { afterEach, describe, expect, it, vi } from "vitest";
import { adminApi } from "./catalog";

afterEach(() => vi.unstubAllGlobals());

describe("admin image upload", () => {
  it("sends multiple images and metadata as browser-managed multipart data", async () => {
    const fetchMock = vi.fn(async (_url: string, init?: RequestInit) => {
      const body = init?.body;
      expect(body).toBeInstanceOf(FormData);
      const form = body as FormData;
      expect(form.getAll("images")).toHaveLength(2);
      expect(form.get("alt")).toBe("Model Elegance");
      expect(form.get("isPrimary")).toBe("true");
      expect(new Headers(init?.headers).has("Content-Type")).toBe(false);
      return new Response(JSON.stringify({ data: { id: "p1" }, error: null }), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      });
    });
    vi.stubGlobal("fetch", fetchMock);

    const file = new File(["image"], "model.webp", { type: "image/webp" });
    const second = new File(["image-2"], "detail.webp", {
      type: "image/webp",
    });
    await adminApi.upload("p1", [file, second], {
      alt: "Model Elegance",
      isPrimary: true,
    });

    expect(fetchMock).toHaveBeenCalledOnce();
  });
});
