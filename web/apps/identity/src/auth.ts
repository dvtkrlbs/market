import { queryOptions } from "@tanstack/react-query";
import {
  ApiConfigurationError,
  login,
  logout,
  register,
  whoami,
  type Credentials,
} from "shared/api";

export class AuthError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = "AuthError";
  }
}

export const sessionQueryOptions = queryOptions({
  queryKey: ["auth", "session"],
  queryFn: async ({ signal }) => {
    const response = await whoami({ signal });
    if (response.status !== 200) {
      throw new AuthError(response.data.message, response.status);
    }
    return response.data;
  },
  staleTime: 30_000,
  retry: false,
});

export async function signIn(credentials: Credentials) {
  const response = await login(credentials);
  if (response.status !== 200) {
    throw new AuthError(response.data.message, response.status);
  }
  return response.data;
}

export async function signOut() {
  const response = await logout();
  // An expired session is already signed out.
  if (response.status !== 204 && response.status !== 401) {
    throw new Error("Unable to sign out. Please try again.");
  }
}

export async function createAccount(credentials: Credentials) {
  const response = await register(credentials);
  if (response.status !== 201) {
    throw new AuthError(response.data.message, response.status);
  }
  return response.data;
}

export function authErrorMessage(error: Error) {
  return error instanceof AuthError || error instanceof ApiConfigurationError
    ? error.message
    : "Unable to connect to the server. Please try again.";
}
