import { Button } from "@astryxdesign/core/Button";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type Store } from "shared/api";
import { approveStoreApplication, storeErrorMessage } from "shared/stores";

export function ApproveStoreButton({ store }: { store: Store }) {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: () => approveStoreApplication(store.id),
    onSuccess: async (approved) => {
      queryClient.setQueriesData<Store[]>({ queryKey: ["stores"] }, (current) =>
        current?.map((item) => (item.id === approved.id ? approved : item)),
      );
      await queryClient.invalidateQueries({ queryKey: ["stores"] });
    },
    retry: false,
  });

  if (store.approved) return null;
  return (
    <>
      <Button
        label={mutation.isPending ? "Approving…" : "Approve"}
        aria-label={`Approve ${store.name}`}
        isDisabled={mutation.isPending}
        isLoading={mutation.isPending}
        onClick={() => {
          if (!mutation.isPending) mutation.mutate();
        }}
      />
      {mutation.isError && <p role="alert">{storeErrorMessage(mutation.error)}</p>}
    </>
  );
}
