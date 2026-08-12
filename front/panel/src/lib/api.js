export const API_BASE = import.meta.env.VITE_API_BASE_URL || "";

export async function fetchJSON(path, options = {}) {
  const isFormData = typeof FormData !== "undefined" && options.body instanceof FormData;
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      ...(!isFormData ? { "Content-Type": "application/json" } : {}),
      ...(options.headers || {})
    },
    credentials: "include"
  });

  if (!response.ok) {
	const payload = await response.json().catch(() => ({}));
	const body = payload.error;
	const error = new Error((body && typeof body === "object" ? body.message : body) || "ارتباط با سرور انجام نشد.");
	error.status = response.status;
	error.code = (body && typeof body === "object" ? body.code : payload.code);
	error.details = body && typeof body === "object" ? body.details : null;
	error.requestId = body && typeof body === "object" ? body.requestId : response.headers.get("X-Request-ID");
	throw error;
  }

  if (response.status === 204) {
    return {};
  }

  const text = await response.text();
  return text ? JSON.parse(text) : {};
}

export function idempotentHeaders(headers = {}) {
  return { "Idempotency-Key": crypto.randomUUID(), ...headers };
}
