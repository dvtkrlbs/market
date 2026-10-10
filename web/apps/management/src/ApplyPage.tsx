import { Button } from "@astryxdesign/core/Button";
import { Heading } from "@astryxdesign/core/Heading";
import { TextInput } from "@astryxdesign/core/TextInput";
import { useForm } from "@tanstack/react-form";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { applyForStore, storeErrorMessage, storeSessionOptions } from "shared/stores";

export function ApplyPage() {
  const session = useQuery(storeSessionOptions);
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: async (name: string) => {
      if (!session.data) throw new Error("Missing user session");
      return applyForStore(session.data.user.id, name);
    },
    onSuccess: async () => {
      form.reset();
      await queryClient.invalidateQueries({ queryKey: ["stores", session.data?.user.id] });
    },
    retry: false,
  });
  const form = useForm({
    defaultValues: { name: "" },
    onSubmit: async ({ value }) => {
      if (!mutation.isPending && !mutation.isSuccess && session.data) {
        await mutation.mutateAsync(value.name).catch(() => {});
      }
    },
  });

  return (
    <section className="page-form">
      <Heading level={1}>Apply for a store</Heading>
      <p>
        Submit your store details for approval. Your store will appear to customers once approved.
      </p>
      {session.isPending ? (
        <p role="status">Loading your account…</p>
      ) : session.isError ? (
        <>
          <p role="alert">{storeErrorMessage(session.error)}</p>
          <Button
            label="Try again"
            onClick={() => {
              void session.refetch();
            }}
          />
        </>
      ) : mutation.isSuccess ? (
        <p role="status" aria-label="Application submitted">
          Application submitted for {mutation.data.name}. Your store is awaiting approval.
        </p>
      ) : (
        <form
          className="form-stack"
          aria-busy={mutation.isPending}
          onSubmit={(event) => {
            event.preventDefault();
            void form.handleSubmit();
          }}
        >
          <p>Applying as {session.data.user.email}.</p>
          <form.Field
            name="name"
            validators={{
              onChange: ({ value }) => (!value.trim() ? "Enter a store name." : undefined),
              onSubmit: ({ value }) => (!value.trim() ? "Enter a store name." : undefined),
            }}
          >
            {(field) => (
              <>
                <TextInput
                  label="Store name"
                  htmlName="name"
                  isRequired
                  {...{ required: true }}
                  value={field.state.value}
                  onChange={field.handleChange}
                  onBlur={field.handleBlur}
                  isDisabled={mutation.isPending}
                  status={field.state.meta.errors.length ? { type: "error" } : undefined}
                />
                {field.state.meta.errors.length > 0 && (
                  <p role="alert">{field.state.meta.errors[0]}</p>
                )}
              </>
            )}
          </form.Field>
          {mutation.isError && <p role="alert">{storeErrorMessage(mutation.error)}</p>}
          <Button
            type="submit"
            label={mutation.isPending ? "Submitting…" : "Submit application"}
            isDisabled={mutation.isPending}
            isLoading={mutation.isPending}
          />
        </form>
      )}
    </section>
  );
}
