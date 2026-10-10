// @vitest-environment jsdom

import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vite-plus/test";
import { createIdentityRouter } from "./router";

const clients: QueryClient[] = [];

beforeEach(() => {
  vi.stubGlobal("scrollTo", vi.fn());
  vi.stubEnv("VITE_API_BASE_URL", "http://localhost:8080");
});

afterEach(() => {
  cleanup();
  for (const client of clients.splice(0)) client.clear();
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

async function renderIdentity(path = "/login") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  clients.push(client);
  const router = createIdentityRouter(client, createMemoryHistory({ initialEntries: [path] }));
  await router.load();
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return router;
}

function fillLogin() {
  fireEvent.change(screen.getByLabelText("Email", { exact: false }), {
    target: { value: "person@example.com" },
  });
  fireEvent.change(screen.getByLabelText("Password", { exact: false }), {
    target: { value: "correct horse battery staple" },
  });
}

const session = {
  user: { id: 1, email: "person@example.com", createdAt: "2026-10-10T09:00:00Z" },
  sessionExpiresAt: "2026-11-09T09:00:00Z",
};

describe("identity login", () => {
  it("routes the home page to an accessible login form", async () => {
    const router = await renderIdentity("/");

    await screen.findByRole("heading", { name: "Sign in" });
    expect(router.state.location.pathname).toBe("/login");
    const email = screen.getByLabelText<HTMLInputElement>("Email", { exact: false });
    const password = screen.getByLabelText<HTMLInputElement>("Password", { exact: false });
    expect(email.type).toBe("email");
    expect(email.autocomplete).toBe("username");
    expect(email.required).toBe(true);
    expect(password.type).toBe("password");
    expect(password.autocomplete).toBe("current-password");
    expect(password.required).toBe(true);
  });

  it("signs in with cookies, loads the session, and signs out", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(Response.json({ expiresAt: session.sessionExpiresAt }))
      .mockResolvedValueOnce(Response.json(session))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    vi.stubGlobal("fetch", fetch);
    const router = await renderIdentity();
    fillLogin();
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await screen.findByRole("heading", { name: "You’re signed in" });
    expect(router.state.location.pathname).toBe("/session");
    expect(screen.getByText("Signed in as person@example.com.")).toBeDefined();
    expect(fetch).toHaveBeenNthCalledWith(1, "http://localhost:8080/auth/login", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: "person@example.com",
        password: "correct horse battery staple",
      }),
    });
    expect(fetch).toHaveBeenNthCalledWith(2, "http://localhost:8080/auth/whoami", {
      method: "GET",
      credentials: "include",
      signal: expect.any(AbortSignal),
    });

    fireEvent.click(screen.getByRole("button", { name: "Sign out" }));
    await screen.findByRole("heading", { name: "Sign in" });
    expect(router.state.location.pathname).toBe("/login");
    expect(fetch).toHaveBeenNthCalledWith(3, "http://localhost:8080/auth/logout", {
      method: "POST",
      credentials: "include",
    });
  });

  it("shows invalid credentials and allows another attempt", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(
        Response.json({ message: "invalid email or password" }, { status: 401 }),
      )
      .mockResolvedValueOnce(Response.json({ expiresAt: session.sessionExpiresAt }))
      .mockResolvedValueOnce(Response.json(session));
    vi.stubGlobal("fetch", fetch);
    await renderIdentity();
    fillLogin();
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    expect((await screen.findByRole("alert")).textContent).toBe("invalid email or password");
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));
    await screen.findByRole("heading", { name: "You’re signed in" });
    expect(fetch).toHaveBeenCalledTimes(3);
  });

  it("disables the form while a request is pending", async () => {
    let resolve!: (response: Response) => void;
    const fetch = vi.fn<typeof globalThis.fetch>().mockReturnValue(
      new Promise<Response>((done) => {
        resolve = done;
      }),
    );
    vi.stubGlobal("fetch", fetch);
    await renderIdentity();
    fillLogin();
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(screen.getByLabelText<HTMLInputElement>("Email", { exact: false }).disabled).toBe(
        true,
      );
      expect(screen.getByLabelText<HTMLInputElement>("Password", { exact: false }).disabled).toBe(
        true,
      );
    });
    fireEvent.submit(screen.getByLabelText("Email", { exact: false }).closest("form")!);
    expect(fetch).toHaveBeenCalledTimes(1);
    resolve(Response.json({ message: "invalid email or password" }, { status: 401 }));
    await screen.findByRole("alert");
  });

  it("shows network failures without navigating away", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")));
    const router = await renderIdentity();
    fillLogin();
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    expect((await screen.findByRole("alert")).textContent).toContain("Unable to connect");
    expect(router.state.location.pathname).toBe("/login");
  });

  it("identifies a missing API URL instead of reporting a network failure", async () => {
    vi.stubEnv("VITE_API_BASE_URL", "");
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    await renderIdentity();
    fillLogin();
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    expect((await screen.findByRole("alert")).textContent).toContain(
      "Set VITE_API_BASE_URL in web/.env",
    );
    expect(fetch).not.toHaveBeenCalled();
  });

  it("redirects an unauthenticated session route to login", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(Response.json({ message: "authentication required" }, { status: 401 })),
    );
    const router = await renderIdentity("/session");

    await screen.findByRole("heading", { name: "Sign in" });
    expect(router.state.location.pathname).toBe("/login");
  });
});

function fillRegistration(confirmation = "correct horse battery staple") {
  fireEvent.change(screen.getByLabelText("Email", { exact: false }), {
    target: { value: "person@example.com" },
  });
  fireEvent.change(screen.getByLabelText(/^Password/), {
    target: { value: "correct horse battery staple" },
  });
  fireEvent.change(screen.getByLabelText(/^Confirm password/), {
    target: { value: confirmation },
  });
}

describe("identity registration", () => {
  it("navigates from login, creates an account, and returns to login", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValue(Response.json(session.user, { status: 201 }));
    vi.stubGlobal("fetch", fetch);
    const router = await renderIdentity();
    fireEvent.click(screen.getByRole("link", { name: "Create an account" }));
    await screen.findByRole("heading", { name: "Create an account" });
    expect(router.state.location.pathname).toBe("/register");
    const password = screen.getByLabelText<HTMLInputElement>(/^Password/);
    expect(password.autocomplete).toBe("new-password");
    expect(password.minLength).toBe(8);
    fillRegistration();
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));

    await screen.findByRole("heading", { name: "Sign in" });
    expect((await screen.findByRole("status", { name: "Registration complete" })).textContent).toBe(
      "Account created. You can now sign in.",
    );
    expect(router.state.location.pathname).toBe("/login");
    expect(fetch).toHaveBeenCalledExactlyOnceWith("http://localhost:8080/auth/register", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: "person@example.com",
        password: "correct horse battery staple",
      }),
    });
  });

  it("does not submit mismatched passwords", async () => {
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    await renderIdentity("/register");
    fillRegistration("different password");
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));

    expect((await screen.findByRole("alert")).textContent).toBe("Passwords do not match.");
    expect(fetch).not.toHaveBeenCalled();
  });

  it("shows an email conflict without leaving the registration form", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          Response.json({ message: "email is already registered" }, { status: 409 }),
        ),
    );
    const router = await renderIdentity("/register");
    fillRegistration();
    fireEvent.click(screen.getByRole("button", { name: "Create account" }));

    expect((await screen.findByRole("alert")).textContent).toBe("email is already registered");
    expect(router.state.location.pathname).toBe("/register");
  });
});
