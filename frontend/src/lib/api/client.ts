import type { ApiError } from "@/types";

// Empty = same-origin; Next.js rewrites /api/* to nginx → api-gateway
const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";

export class ApiClientError extends Error {
  code: string;
  status: number;

  constructor(message: string, code: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

export function authHeader(token: string): string {
  if (!token) return "";
  return token.startsWith("Bearer ") ? token : `Bearer ${token}`;
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit & { token?: string } = {}
): Promise<T> {
  const { token, ...init } = options;
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(init.headers ?? {}),
  };
  if (token) {
    (headers as Record<string, string>)["Authorization"] = authHeader(token);
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });

  if (!res.ok) {
    let err: ApiError = { code: "UNKNOWN", message: res.statusText };
    const contentType = res.headers.get("content-type") ?? "";
    try {
      if (contentType.includes("application/json")) {
        err = await res.json();
      } else {
        const text = await res.text();
        if (res.status === 404) {
          err = {
            code: "API_UNAVAILABLE",
            message:
              "API not found. Start the backend (docker compose up) and ensure NEXT_PUBLIC_API_URL points to api-gateway :8080 — not Apache on :80.",
          };
        } else if (text) {
          err = { code: "HTTP_ERROR", message: text.slice(0, 200) };
        }
      }
    } catch {
      /* empty body */
    }
    throw new ApiClientError(err.message || res.statusText, err.code, res.status);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
