import { Button } from "@astryxdesign/core/Button";
import { Heading } from "@astryxdesign/core/Heading";
import { storeErrorMessage, useStores } from "shared/stores";

export function StoresPage() {
  const stores = useStores("shop");
  return (
    <section>
      <Heading level={1}>Stores</Heading>
      <p>Discover approved stores on Market.</p>
      {stores.isPending ? (
        <p role="status">Loading stores…</p>
      ) : stores.error ? (
        <>
          <p role="alert">{storeErrorMessage(stores.error)}</p>
          <Button label="Try again" onClick={stores.retry} />
        </>
      ) : stores.data?.length ? (
        <ul className="store-list">
          {stores.data.map((store) => (
            <li key={store.id} className="store-card">
              <Heading level={2}>{store.name}</Heading>
            </li>
          ))}
        </ul>
      ) : (
        <p>No stores available yet.</p>
      )}
    </section>
  );
}
