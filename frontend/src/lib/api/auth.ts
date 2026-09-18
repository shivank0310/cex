import { apiFetch } from "./client";
import type { AuthResponse, AuthUser } from "@/types";

export async function register(
  email: string,
  password: string,
  username?: string
): Promise<AuthResponse> {
  return apiFetch<AuthResponse>("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password, username: username ?? "" }),
  });
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  return apiFetch<AuthResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function refresh(refreshToken: string): Promise<AuthResponse> {
  return apiFetch<AuthResponse>("/api/v1/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
}

export async function logout(refreshToken: string): Promise<void> {
  await apiFetch<{ status: string }>("/api/v1/auth/logout", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
}

export async function getMe(accessToken: string): Promise<AuthUser> {
  return apiFetch<AuthUser>("/api/v1/auth/me", { token: accessToken });
}
