export const PHONE_COUNTRIES = [
  { iso: "IR", label: "Iran", code: "+98", lengths: [10] },
  { iso: "AE", label: "UAE", code: "+971", lengths: [9] },
  { iso: "IQ", label: "Iraq", code: "+964", lengths: [10] },
  { iso: "TR", label: "Turkey", code: "+90", lengths: [10] },
  { iso: "QA", label: "Qatar", code: "+974", lengths: [8] },
  { iso: "OM", label: "Oman", code: "+968", lengths: [8] },
  { iso: "SA", label: "Saudi Arabia", code: "+966", lengths: [9] },
  { iso: "KW", label: "Kuwait", code: "+965", lengths: [8] },
  { iso: "IN", label: "India", code: "+91", lengths: [10] },
  { iso: "CN", label: "China", code: "+86", lengths: [11] },
  { iso: "RU", label: "Russia", code: "+7", lengths: [10] },
  { iso: "GB", label: "United Kingdom", code: "+44", lengths: [10] },
  { iso: "DE", label: "Germany", code: "+49", lengths: [10, 11] },
  { iso: "US", label: "United States", code: "+1", lengths: [10] }
];

export const getPhoneCountry = (iso) => PHONE_COUNTRIES.find((country) => country.iso === iso) || PHONE_COUNTRIES[0];

export const normalizeLocalPhoneNumber = (value) => String(value || "").replace(/\D/g, "").replace(/^0+/, "");

export const isValidLocalPhoneNumber = (value, country) => {
  const normalized = normalizeLocalPhoneNumber(value);
  return Boolean(normalized && country?.lengths?.includes(normalized.length));
};
