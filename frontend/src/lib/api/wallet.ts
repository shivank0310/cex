import { apiFetch } from "./client";
import type { Balance, WalletAddress } from "@/types";

export interface WithdrawRequest {
  user_id: string;
  asset: string;
  amount: number;
  to_address: string;
}

export async function getBalance(userId: string, asset: string): Promise<Balance> {
  return apiFetch<Balance>(`/api/v1/wallet/balance/${userId}/${asset}`);
}

export async function createDepositAddress(
  userId: string,
  asset: string,
  chain = "ethereum"
): Promise<WalletAddress> {
  return apiFetch<WalletAddress>("/api/v1/wallet/address", {
    method: "POST",
    body: JSON.stringify({ user_id: userId, asset, chain }),
  });
}

export async function requestWithdrawal(req: WithdrawRequest): Promise<WithdrawalResponse> {
  return apiFetch<WithdrawalResponse>("/api/v1/wallet/withdraw", {
    method: "POST",
    body: JSON.stringify(req),
  });
}

// wallet withdraw response type
export interface WithdrawalResponse {
  id: string;
  user_id: string;
  asset: string;
  amount: number;
  to_address: string;
  status: string;
  tx_hash: string;
  created_at: string;
  updated_at: string;
}
