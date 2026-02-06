const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

export async function fetchJSON(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {})
    },
    credentials: "include"
  });

  const text = await response.text();
  const data = text ? JSON.parse(text) : {};

  if (!response.ok) {
    const message = data?.error || data?.message || response.statusText || "Request failed";
    const err = new Error(message);
    err.status = response.status;
    err.body = data;
    throw err;
  }

  if (response.status === 204) {
    return {};
  }

  return data;
}
