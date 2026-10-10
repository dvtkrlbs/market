import { queryOptions, useQuery } from "@tanstack/react-query";
import { ApiConfigurationError, createStore, listStores, whoami, type ListStoresFor } from "./api";

export class StoreRequestError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = "StoreRequestError";
  }
}

export function storeErrorMessage(error: Error) {
  if (error instanceof StoreRequestError && error.status === 401) {
    return "Please sign in to your Market account and try again.";
  }
  return error instanceof StoreRequestError || error instanceof ApiConfigurationError
    ? error.message
    : "Unable to connect to the server. Please try again.";
}

export const storeSessionOptions = queryOptions({
  queryKey: ["auth", "session"],
  queryFn: async ({ signal }) => {
    const response = await whoami({ signal });
    if (response.status !== 200)
      throw new StoreRequestError(response.data.message, response.status);
    return response.data;
  },
  staleTime: 30_000,
  retry: false,
});

export function useStores(mode: ListStoresFor) {
  const session = useQuery(storeSessionOptions);
  const userId = session.data?.user.id;
  const stores = useQuery({
    queryKey: ["stores", userId, mode],
    enabled: userId !== undefined,
    queryFn: async ({ signal }) => {
      if (userId === undefined) throw new Error("Missing user session");
      const response = await listStores(userId, { for: mode }, { signal });
      if (response.status !== 200)
        throw new StoreRequestError(response.data.message, response.status);
      return response.data;
    },
    retry: false,
  });
  return {
    data: stores.data,
    isPending: session.isPending || (session.isSuccess && stores.isPending),
    error: session.error ?? stores.error,
    retry: () => {
      void (session.isError ? session.refetch() : stores.refetch());
    },
  };
}

export async function applyForStore(userId: number, name: string) {
  const response = await createStore(userId, { name: name.trim() });
  if (response.status !== 201) throw new StoreRequestError(response.data.message, response.status);
  return response.data;
}
