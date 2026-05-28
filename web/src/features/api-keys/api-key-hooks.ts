import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createPersonalAPIKey,
  listPersonalAPIKeys,
  revokePersonalAPIKey,
} from "./api-key-api";
import type { CreatePersonalAPIKeyInput } from "./api-key-types";

export const apiKeyKeys = {
  personal: () => ["api-keys", "personal"] as const,
};

export function usePersonalAPIKeys() {
  return useQuery({
    queryKey: apiKeyKeys.personal(),
    queryFn: listPersonalAPIKeys,
    retry: false,
  });
}

export function useCreatePersonalAPIKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreatePersonalAPIKeyInput) => createPersonalAPIKey(input),
    onSuccess: () =>
      void queryClient.invalidateQueries({ queryKey: apiKeyKeys.personal() }),
  });
}

export function useRevokePersonalAPIKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (apiKeyId: string) => revokePersonalAPIKey(apiKeyId),
    onSuccess: () =>
      void queryClient.invalidateQueries({ queryKey: apiKeyKeys.personal() }),
  });
}
