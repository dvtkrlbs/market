import { Button } from "@astryxdesign/core/Button";
import { Heading } from "@astryxdesign/core/Heading";
import { TextInput } from "@astryxdesign/core/TextInput";
import * as stylex from "@stylexjs/stylex";
import { useForm } from "@tanstack/react-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { authErrorMessage, sessionQueryOptions, signIn } from "./auth";

const styles = stylex.create({
  form: { display: "flex", flexDirection: "column", gap: "var(--spacing-4)" },
});

export function LoginPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { registered } = useSearch({ from: "/login" });
  const mutation = useMutation({
    mutationFn: signIn,
    onSuccess: async () => {
      form.setFieldValue("password", "");
      await queryClient.cancelQueries({ queryKey: sessionQueryOptions.queryKey });
      queryClient.removeQueries({ queryKey: sessionQueryOptions.queryKey });
      await navigate({ to: "/session", replace: true });
    },
    retry: false,
  });

  const form = useForm({
    defaultValues: { email: "", password: "" },
    onSubmit: async ({ value }) => {
      if (!mutation.isPending) {
        await mutation
          .mutateAsync({ email: value.email.trim(), password: value.password })
          .catch(() => {});
      }
    },
  });

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        void form.handleSubmit();
      }}
      aria-busy={mutation.isPending}
      {...stylex.props(styles.form)}
    >
      <Heading level={2}>Sign in</Heading>
      <p>Use your Market account to sign in.</p>
      {registered && (
        <p role="status" aria-label="Registration complete">
          Account created. You can now sign in.
        </p>
      )}
      <form.Field name="email">
        {(field) => (
          <TextInput
            {...{ required: true, maxLength: 320 }}
            label="Email"
            htmlName="email"
            type="email"
            autoComplete="username"
            isRequired
            isDisabled={mutation.isPending}
            value={field.state.value}
            onChange={field.handleChange}
            onBlur={field.handleBlur}
          />
        )}
      </form.Field>
      <form.Field name="password">
        {(field) => (
          <TextInput
            {...{ required: true, minLength: 8, maxLength: 128 }}
            label="Password"
            htmlName="password"
            type="password"
            autoComplete="current-password"
            isRequired
            isDisabled={mutation.isPending}
            value={field.state.value}
            onChange={field.handleChange}
            onBlur={field.handleBlur}
          />
        )}
      </form.Field>
      {mutation.isError && <p role="alert">{authErrorMessage(mutation.error)}</p>}
      <Button
        type="submit"
        label={mutation.isPending ? "Signing in…" : "Sign in"}
        isLoading={mutation.isPending}
        isDisabled={mutation.isPending}
      />
      <p>
        Don’t have an account? <Link to="/register">Create an account</Link>
      </p>
    </form>
  );
}
