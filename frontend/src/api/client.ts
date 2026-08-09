import { auth } from "../config/firebase";
import type { ApiError } from "../types";
const base = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api";
export class RequestError extends Error {
  detail: ApiError;
  status: number;
  constructor(detail: ApiError, status: number) {
    super(detail.message);
    this.detail = detail;
    this.status = status;
  }
}
export async function api<T>(
  path: string,
  options: RequestInit = {},
  admin = false,
): Promise<T> {
  const headers = new Headers(options.headers);
  if (!(options.body instanceof FormData))
    headers.set("Content-Type", "application/json");
  if (admin) {
    const token = await auth?.currentUser?.getIdToken();
    if (token) headers.set("Authorization", `Bearer ${token}`);
    else if (import.meta.env.DEV) headers.set("X-Dev-Admin-UID", "local-admin");
  }
  const res = await fetch(`${base}${path}`, { ...options, headers });
  const body = await res.json().catch(() => ({
    data: null,
    error: {
      code: "invalid_response",
      message: "Server je vratio neispravan odgovor.",
    },
  }));
  if (!res.ok)
    throw new RequestError(
      body.error || { code: "request_failed", message: "Zahtev nije uspeo." },
      res.status,
    );
  return body.data as T;
}
