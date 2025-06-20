"use client";

import { columns, Queries } from "@/app/logs/columns";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger
} from "@/components/ui/tooltip";
import { DeleteRequest, GetRequest } from "@/util";
import {
  CaretDoubleLeftIcon,
  CaretDoubleRightIcon,
  CaretDownIcon,
  CaretLeftIcon,
  CaretRightIcon,
  WarningIcon
} from "@phosphor-icons/react";
import {
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  SortingState,
  useReactTable,
  VisibilityState
} from "@tanstack/react-table";
import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

interface Client {
  ip: string;
  name: string;
  mac: string;
}

export interface IPEntry {
  ip: string;
  rtype: string;
}

interface QueryDetail {
  id: number;
  domain: string;
  status: string;
  queryType: string;
  ip: IPEntry[];
  responseSizeBytes: number;
  timestamp: string;
  responseTimeNS: number;
  blocked: boolean;
  cached: boolean;
  client: Client;
}

interface QueryResponse {
  details: QueryDetail[];
  draw: string;
  recordsFiltered: number;
  recordsTotal: number;
}

async function fetchQueries(
  page: number,
  pageSize: number,
  domainFilter: string = "",
  sortField: string = "timestamp",
  sortDirection: string = "desc"
): Promise<QueryResponse> {
  try {
    let url = `queries?page=${page}&pageSize=${pageSize}&sortColumn=${encodeURIComponent(
      sortField
    )}&sortDirection=${encodeURIComponent(sortDirection)}`;
    if (domainFilter) {
      url += `&search=${encodeURIComponent(domainFilter)}`;
    }

    const [, response] = await GetRequest(url);

    if (response?.details && Array.isArray(response.details)) {
      return {
        details: response.details.map(
          (item: {
            client: { ip?: string; name?: string; mac?: string };
            ip?: { ip?: string; rtype?: string }[];
            [key: string]: string;
          }) => ({
            ...item,
            client: {
              ip: item.client?.ip || "",
              name: item.client?.name || "",
              mac: item.client?.mac || ""
            },
            ip: Array.isArray(item.ip)
              ? item.ip.map((entry) => ({
                  ip: String(entry?.ip || ""),
                  rtype: String(entry?.rtype || "")
                }))
              : []
          })
        ),
        draw: response.draw || "1",
        recordsFiltered: response.recordsFiltered || 0,
        recordsTotal: response.recordsTotal || 0
      };
    } else {
      return {
        details: [],
        draw: "1",
        recordsFiltered: 0,
        recordsTotal: 0
      };
    }
  } catch {
    return { details: [], draw: "1", recordsFiltered: 0, recordsTotal: 0 };
  }
}

