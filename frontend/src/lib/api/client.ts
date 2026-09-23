import type { ApiError } from "@/types";

// Empty = same-origin; Next.js rewrites /api/* to api-gateway
const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";
const FETCH_TIMEOUT_MS = 12_000;

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
  options: RequestInit & { token?: string; timeoutMs?: number } = {}
): Promise<T> {
  const { token, timeoutMs = FETCH_TIMEOUT_MS, signal: parentSignal, ...init } = options;
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(init.headers ?? {}),
  };
  if (token) {
    (headers as Record<string, string>)["Authorization"] = authHeader(token);
  }

  const controller = new AbortController();
  const onParentAbort = () => controller.abort();
  parentSignal?.addEventListener("abort", onParentAbort);
  const timeout = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const res = await fetch(`${API_BASE}${path}`, {
      ...init,
      headers,
      signal: controller.signal,
    });

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
                "API not found. Start the backend: docker compose -f docker/docker-compose.yml up -d",
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
  } catch (e) {
    if (e instanceof DOMException && e.name === "AbortError") {
      if (parentSignal?.aborted) {
        throw new ApiClientError("Request cancelled", "ABORTED", 0);
      }
      throw new ApiClientError(
        "Request timed out — is api-gateway running on :8080?",
        "TIMEOUT",
        408
      );
    }
    throw e;
  } finally {
    clearTimeout(timeout);
    parentSignal?.removeEventListener("abort", onParentAbort);
  }
}
