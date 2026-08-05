import { describe, expect, it } from "vitest";
import { inquirySchema } from "./validation";
describe("inquirySchema", () => {
  it("accepts a complete Serbian inquiry", () => {
    expect(
      inquirySchema.safeParse({
        fullName: "Petar Petrović",
        email: "petar@example.com",
        phone: "+381 64 123 456",
        message: "Molim vas za informacije.",
      }).success,
    ).toBe(true);
  });
  it("rejects malformed contact data and short messages", () => {
    const r = inquirySchema.safeParse({
      fullName: "P",
      email: "bad",
      phone: "x",
      message: "kratko",
    });
    expect(r.success).toBe(false);
    if (!r.success) expect(r.error.issues.length).toBeGreaterThanOrEqual(4);
  });
});