export function Logs() {
  const [queries, setQueries] = useState<Queries[]>([]);
  const [pageIndex, setPageIndex] = useState(0);
  const [pageSize, setPageSize] = useState(15);
  const [totalRecords, setTotalRecords] = useState(0);
  const [loading, setLoading] = useState(true);
  const [domainFilter, setDomainFilter] = useState("");
  const [wsConnected, setWsConnected] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});
  const totalPages = Math.ceil(totalRecords / pageSize);
  const [sorting, setSorting] = useState<SortingState>([
    { id: "timestamp", desc: true }
  ]);

  const debounce = (func: (...args: unknown[]) => void, delay: number) => {
    let timeoutId: NodeJS.Timeout | undefined;
    return (...args: unknown[]) => {
      if (timeoutId) clearTimeout(timeoutId);
      timeoutId = setTimeout(() => {
        func(...args);
      }, delay);
    };
  };

  const debouncedSetDomainFilter = useMemo(
    () =>
      debounce((value: string) => {
        setDomainFilter(value);
        setPageIndex(0);
      }, 500),
    []
  );

  useEffect(() => {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/api/liveQueries`;
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setWsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const newQuery = JSON.parse(event.data);

        const formattedQuery: Queries = {
          ...newQuery,
          client: {
            ip: newQuery.client?.ip || "",
            name: newQuery.client?.name || "",
            mac: newQuery.client?.mac || ""
          },
          ip: Array.isArray(newQuery.ip)
            ? newQuery.ip.map((entry: IPEntry) => ({
                ip: String(entry?.ip || ""),
                rtype: String(entry?.rtype || "")
              }))
            : []
        };

        if (
          !domainFilter ||
          (formattedQuery.domain &&
            formattedQuery.domain
              .toLowerCase()
              .includes(domainFilter.toLowerCase()))
        ) {
          setQueries((prevQueries) => {
            const updatedQueries = [formattedQuery, ...prevQueries];
            if (updatedQueries.length > pageSize) {
              updatedQueries.pop();
            }
            return updatedQueries;
          });

          setTotalRecords((prev) => prev + 1);
        }
      } catch (error) {
        console.error("Error handling WebSocket message:", error);
      }
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
      setWsConnected(false);
    };

    ws.onclose = () => {
      console.log("WebSocket connection closed");
      setWsConnected(false);
    };

    return () => {
      if (ws) {
        ws.close();
      }
    };
  }, [pageIndex, pageSize, domainFilter]);

  const fetchData = useCallback(async () => {
    setLoading(true);

    const sortField = sorting.length > 0 ? sorting[0].id : "timestamp";
    const sortDirection =
      sorting.length > 0 ? (sorting[0].desc ? "desc" : "asc") : "desc";

    const result = await fetchQueries(
      pageIndex + 1,
      pageSize,
      domainFilter,
      sortField,
      sortDirection
    );

    setQueries(result.details);
    setTotalRecords(result.recordsFiltered);
    setLoading(false);
  }, [pageIndex, pageSize, domainFilter, sorting]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const table = useReactTable({
    data: queries,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    manualSorting: true,
    pageCount: totalPages,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
      pagination: {
        pageIndex,
        pageSize
      }
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection
  });

  async function clearLogs() {
    const [responseCode] = await DeleteRequest("queries", null);
    if (responseCode === 200) {
      toast.success("Logs cleared successfully!");
      setQueries([]);
      setTotalRecords(0);
      setIsModalOpen(false);
    }
  }

  return (
    <div className="w-full">
      <div className="flex items-center">
        <Input
          placeholder="Filter domain..."
          onChange={(event) => debouncedSetDomainFilter(event.target.value)}
          className="max-w-sm"
        />
        <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
          <DialogTrigger asChild className="ml-5">
            <Button
              disabled={queries.length === 0}
              variant="outline"
              className="bg-red-950 hover:bg-red-900 border-1 border-red-900 text-white"
            >
              Clear logs
            </Button>
          </DialogTrigger>
          <DialogContent className="bg-zinc-900 text-white border-zinc-800 md:w-auto max-w-md p-6 rounded-xl shadow-lg">
            <div className="flex flex-col items-center text-center">
              <WarningIcon className="h-12 w-12 text-amber-500 mb-4" />
              <DialogTitle className="text-xl font-semibold mb-2">
                Confirm Log Clearance
              </DialogTitle>
              <DialogDescription className="text-base text-zinc-300 mb-6">
                <div className="bg-red-600/50 p-4">
                  <p>Are you sure you want to clear all logs?</p>{" "}
                  <p>
                    This action is
                    <span className="font-semibold"> irreversible</span>.
                  </p>
                </div>
              </DialogDescription>
              <div className="flex gap-4">
                <Button
                  className="bg-red-700 hover:bg-red-600 text-white"
                  onClick={clearLogs}
                >
                  Yes, clear logs
                </Button>
                <Button
                  variant="ghost"
                  className="text-zinc-400 hover:text-white"
                  onClick={() => setIsModalOpen(false)}
                >
                  Cancel
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto">
              Columns <CaretDownIcon />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {table
              .getAllColumns()
              .filter((column) => column.getCanHide())
              .map((column) => (
                <DropdownMenuCheckboxItem
                  key={column.id}
                  className="capitalize"
                  checked={column.getIsVisible()}
                  onCheckedChange={(value) => column.toggleVisibility(!!value)}
                >
                  {column.id}
                </DropdownMenuCheckboxItem>
              ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <div className="rounded-md border mt-4">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  Loading...
                </TableCell>
              </TableRow>
            ) : queries.length > 0 ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  className={
                    row.index === 0 && wsConnected
                      ? "bg-zinc-700 bg-opacity-40 transition-colors duration-1000"
                      : ""
                  }
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell
                      className="max-w-60 truncate cursor-pointer"
                      key={cell.id}
                    >
                      {cell.column.id === "action" ||
                      cell.column.id === "responseSizeBytes" ||
                      cell.column.id === "queryType" ? (
                        <span className="block truncate">
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </span>
                      ) : (
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <span
                                ref={(el) => {
                                  if (el && el.scrollWidth > el.clientWidth) {
                                    el.setAttribute("data-truncated", "true");
                                  }
                                }}
                                className="block truncate"
                              >
                                {(() => {
                                  if (cell.column.id === "ip") {
                                    const ipValue =
                                      cell.getValue() as IPEntry[];
                                    if (
                                      Array.isArray(ipValue) &&
                                      ipValue.length > 0
                                    ) {
                                      return (
                                        <div className="flex items-center gap-1">
                                          <span>{ipValue[0]?.ip || ""}</span>
                                          {ipValue.length > 1 && (
                                            <span className="text-xs text-stone-400 border-1 ml-1 px-1 rounded border-green-600/60">
                                              +{ipValue.length - 1}
                                            </span>
                                          )}
                                        </div>
                                      );
                                    }
                                    return "";
                                  }
                                  return flexRender(
                                    cell.column.columnDef.cell,
                                    cell.getContext()
                                  );
                                })()}
                              </span>
                            </TooltipTrigger>
                            <TooltipContent className="bg-stone-800 border border-stone-700 text-white text-sm p-3 rounded-md shadow-md font-mono">
                              {(() => {
                                if (cell.column.id === "ip") {
                                  const ipValue = cell.getValue() as IPEntry[];
                                  return Array.isArray(ipValue) ? (
                                    <div className="space-y-1">
                                      {ipValue.map((entry, i) => (
                                        <div key={i} className="flex gap-2">
                                          <span className="inline-block w-[80px] text-stone-400">
                                            {entry?.rtype
                                              ? `[${entry.rtype}]`
                                              : ""}
                                          </span>
                                          <span>{entry?.ip || ""}</span>
                                        </div>
                                      ))}
                                    </div>
                                  ) : (
                                    ""
                                  );
                                }

                                return flexRender(
                                  cell.column.columnDef.cell,
                                  cell.getContext()
                                );
                              })()}
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No queries saved in the database.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between px-2 mt-4">
        <div className="flex text-sm text-muted-foreground">
          Displaying {table.getPreSelectedRowModel().rows.length} of{" "}
          {totalRecords.toLocaleString()} record(s).
          <div>
            <div className="flex items-center ml-5">
              {wsConnected ? (
                <>
                  <span className="flex text-sm text-green-500/50">
                    <div className="w-3 h-3 bg-green-500/50 rounded-full mr-2 mt-1 animate-pulse"></div>
                    Live updates
                  </span>
                </>
              ) : (
                <>
                  <div className="w-3 h-3 bg-red-500/50 rounded-full mr-2"></div>
                  <span className="text-sm text-red-500/50">
                    no websocket connection
                  </span>
                </>
              )}
            </div>
          </div>
        </div>
        <div className="flex items-center space-x-6 lg:space-x-8">
          <div className="flex items-center space-x-2">
            <p className="text-sm font-medium">Rows per page</p>
            <Select
              value={`${pageSize}`}
              onValueChange={(value) => {
                setPageSize(Number(value));
                setPageIndex(0);
              }}
            >
              <SelectTrigger className="h-8 fit-content">
                <SelectValue placeholder={pageSize} />
              </SelectTrigger>
              <SelectContent side="top">
                {[5, 15, 30, 50, 100, 250].map((size) => (
                  <SelectItem key={size} value={`${size}`}>
                    {size}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex w-[120px] items-center justify-center text-sm font-medium">
            Page {pageIndex + 1} of {totalPages || 1}
          </div>
          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              className="hidden h-8 w-8 p-0 lg:flex"
              onClick={() => setPageIndex(0)}
              disabled={pageIndex === 0 || loading}
            >
              <span className="sr-only">Go to first page</span>
              <CaretDoubleLeftIcon />
            </Button>
            <Button
              variant="outline"
              className="h-8 w-8 p-0"
              onClick={() => setPageIndex((prev) => Math.max(0, prev - 1))}
              disabled={pageIndex === 0 || loading}
            >
              <span className="sr-only">Go to previous page</span>
              <CaretLeftIcon />
            </Button>
            <Button
              variant="outline"
              className="h-8 w-8 p-0"
              onClick={() =>
                setPageIndex((prev) => Math.min(totalPages - 1, prev + 1))
              }
              disabled={pageIndex >= totalPages - 1 || loading}
            >
              <span className="sr-only">Go to next page</span>
              <CaretRightIcon />
            </Button>
            <Button
              variant="outline"
              className="hidden h-8 w-8 p-0 lg:flex"
              onClick={() => setPageIndex(totalPages - 1)}
              disabled={pageIndex >= totalPages - 1 || loading}
            >
              <span className="sr-only">Go to last page</span>
              <CaretDoubleRightIcon />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
