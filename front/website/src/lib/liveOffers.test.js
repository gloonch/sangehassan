import { describe, expect, it } from "vitest";
import en from "@shared/i18n/en.json";
import fa from "@shared/i18n/fa.json";
import ar from "@shared/i18n/ar.json";
import { renderLiveOfferMessage } from "./liveOffers";

const dictionaries = { en, fa, ar };
const translator = (lang) => (key) => key.split(".").reduce((value, part) => value?.[part], dictionaries[lang]) ?? key;

const listing = {
  id: 42,
  title: "",
  productType: "block",
  stoneName: "تراورتن عباس‌آباد",
  quantity: 24,
  unit: "ton"
};

describe("live offer messages", () => {
  it("formats Persian numbers and listing fields", () => {
    expect(renderLiveOfferMessage(listing, "fa", translator("fa"))).toBe(
      "آگهی جدید: ۲۴ تن کوپ تراورتن عباس‌آباد"
    );
  });

  it("uses the localized English template", () => {
    expect(renderLiveOfferMessage({ ...listing, stoneName: "Abbasabad Travertine" }, "en", translator("en"))).toBe(
      "New listing: 24 tons of Abbasabad Travertine Block"
    );
  });

  it("formats finished listings in square metres", () => {
    const finished = {
      ...listing,
      productType: "finished",
      stoneName: "مرمریت مشکی",
      quantity: 35,
      unit: "square_meter"
    };
    expect(renderLiveOfferMessage(finished, "fa", translator("fa"))).toBe(
      "آگهی جدید: ۳۵ مترمربع فرآوری‌شده مرمریت مشکی"
    );
  });

  it("uses the listing title without guessing unknown codes", () => {
    expect(renderLiveOfferMessage({ title: "Special stone listing", productType: "unknown" }, "en", translator("en"))).toBe(
      "New listing: Special stone listing"
    );
  });

  it("formats Arabic content without exposing database codes", () => {
    expect(renderLiveOfferMessage({ ...listing, stoneName: "ترافرتين عباس آباد" }, "ar", translator("ar"))).toBe(
      "عرض جديد: ٢٤ طن من بلوك ترافرتين عباس آباد"
    );
  });
});
