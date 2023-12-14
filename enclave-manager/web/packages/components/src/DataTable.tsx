import { Box, Button, Icon, Table, Tbody, Td, Th, Thead, Tr } from "@chakra-ui/react";
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  RowData,
  SortingState,
  TableState,
  useReactTable,
} from "@tanstack/react-table";
import { type RowSelectionState } from "@tanstack/table-core/src/features/RowSelection";
import { type OnChangeFn } from "@tanstack/table-core/src/types";
import { useState } from "react";
import { BiDownArrowAlt, BiUpArrowAlt } from "react-icons/bi";
import { assertDefined, isDefined } from "./utils";

declare module "@tanstack/table-core" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData extends RowData, TValue> {
    isNumeric?: boolean;
    centerAligned?: boolean;
  }
}

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
    <Box
      w={"100%"}
      h={"100%"}
      borderRadius={"6px"}
      border={"1px solid"}
      borderColor={"whiteAlpha.300"}
      overflow={"clip"}
    >
      <Table>
        <Thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <Tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => {
                const meta = header.column.columnDef.meta;
                return (
                  <Th
                    key={header.id}
                    onClick={header.column.getToggleSortingHandler()}
                    isNumeric={meta?.isNumeric}
                    textAlign={!!meta?.centerAligned ? "center" : undefined}
                  >
                    {header.column.getCanSort() && (
                      <Button
                        variant={"sortableHeader"}
                        size={"xs"}
                        rightIcon={
                          header.column.getIsSorted() === "desc" ? (
                            <Icon as={BiDownArrowAlt} color={"gray.400"} />
                          ) : header.column.getIsSorted() === "asc" ? (
                            <Icon as={BiUpArrowAlt} color={"gray.400"} />
                          ) : undefined
                        }
                      >
                        {flexRender(header.column.columnDef.header, header.getContext())}
                      </Button>
                    )}
                    {!header.column.getCanSort() && flexRender(header.column.columnDef.header, header.getContext())}
                  </Th>
                );
              })}
            </Tr>
          ))}
        </Thead>
        <Tbody>
          {table.getRowModel().rows.map((row) => (
            <Tr key={row.id} bg={row.getIsSelected() ? "gray.700" : ""}>
              {row.getVisibleCells().map((cell) => {
                const meta = cell.column.columnDef.meta;
                return (
                  <Td
                    key={cell.id}
                    isNumeric={meta?.isNumeric}
                    textAlign={!!meta?.centerAligned ? "center" : undefined}
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </Td>
                );
              })}
            </Tr>
          ))}
        </Tbody>
      </Table>
    </Box>
  );
}
