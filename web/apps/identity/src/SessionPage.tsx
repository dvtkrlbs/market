import { Button } from "@astryxdesign/core/Button";
import { Heading } from "@astryxdesign/core/Heading";
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { authErrorMessage, sessionQueryOptions, signOut } from "./auth";

export function SessionPage() {
  const { data: session } = useSuspenseQuery(sessionQueryOptions);
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const mutation = useMutation({
    mutationFn: signOut,
    onSuccess: async () => {
      await queryClient.cancelQueries({ queryKey: sessionQueryOptions.queryKey });
      queryClient.removeQueries({ queryKey: sessionQueryOptions.queryKey });
      await navigate({ to: "/login", replace: true });
    },
    retry: false,
  });

  return (
    <>
      <Heading level={2}>You’re signed in</Heading>
      <p>Signed in as {session.user.email}.</p>
      <p>
        Your session expires{" "}
        <time dateTime={session.sessionExpiresAt}>
          {new Date(session.sessionExpiresAt).toLocaleString()}
        </time>
        .
      </p>
      {mutation.isError && <p role="alert">{authErrorMessage(mutation.error)}</p>}
      <Button
        label={mutation.isPending ? "Signing out…" : "Sign out"}
        onClick={() => mutation.mutate()}
        isLoading={mutation.isPending}
        isDisabled={mutation.isPending}
      />
    </>
  );
}
