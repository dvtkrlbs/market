import { Button } from "@astryxdesign/core/Button";
import { Heading } from "@astryxdesign/core/Heading";
import { TextInput } from "@astryxdesign/core/TextInput";
import * as stylex from "@stylexjs/stylex";
import { useMutation } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";
import { useState, type FormEvent } from "react";
import { authErrorMessage, createAccount } from "./auth";

const styles = stylex.create({
  form: { display: "flex", flexDirection: "column", gap: "var(--spacing-4)" },
});

export function RegisterPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [validationError, setValidationError] = useState<string>();
  const navigate = useNavigate();
  const mutation = useMutation({
    mutationFn: createAccount,
    onSuccess: async () => {
      setPassword("");
      setConfirmation("");
      await navigate({ to: "/login", search: { registered: true }, replace: true });
    },
    retry: false,
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (mutation.isPending) return;
    mutation.reset();
    if (password !== confirmation) {
      setValidationError("Passwords do not match.");
      return;
    }
    setValidationError(undefined);
    mutation.mutate({ email: email.trim(), password });
  }

  const error =
    validationError ?? (mutation.isError ? authErrorMessage(mutation.error) : undefined);

  return (
    <form onSubmit={handleSubmit} aria-busy={mutation.isPending} {...stylex.props(styles.form)}>
      <Heading level={2}>Create an account</Heading>
      <p>Register a Market account to get started.</p>
      <TextInput
        {...{ required: true, maxLength: 320 }}
        label="Email"
        htmlName="email"
        type="email"
        autoComplete="username"
        isRequired
        isDisabled={mutation.isPending}
        value={email}
        onChange={setEmail}
      />
      <TextInput
        {...{ required: true, minLength: 8, maxLength: 128 }}
        label="Password"
        description="Use between 8 and 128 characters."
        htmlName="password"
        type="password"
        autoComplete="new-password"
        isRequired
        isDisabled={mutation.isPending}
        value={password}
        onChange={setPassword}
      />
      <TextInput
        {...{ required: true, minLength: 8, maxLength: 128 }}
        label="Confirm password"
        htmlName="passwordConfirmation"
        type="password"
        autoComplete="new-password"
        isRequired
        isDisabled={mutation.isPending}
        value={confirmation}
        onChange={setConfirmation}
        status={validationError ? { type: "error" } : undefined}
      />
      {error && <p role="alert">{error}</p>}
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
