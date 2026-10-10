// @vitest-environment jsdom

import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vite-plus/test";
import { createAdminRouter } from "./router";

const clients: QueryClient[] = [];
const session = {
  user: { id: 7, email: "owner@example.com", createdAt: "2026-10-10T09:00:00Z" },
  sessionExpiresAt: "2026-11-09T09:00:00Z",
};
const stores = [
  { id: 2, name: "Zebra Market", approved: false },
  { id: 1, name: "Apple Market", approved: true },
];

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

async function renderAdmin(path = "/stores") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  clients.push(client);
  const router = createAdminRouter(client, createMemoryHistory({ initialEntries: [path] }));
  await router.load();
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return router;
}

describe("admin stores", () => {
  it("navigates from the sidebar, fetches admin stores and sorts the table", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json(session))
      .mockResolvedValueOnce(Response.json(stores));
    vi.stubGlobal("fetch", fetch);
    const router = await renderAdmin("/");
    fireEvent.click(
      within(screen.getByRole("navigation", { name: "Admin navigation" })).getByRole("link", {
        name: "Stores",
      }),
    );
    await screen.findByRole("table");
    expect(router.state.location.pathname).toBe("/stores");
    expect(screen.getByText("Awaiting approval")).toBeDefined();
    expect(screen.getByText("Approved")).toBeDefined();
    expect(fetch).toHaveBeenNthCalledWith(2, "http://localhost:8080/users/7/stores?for=admin", {
      method: "GET",
      credentials: "include",
      signal: expect.any(AbortSignal),
    });
    fireEvent.click(screen.getByRole("button", { name: "Store name" }));
    await waitFor(() =>
      expect(screen.getAllByRole("row")[1].textContent).toContain("Apple Market"),
    );
    fireEvent.click(screen.getByRole("button", { name: /Store name/ }));
    await waitFor(() =>
      expect(screen.getAllByRole("row")[1].textContent).toContain("Zebra Market"),
    );
  });

  it("shows an empty table when there are no stores", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValueOnce(Response.json(session))
        .mockResolvedValueOnce(Response.json([])),
    );
    await renderAdmin();
    expect(await screen.findByText("No stores found.")).toBeDefined();
    expect(screen.getByRole("table")).toBeDefined();
  });

  it("shows request errors and retries the store list", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json(session))
      .mockResolvedValueOnce(Response.json({ message: "user mismatch" }, { status: 403 }))
      .mockResolvedValueOnce(Response.json(stores));
    vi.stubGlobal("fetch", fetch);
    await renderAdmin();
    expect((await screen.findByRole("alert")).textContent).toBe("user mismatch");
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    await screen.findByRole("table");
    expect(fetch).toHaveBeenCalledTimes(3);
  });

  it("approves a pending store through the admin endpoint and refreshes its status", async () => {
    const approved = { ...stores[0], approved: true };
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json(session))
      .mockResolvedValueOnce(Response.json(stores))
      .mockResolvedValueOnce(Response.json(approved))
      .mockResolvedValueOnce(Response.json([approved, stores[1]]));
    vi.stubGlobal("fetch", fetch);
    await renderAdmin();
    const button = await screen.findByRole("button", { name: "Approve Zebra Market" });
    expect(screen.queryByRole("button", { name: "Approve Apple Market" })).toBeNull();
    fireEvent.click(button);
    await waitFor(() => expect(screen.queryByText("Awaiting approval")).toBeNull());
    expect(fetch).toHaveBeenNthCalledWith(3, "http://localhost:8080/stores/2/approve", {
      method: "POST",
      credentials: "include",
    });
    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(4));
    expect(screen.queryByRole("button", { name: "Approve Zebra Market" })).toBeNull();
    expect(screen.getAllByText("Approved")).toHaveLength(2);
  });

  it("shows authorization errors without approving the row and allows retry", async () => {
    const approved = { ...stores[0], approved: true };
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json(session))
      .mockResolvedValueOnce(Response.json(stores))
      .mockResolvedValueOnce(Response.json({ message: "admin access required" }, { status: 403 }))
      .mockResolvedValueOnce(Response.json(approved))
      .mockResolvedValueOnce(Response.json([approved, stores[1]]));
    vi.stubGlobal("fetch", fetch);
    await renderAdmin();
    fireEvent.click(await screen.findByRole("button", { name: "Approve Zebra Market" }));
    expect((await screen.findByRole("alert")).textContent).toBe("admin access required");
    expect(screen.getByText("Awaiting approval")).toBeDefined();
    fireEvent.click(screen.getByRole("button", { name: "Approve Zebra Market" }));
    await waitFor(() => expect(screen.queryByText("Awaiting approval")).toBeNull());
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("disables the approve action during a request", async () => {
    let resolve!: (response: Response) => void;
    const approved = { ...stores[0], approved: true };
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json(session))
      .mockResolvedValueOnce(Response.json(stores))
      .mockReturnValueOnce(
        new Promise<Response>((done) => {
          resolve = done;
        }),
      )
      .mockResolvedValueOnce(Response.json([approved, stores[1]]));
    vi.stubGlobal("fetch", fetch);
    await renderAdmin();
    const button = await screen.findByRole<HTMLButtonElement>("button", {
      name: "Approve Zebra Market",
    });
    fireEvent.click(button);
    await waitFor(() => expect(button.disabled).toBe(true));
    fireEvent.click(button);
    expect(fetch).toHaveBeenCalledTimes(3);
    resolve(Response.json(approved));
    await waitFor(() => expect(screen.queryByText("Awaiting approval")).toBeNull());
  });
});
