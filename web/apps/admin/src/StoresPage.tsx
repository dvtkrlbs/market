import { Button } from "@astryxdesign/core/Button";
import { Heading } from "@astryxdesign/core/Heading";
import {
  flexRender,
  createSortedRowModel,
  rowSortingFeature,
  sortFns,
  tableFeatures,
  useTable,
  type ColumnDef,
  type SortingState,
} from "@tanstack/react-table";
import { useState } from "react";
import { type Store } from "shared/api";
import { storeErrorMessage, useStores } from "shared/stores";

const features = tableFeatures({
  rowSortingFeature,
  sortedRowModel: createSortedRowModel(),
  sortFns,
});
const columns: ColumnDef<typeof features, Store>[] = [
  { accessorKey: "id", header: "ID" },
  { accessorKey: "name", header: "Store name" },
  {
    accessorKey: "approved",
    header: "Status",
    cell: ({ row }) => (row.original.approved ? "Approved" : "Awaiting approval"),
  },
];
const emptyStores: Store[] = [];

export function StoresPage() {
  const stores = useStores("admin");
  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useTable({
    features,
    data: stores.data ?? emptyStores,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getRowId: (store) => String(store.id),
  });

  return (
    <section>
      <Heading level={1}>Stores</Heading>
      <p>Stores you administer, including pending applications.</p>
      {stores.isPending ? (
        <p role="status">Loading stores…</p>
      ) : stores.error ? (
        <>
          <p role="alert">{storeErrorMessage(stores.error)}</p>
          <Button label="Try again" onClick={stores.retry} />
        </>
      ) : (
        <div className="table-scroll">
          <table className="stores-table">
            <caption>Stores you administer</caption>
            <thead>
              {table.getHeaderGroups().map((group) => (
                <tr key={group.id}>
                  {group.headers.map((header) => (
                    <th
                      key={header.id}
                      scope="col"
                      aria-sort={
                        header.column.getIsSorted() === "asc"
                          ? "ascending"
                          : header.column.getIsSorted() === "desc"
                            ? "descending"
                            : "none"
                      }
                    >
                      <button type="button" onClick={header.column.getToggleSortingHandler()}>
                        {flexRender(header.column.columnDef.header, header.getContext())}
                        {header.column.getIsSorted() === "asc"
                          ? " ↑"
                          : header.column.getIsSorted() === "desc"
                            ? " ↓"
                            : ""}
                      </button>
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody>
              {table.getRowModel().rows.length ? (
                table.getRowModel().rows.map((row) => (
                  <tr key={row.id}>
                    {row.getAllCells().map((cell) => (
                      <td key={cell.id}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </td>
                    ))}
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={columns.length}>No stores found.</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
