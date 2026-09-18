import { apiFetch } from "./client";
import type { Order, PlaceOrderRequest, PlaceOrderResponse } from "@/types";

export async function placeOrder(
  req: PlaceOrderRequest,
  token: string
): Promise<PlaceOrderResponse> {
  return apiFetch<PlaceOrderResponse>("/api/v1/orders", {
    method: "POST",
    body: JSON.stringify(req),
    token,
  });
}

export async function getOrder(orderId: string, token: string): Promise<Order> {
  return apiFetch<Order>(`/api/v1/orders/${orderId}`, { token });
}

export async function cancelOrder(
  orderId: string,
  symbol: string,
  token: string
): Promise<Order> {
  const encoded = symbol.replace("/", "%2F");
  return apiFetch<Order>(`/api/v1/orders/${orderId}?symbol=${encoded}`, {
    method: "DELETE",
    token,
  });
}
