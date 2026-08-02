import { describe, expect, it } from "vitest";
import fa from "@shared/i18n/fa.json";
import { formatListingQuantity, getListingQuantityFieldLabel, getListingQuantityUnit } from "./listings";

const t = (key) => key.split(".").reduce((value, part) => value?.[part], fa) ?? key;

describe("listing quantities", () => {
  it("keeps block quantities in tons", () => {
    expect(getListingQuantityUnit("block")).toBe("ton");
    expect(formatListingQuantity({ form: "block", tonnage: 24.5 }, "fa", t)).toBe("۲۴٫۵ تن");
    expect(getListingQuantityFieldLabel("block", t)).toBe("مقدار کوپ (تن)");
  });

  it("uses square metres for finished and other non-block listings", () => {
    expect(getListingQuantityUnit("finished")).toBe("square_meter");
    expect(getListingQuantityUnit("tile")).toBe("square_meter");
    expect(formatListingQuantity({ form: "finished", tonnage: 35 }, "fa", t)).toBe("۳۵ مترمربع");
    expect(getListingQuantityFieldLabel("finished", t)).toBe("متراژ (مترمربع)");
  });
});
