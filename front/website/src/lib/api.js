const API_BASE = import.meta.env.VITE_API_BASE_URL || "";

export async function fetchJSON(path, options = {}) {
  const isFormData = typeof FormData !== "undefined" && options.body instanceof FormData;
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      ...(isFormData ? {} : { "Content-Type": "application/json" }),
      ...(options.headers || {})
    },
    credentials: "include"
  });

  const text = await response.text();
  const data = text ? JSON.parse(text) : {};

  if (!response.ok) {
    const body = data?.error;
    const message = (body && typeof body === "object" ? body.message : body) || data?.message || response.statusText || "ارتباط با سرور انجام نشد.";
    const err = new Error(message);
    err.status = response.status;
    err.code = body && typeof body === "object" ? body.code : data?.code;
    err.requestId = body && typeof body === "object" ? body.requestId : response.headers.get("X-Request-ID");
    err.body = data;
    throw err;
  }

  if (response.status === 204) {
    return {};
  }

  return data;
}
