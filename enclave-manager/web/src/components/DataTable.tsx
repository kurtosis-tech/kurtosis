import { TriangleDownIcon, TriangleUpIcon } from "@chakra-ui/icons";
import { chakra, Table, Tbody, Td, Th, Thead, Tr } from "@chakra-ui/react";
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  TableState,
  useReactTable,
} from "@tanstack/react-table";
import { type RowSelectionState } from "@tanstack/table-core/src/features/RowSelection";
import { type OnChangeFn } from "@tanstack/table-core/src/types";
import { useState } from "react";
import { assertDefined, isDefined } from "../utils";

export type DataTableProps<Data extends object> = {
  data: Data[];
  columns: ColumnDef<Data, any>[];
  defaultSorting?: SortingState;
  rowSelection?: Record<string, boolean>;
  onRowSelectionChange?: OnChangeFn<RowSelectionState>;
};

export function DataTable<Data extends object>({
  data,
  columns,
  defaultSorting,
  rowSelection,
  onRowSelectionChange,
}: DataTableProps<Data>) {
  if (isDefined(rowSelection) || isDefined(onRowSelectionChange)) {
    assertDefined(
      rowSelection,
      `rowSelection and onRowSelectionChange must both be defined in DataTable if either are defined.`,
    );
    assertDefined(
      onRowSelectionChange,
      `rowSelection and onRowSelectionChange must both be defined in DataTable if either are defined.`,
    );
  }
  const [sorting, setSorting] = useState<SortingState>(defaultSorting || []);
  const tableState: Partial<TableState> = { sorting };
  if (isDefined(rowSelection)) {
    tableState["rowSelection"] = rowSelection;
  }
  const table = useReactTable({
    columns,
    data,
    enableSortingRemoval: false,
    getCoreRowModel: getCoreRowModel(),
    onSortingChange: setSorting,
    getSortedRowModel: getSortedRowModel(),
    enableRowSelection: isDefined(rowSelection),
    onRowSelectionChange: onRowSelectionChange,
    state: tableState,
  });

  return (
    <Table variant={"kurtosis"}>
      <Thead>
        {table.getHeaderGroups().map((headerGroup) => (
          <Tr key={headerGroup.id}>
            {headerGroup.headers.map((header) => {
              // see https://tanstack.com/table/v8/docs/api/core/column-def#meta to type this correctly
              const meta: any = header.column.columnDef.meta;
              return (
                <Th key={header.id} onClick={header.column.getToggleSortingHandler()} isNumeric={meta?.isNumeric}>
                  {flexRender(header.column.columnDef.header, header.getContext())}
                  <chakra.span pl="4">
                    {header.column.getIsSorted() ? (
                      header.column.getIsSorted() === "desc" ? (
                        <TriangleDownIcon aria-label="sorted descending" />
                      ) : (
                        <TriangleUpIcon aria-label="sorted ascending" />
                      )
                    ) : null}
                  </chakra.span>
                </Th>
              );
            })}
          </Tr>
        ))}
      </Thead>
      <Tbody>
        {table.getRowModel().rows.map((row) => (
          <Tr key={row.id} bg={row.getIsSelected() ? "kurtosisSelected.100" : ""}>
            {row.getVisibleCells().map((cell) => {
              // see https://tanstack.com/table/v8/docs/api/core/column-def#meta to type this correctly
              const meta: any = cell.column.columnDef.meta;
              return (
                <Td key={cell.id} isNumeric={meta?.isNumeric}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </Td>
              );
            })}
          </Tr>
        ))}
      </Tbody>
    </Table>
  );
}
