import { Button } from "@astryxdesign/core/Button";
import { Heading } from "@astryxdesign/core/Heading";
import { TextInput } from "@astryxdesign/core/TextInput";
import * as stylex from "@stylexjs/stylex";
import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";
import { authErrorMessage, createAccount } from "./auth";

const styles = stylex.create({
  form: { display: "flex", flexDirection: "column", gap: "var(--spacing-4)" },
});

export function RegisterPage() {
  const navigate = useNavigate();
  const mutation = useMutation({
    mutationFn: createAccount,
    onSuccess: async () => {
      form.setFieldValue("password", "");
      form.setFieldValue("confirmation", "");
      await navigate({ to: "/login", search: { registered: true }, replace: true });
    },
    retry: false,
  });

  const form = useForm({
    defaultValues: { email: "", password: "", confirmation: "" },
    onSubmit: async ({ value }) => {
      if (mutation.isPending) return;
      mutation.reset();
      await mutation
        .mutateAsync({ email: value.email.trim(), password: value.password })
        .catch(() => {});
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
      <Heading level={2}>Create an account</Heading>
      <p>Register a Market account to get started.</p>
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
            description="Use between 8 and 128 characters."
            htmlName="password"
            type="password"
            autoComplete="new-password"
            isRequired
            isDisabled={mutation.isPending}
            value={field.state.value}
            onChange={field.handleChange}
            onBlur={field.handleBlur}
          />
        )}
      </form.Field>
      <form.Field
        name="confirmation"
        validators={{
          onChangeListenTo: ["password"],
          onChange: ({ value, fieldApi }) =>
            value !== fieldApi.form.getFieldValue("password")
              ? "Passwords do not match."
              : undefined,
          onSubmit: ({ value, fieldApi }) =>
            value !== fieldApi.form.getFieldValue("password")
              ? "Passwords do not match."
              : undefined,
        }}
      >
        {(field) => (
          <>
            <TextInput
              {...{ required: true, minLength: 8, maxLength: 128 }}
              label="Confirm password"
              htmlName="passwordConfirmation"
              type="password"
              autoComplete="new-password"
              isRequired
              isDisabled={mutation.isPending}
              value={field.state.value}
              onChange={field.handleChange}
              onBlur={field.handleBlur}
              status={field.state.meta.errors.length ? { type: "error" } : undefined}
            />
            {field.state.meta.errors.length > 0 && <p role="alert">{field.state.meta.errors[0]}</p>}
          </>
        )}
      </form.Field>
      {mutation.isError && <p role="alert">{authErrorMessage(mutation.error)}</p>}
      <Button
        type="submit"
        label={mutation.isPending ? "Creating account…" : "Create account"}
        isLoading={mutation.isPending}
        isDisabled={mutation.isPending}
      />
      <p>
        Already have an account? <Link to="/login">Sign in</Link>
      </p>
    </form>
  );
}
