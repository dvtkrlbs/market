import { afterEach, beforeEach, describe, expect, it, vi } from "vite-plus/test";

import {
  createEntry,
  getHealth,
  getOrdersPlaceholder,
  listEntries,
  login,
  logout,
  refresh,
  register,
  whoami,
} from "./api";

const baseUrl = "http://localhost:8080";

beforeEach(() => {
  vi.stubEnv("VITE_API_BASE_URL", baseUrl);
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

function mockFetch(response: Response) {
  const fetch = vi.fn<typeof globalThis.fetch>().mockResolvedValue(response);
  vi.stubGlobal("fetch", fetch);
  return fetch;
}

describe("generated Market client", () => {
  it("returns health data, status, and response headers", async () => {
    const fetch = mockFetch(Response.json({ status: "ok" }));

    const result = await getHealth();

    expect(fetch).toHaveBeenCalledExactlyOnceWith(`${baseUrl}/health`, {
      method: "GET",
      credentials: "include",
    });
    expect(result.status).toBe(200);
    expect(result.data).toEqual({ status: "ok" });
    expect(result.headers.get("content-type")).toContain("application/json");
  });

  it.each([
    { "x-request-id": "test-request" },
    new Headers({ "X-Request-Id": "test-request" }),
    [["x-request-id", "test-request"]] as [string, string][],
  ])("serializes login credentials and preserves headers: %j", async (headers) => {
    const session = { expiresAt: "2026-11-01T00:00:00Z" };
    const fetch = mockFetch(Response.json(session));
    const credentials = { email: "person@example.com", password: "correct horse battery staple" };

    const result = await login(credentials, { headers });

    expect(fetch).toHaveBeenCalledExactlyOnceWith(`${baseUrl}/auth/login`, {
      credentials: "include",
      method: "POST",
      headers: { "Content-Type": "application/json", "x-request-id": "test-request" },
      body: JSON.stringify(credentials),
    });
    expect(result.data).toEqual(session);
  });

  it("returns documented authentication failures without throwing", async () => {
    mockFetch(Response.json({ message: "Invalid credentials" }, { status: 401 }));

    const result = await login({ email: "person@example.com", password: "wrong password" });

    expect(result.status).toBe(401);
    if (result.status === 401) {
      expect(result.data.message).toBe("Invalid credentials");
    }
  });

  it.each([
    [listEntries, "/entries", "GET"],
    [whoami, "/auth/whoami", "GET"],
    [refresh, "/auth/refresh", "POST"],
    [getOrdersPlaceholder, "/orders", "GET"],
  ] as const)("passes authentication and cancellation to %s", async (operation, url, method) => {
    const fetch = mockFetch(Response.json({}));
    const options = {
      credentials: "include" as const,
      signal: new AbortController().signal,
    };

    await operation(options);

    expect(fetch).toHaveBeenCalledExactlyOnceWith(`${baseUrl}${url}`, { ...options, method });
  });

  it("handles logout's empty 204 response", async () => {
    const fetch = mockFetch(new Response(null, { status: 204 }));
    const options = { credentials: "include" as const };

    const result = await logout(options);

    expect(fetch).toHaveBeenCalledExactlyOnceWith(`${baseUrl}/auth/logout`, {
      ...options,
      method: "POST",
    });
    expect(result.status).toBe(204);
    expect(result.data).toBeUndefined();
  });

  it("serializes an entry and returns its created model", async () => {
    const entry = {
      description: "Coffee",
      amountCents: -450,
      occurredAt: "2026-10-10T09:00:00Z",
    };
    const created = { ...entry, id: 1, createdAt: "2026-10-10T09:00:00Z" };
    const fetch = mockFetch(Response.json(created, { status: 201 }));

    const result = await createEntry(entry);

    expect(fetch).toHaveBeenCalledExactlyOnceWith(`${baseUrl}/entries`, {
      credentials: "include",
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(entry),
    });
    expect(result.status).toBe(201);
    expect(result.data).toEqual(created);
  });

  it("registers a user", async () => {
    const credentials = { email: "person@example.com", password: "correct horse battery staple" };
    const user = { id: 1, email: credentials.email, createdAt: "2026-10-10T09:00:00Z" };
    const fetch = mockFetch(Response.json(user, { status: 201 }));

    const result = await register(credentials);

    expect(fetch).toHaveBeenCalledExactlyOnceWith(`${baseUrl}/auth/register`, {
      credentials: "include",
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(credentials),
    });
    expect(result.status).toBe(201);
    expect(result.data).toEqual(user);
  });

  it("propagates network failures", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof globalThis.fetch>().mockRejectedValue(new TypeError("offline")),
    );

    await expect(getHealth()).rejects.toThrow("offline");
  });
});
