// @vitest-environment jsdom

import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vite-plus/test";
import { createManagementRouter } from "./router";

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

async function renderManagement(path = "/apply") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  clients.push(client);
  const router = createManagementRouter(client, createMemoryHistory({ initialEntries: [path] }));
  await router.load();
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return router;
}

describe("management applications", () => {
  it("opens /apply from navigation and submits the signed-in user's store", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json(session))
      .mockResolvedValueOnce(
        Response.json({ id: 3, name: "New Market", approved: false }, { status: 201 }),
      );
    vi.stubGlobal("fetch", fetch);
    const router = await renderManagement("/");
    fireEvent.click(screen.getByRole("link", { name: "Apply for a store" }));
    await screen.findByText("Applying as owner@example.com.");
    expect(router.state.location.pathname).toBe("/apply");
    fireEvent.change(screen.getByLabelText(/^Store name/), { target: { value: "  New Market  " } });
    fireEvent.click(screen.getByRole("button", { name: "Submit application" }));
    expect(
      (await screen.findByRole("status", { name: "Application submitted" })).textContent,
    ).toContain("Application submitted for New Market");
    expect(fetch).toHaveBeenNthCalledWith(2, "http://localhost:8080/users/7/stores", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: "New Market" }),
    });
    expect(screen.queryByRole("button", { name: "Submit application" })).toBeNull();
  });

  it("rejects a whitespace-only store name without submitting", async () => {
    const fetch = vi.fn().mockResolvedValue(Response.json(session));
    vi.stubGlobal("fetch", fetch);
    await renderManagement();
    await screen.findByText("Applying as owner@example.com.");
    fireEvent.change(screen.getByLabelText(/^Store name/), { target: { value: "   " } });
    fireEvent.click(screen.getByRole("button", { name: "Submit application" }));
    expect((await screen.findByRole("alert")).textContent).toBe("Enter a store name.");
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it("keeps entered details after a failed application and allows retry", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json(session))
      .mockResolvedValueOnce(Response.json({ message: "Please try again later" }, { status: 400 }))
      .mockResolvedValueOnce(
        Response.json({ id: 3, name: "New Market", approved: false }, { status: 201 }),
      );
    vi.stubGlobal("fetch", fetch);
    await renderManagement();
    await screen.findByText("Applying as owner@example.com.");
    fireEvent.change(screen.getByLabelText(/^Store name/), { target: { value: "New Market" } });
    fireEvent.click(screen.getByRole("button", { name: "Submit application" }));
    expect((await screen.findByRole("alert")).textContent).toBe("Please try again later");
    expect(screen.getByLabelText<HTMLInputElement>(/^Store name/).value).toBe("New Market");
    fireEvent.click(screen.getByRole("button", { name: "Submit application" }));
    expect(
      (await screen.findByRole("status", { name: "Application submitted" })).textContent,
    ).toContain("Application submitted");
  });

  it("requires a session before allowing an application", async () => {
    const fetch = vi
      .fn()
      .mockResolvedValue(Response.json({ message: "authentication required" }, { status: 401 }));
    vi.stubGlobal("fetch", fetch);
    await renderManagement();
    expect((await screen.findByRole("alert")).textContent).toContain("Please sign in");
    expect(screen.queryByRole("button", { name: "Submit application" })).toBeNull();
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  it("disables the application form while submitting and prevents duplicate requests", async () => {
    let resolve!: (response: Response) => void;
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json(session))
      .mockReturnValueOnce(
        new Promise<Response>((done) => {
          resolve = done;
        }),
      );
    vi.stubGlobal("fetch", fetch);
    await renderManagement();
    await screen.findByText("Applying as owner@example.com.");
    const input = screen.getByLabelText<HTMLInputElement>(/^Store name/);
    fireEvent.change(input, { target: { value: "New Market" } });
    fireEvent.click(screen.getByRole("button", { name: "Submit application" }));
    await waitFor(() => expect(input.disabled).toBe(true));
    fireEvent.submit(input.closest("form")!);
    expect(fetch).toHaveBeenCalledTimes(2);
    resolve(Response.json({ id: 3, name: "New Market", approved: false }, { status: 201 }));
    await screen.findByRole("status", { name: "Application submitted" });
  });
});
