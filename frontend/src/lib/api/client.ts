import type { ApiError } from "@/types";

// Empty = same-origin; Next.js rewrites /api/* to backend
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
    (headers as Record<string, string>)["Authorization"] = token;
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });

  if (!res.ok) {
    let err: ApiError = { code: "UNKNOWN", message: res.statusText };
    try {
      err = await res.json();
    } catch {
      /* empty body */
    }
    throw new ApiClientError(err.message || res.statusText, err.code, res.status);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
