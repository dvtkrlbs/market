import { type QueryClient } from "@tanstack/react-query";
import { createRootRouteWithContext, Link } from "@tanstack/react-router";
import App from "../App";

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  component: App,
  errorComponent: () => (
    <main>
      <p role="alert">Unable to load your session. Please try signing in again.</p>
      <Link to="/login">Back to sign in</Link>
    </main>
  ),
  notFoundComponent: () => (
    <>
      <p>Page not found.</p>
      <Link to="/login">Go to sign in</Link>
    </>
  ),
});
