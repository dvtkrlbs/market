import { createFileRoute, redirect } from "@tanstack/react-router";
import { AuthError, sessionQueryOptions } from "../auth";
import { SessionPage } from "../SessionPage";

export const Route = createFileRoute("/session")({
  beforeLoad: async ({ context }) => {
    try {
      await context.queryClient.ensureQueryData(sessionQueryOptions);
    } catch (error) {
      if (error instanceof AuthError && error.status === 401) {
        throw redirect({ to: "/login", replace: true });
      }
      throw error;
    }
  },
  component: SessionPage,
});
