import { apiRequest } from "@/shared/api/client";
import type {
  APIKey,
  CreatePersonalAPIKeyInput,
  CreatedAPIKey,
} from "./api-key-types";

export function listPersonalAPIKeys() {
  return apiRequest<{ api_keys: APIKey[] }>("/api/v1/api-keys", {
    skipTenant: true,
  });
}

export function createPersonalAPIKey(input: CreatePersonalAPIKeyInput) {
  return apiRequest<CreatedAPIKey>("/api/v1/api-keys", {
    method: "POST",
    skipTenant: true,
    body: input,
  });
}

export function revokePersonalAPIKey(apiKeyId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/api-keys/${apiKeyId}`, {
    method: "DELETE",
    skipTenant: true,
  });
}
