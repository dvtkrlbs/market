// @vitest-environment jsdom

import { cleanup, render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vite-plus/test";
import { createCustomerRouter } from "./router";

const clients: QueryClient[] = [];
const session = {
  user: { id: 7, email: "owner@example.com", createdAt: "2026-10-10T09:00:00Z" },
  sessionExpiresAt: "2026-11-09T09:00:00Z",
};

beforeEach(() => {
  vi.stubGlobal("scrollTo", vi.fn());
  vi.stubEnv("VITE_API_BASE_URL", "http://localhost:8080");
});

afterEach(() => {
  cleanup();
  clients.splice(0).forEach((client) => client.clear());
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

async function renderCustomer(path = "/stores") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  clients.push(client);
  const router = createCustomerRouter(client, createMemoryHistory({ initialEntries: [path] }));
  await router.load();
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return router;
}

describe("customer stores", () => {
  it("loads /stores using the approved store listing mode", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json(session))
      .mockResolvedValueOnce(Response.json([{ id: 1, name: "Apple Market", approved: true }]));
    vi.stubGlobal("fetch", fetch);
    await renderCustomer();
    await screen.findByRole("heading", { name: "Apple Market" });
    expect(screen.queryByText("Zebra Market")).toBeNull();
    expect(fetch).toHaveBeenNthCalledWith(2, "http://localhost:8080/users/7/stores?for=shop", {
      method: "GET",
      credentials: "include",
      signal: expect.any(AbortSignal),
    });
  });

  it("shows an empty state", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(Response.json(session))
        .mockResolvedValueOnce(Response.json([])),
    );
    await renderCustomer();
    expect(await screen.findByText("No stores available yet.")).toBeDefined();
  });

  it("does not request stores if the user is signed out", async () => {
    const fetch = vi
      .fn()
      .mockResolvedValue(Response.json({ message: "authentication required" }, { status: 401 }));
    vi.stubGlobal("fetch", fetch);
    await renderCustomer();
    expect((await screen.findByRole("alert")).textContent).toContain("Please sign in");
    expect(fetch).toHaveBeenCalledTimes(1);
  });
});
