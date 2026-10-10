import { createFileRoute } from "@tanstack/react-router";
import { LoginPage } from "../LoginPage";

export const Route = createFileRoute("/login")({
  validateSearch: (search: Record<string, unknown>): { registered?: boolean } => ({
    registered: search.registered === true || search.registered === "true" ? true : undefined,
  }),
  component: LoginPage,
});
