"use client";

import { CardDetails } from "@/app/clients/details";
import { Queries } from "@/app/logs/columns";
import { columns } from "@/app/logs/columnsData";
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
import { NoContent } from "@/shared";
import { DeleteRequest, GetRequest } from "@/util";
import {
  CaretDoubleLeftIcon,
  CaretDoubleRightIcon,
  CaretDownIcon,
  CaretLeftIcon,
  CaretRightIcon,
  QuestionIcon,
  WarningIcon,
  MagnifyingGlassIcon
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
  queries: QueryDetail[];
  draw: string;
  recordsFiltered: number;
  recordsTotal: number;
}

async function fetchQueries(
  page: number,
  pageSize: number,
  domainFilter: string = "",
  clientFilter: string = "",
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

    if (clientFilter) {
      url += `&client=${encodeURIComponent(clientFilter)}`;
    }

    const [, response] = await GetRequest(url);

    if (response?.queries && Array.isArray(response.queries)) {
      return {
        queries: response.queries.map(
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
        queries: [],
        draw: "1",
        recordsFiltered: 0,
        recordsTotal: 0
      };
    }
  } catch {
    return { queries: [], draw: "1", recordsFiltered: 0, recordsTotal: 0 };
  }
}

function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}

export function Logs() {
  const [queries, setQueries] = useState<Queries[]>([]);
  const [pageIndex, setPageIndex] = useState(0);
  const [pageSize, setPageSize] = useState(() => {
    const saved = localStorage.getItem("logsPageSize");
    return saved ? Number(saved) : 15;
  });
  const [totalRecords, setTotalRecords] = useState(0);
  const [loading, setLoading] = useState(true);
  const [filterLoading, setFilterLoading] = useState(false);

  const [domainInputValue, setDomainInputValue] = useState("");
  const [domainFilter, setDomainFilter] = useState("");
  const debouncedDomainFilter = useDebounce(domainInputValue, 200);

  const [clientInputValue, setClientInputValue] = useState("");
  const [clientFilter, setClientFilter] = useState("");
  const debouncedClientFilter = useDebounce(clientInputValue, 200);

  const [wsConnected, setWsConnected] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});

  const [selectedClient, setSelectedClient] = useState<Client | null>(null);
  const [isClientDetailsOpen, setIsClientDetailsOpen] = useState(false);

  const [showHelp, setShowHelp] = useState(false);

  const totalPages = Math.ceil(totalRecords / pageSize);
  const [sorting, setSorting] = useState<SortingState>([
    { id: "timestamp", desc: true }
  ]);

  const showClientDetails = useCallback((client: Client) => {
    setSelectedClient(client);
    setIsClientDetailsOpen(true);
  }, []);

  const handleDomainInputChange = (value: string) => {
    setDomainInputValue(value);
    if (value !== domainFilter) {
      setFilterLoading(true);
    }
  };

  const handleClientInputChange = (value: string) => {
    setClientInputValue(value);
    if (value !== clientFilter) {
      setFilterLoading(true);
    }
  };

  useEffect(() => {
    if (
      debouncedDomainFilter !== domainFilter ||
      debouncedClientFilter !== clientFilter
    ) {
      setDomainFilter(debouncedDomainFilter);
      setClientFilter(debouncedClientFilter);
      setPageIndex(0);
      setFilterLoading(false);
    }
  }, [
    debouncedDomainFilter,
    domainFilter,
    debouncedClientFilter,
    clientFilter
  ]);

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
      clientFilter,
      sortField,
      sortDirection
    );

    setQueries(result.queries);
    setTotalRecords(result.recordsFiltered);
    setLoading(false);
  }, [pageIndex, pageSize, domainFilter, clientFilter, sorting]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const columnsWithClientHandler = useMemo(() => {
    return columns.map((column) => {
      if (column.id === "client") {
        return {
          ...column,
          cell: ({ row }) => {
            const client = row.original.client;
            return (
              <div
                onClick={() => showClientDetails(client)}
                className="cursor-pointer text-blue-300 hover:text-blue-500 transition-colors"
              >
                {client.name} | {client.ip}
              </div>
            );
          }
        };
      }
      return column;
    });
  }, [showClientDetails]);

  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data: queries,
    columns: columnsWithClientHandler,
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
        <QuestionIcon
          size={20}
          className="mr-4 hover:text-orange-400 cursor-pointer transition-colors"
          onClick={() => setShowHelp(true)}
        />

        <div className="relative max-w-sm">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="Filter domain..."
            value={domainInputValue}
            onChange={(event) => handleDomainInputChange(event.target.value)}
            className="pl-10 pr-10 transition-all duration-200 focus:ring-2 focus:ring-primary/20"
          />
          {(filterLoading || domainInputValue !== domainFilter) && (
            <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
              <div className="animate-spin rounded-full h-4 w-4 border-2 border-primary border-t-transparent"></div>
            </div>
          )}
          {domainInputValue &&
            !filterLoading &&
            domainInputValue === domainFilter && (
              <button
                onClick={() => {
                  setDomainInputValue("");
                  setDomainFilter("");
                  setPageIndex(0);
                }}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
              >
                <span className="sr-only">Clear filter</span>×
              </button>
            )}
        </div>

        <div className="ml-5 relative max-w-sm">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="Filter client..."
            value={clientInputValue}
            onChange={(event) => handleClientInputChange(event.target.value)}
            className="pl-10 pr-10 transition-all duration-200 focus:ring-2 focus:ring-primary/20"
          />
          {clientInputValue &&
            !filterLoading &&
            clientInputValue === clientFilter && (
              <button
                onClick={() => {
                  setClientInputValue("");
                  setClientFilter("");
                  setPageIndex(0);
                }}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
              >
                <span className="sr-only">Clear filter</span>×
              </button>
            )}
        </div>

        {clientFilter && (
          <div className="ml-3 flex items-center text-sm text-muted-foreground animate-in fade-in-50 slide-in-from-left-2 duration-200">
            <span className="bg-primary/10 text-primary px-2 py-1 rounded-md border">
              Filtered: "{clientFilter}"
            </span>
          </div>
        )}

        <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
          <DialogTrigger asChild className="ml-5">
            <Button disabled={queries.length === 0} variant="destructive">
              Clear
            </Button>
          </DialogTrigger>
          <DialogContent className="md:w-auto max-w-md p-6 rounded-xl shadow-lg">
            <div className="flex flex-col items-center text-center">
              <WarningIcon className="h-12 w-12 text-amber-500 mb-4" />
              <DialogTitle className="text-xl font-semibold mb-2">
                Confirm Log Clearance
              </DialogTitle>
              <DialogDescription className="text-base mb-6">
                <div className="bg-destructive/20 border border-destructive text-destructive p-4 rounded-xl">
                  <p>Are you sure you want to clear all logs?</p>{" "}
                  <p>
                    This action is{" "}
                    <span className="font-bold underline">irreversible</span>.
                  </p>
                </div>
              </DialogDescription>
              <div className="flex gap-4">
                <Button
                  variant="outline"
                  className="hover:font-bold transition-all duration-200"
                  onClick={() => setIsModalOpen(false)}
                >
                  Cancel
                </Button>
                <Button
                  variant="destructive"
                  onClick={clearLogs}
                  className="hover:font-bold transition-all duration-200 bg-destructive/20"
                >
                  Yes, clear logs
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="outline"
              className="ml-auto transition-all duration-200 hover:scale-105"
            >
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

      <div
        className="rounded-md border mt-4 transition-opacity duration-200"
        style={{ opacity: loading ? 0.7 : 1 }}
      >
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
                  colSpan={columnsWithClientHandler.length}
                  className="h-24 text-center"
                >
                  <div className="flex items-center justify-center space-x-2">
                    <div className="animate-spin rounded-full h-6 w-6 border-2 border-primary border-t-transparent"></div>
                    <span>Loading...</span>
                  </div>
                </TableCell>
              </TableRow>
            ) : queries.length > 0 ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  className={
                    row.index === 0 && wsConnected
                      ? "bg-zinc-700 bg-opacity-40 transition-colors duration-1000"
                      : "transition-colors duration-200 hover:bg-muted/50"
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
                                        <div className="flex items-center gap-1 min-w-0">
                                          <span className="truncate flex-1 min-w-0">
                                            {ipValue[0]?.ip || ""}
                                          </span>
                                          {ipValue.length > 1 && (
                                            <span className="text-xs text-muted-foreground bg-card border px-1 rounded border-muted shrink-0">
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
                  colSpan={columnsWithClientHandler.length}
                  className="h-24 text-center"
                >
                  {domainFilter ? (
                    <div className="flex flex-col items-center space-y-2">
                      <span>
                        No queries found matching <b>{domainFilter}</b>
                      </span>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          setDomainInputValue("");
                          setDomainFilter("");
                          setPageIndex(0);
                        }}
                      >
                        Clear filter
                      </Button>
                    </div>
                  ) : (
                    <NoContent text="No queries recored" />
                  )}
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
                    live feed disabled
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
                const newPageSize = Number(value);
                setPageSize(newPageSize);
                localStorage.setItem("logsPageSize", String(newPageSize));
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
              className="hidden h-8 w-8 p-0 lg:flex transition-all duration-200 hover:scale-110"
              onClick={() => setPageIndex(0)}
              disabled={pageIndex === 0 || loading}
            >
              <span className="sr-only">Go to first page</span>
              <CaretDoubleLeftIcon />
            </Button>
            <Button
              variant="outline"
              className="h-8 w-8 p-0 transition-all duration-200 hover:scale-110"
              onClick={() => setPageIndex((prev) => Math.max(0, prev - 1))}
              disabled={pageIndex === 0 || loading}
            >
              <span className="sr-only">Go to previous page</span>
              <CaretLeftIcon />
            </Button>
            <Button
              variant="outline"
              className="h-8 w-8 p-0 transition-all duration-200 hover:scale-110"
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
              className="hidden h-8 w-8 p-0 lg:flex transition-all duration-200 hover:scale-110"
              onClick={() => setPageIndex(totalPages - 1)}
              disabled={pageIndex >= totalPages - 1 || loading}
            >
              <span className="sr-only">Go to last page</span>
              <CaretDoubleRightIcon />
            </Button>
          </div>
        </div>
      </div>

      {selectedClient && (
        <Dialog
          open={isClientDetailsOpen}
          onOpenChange={setIsClientDetailsOpen}
        >
          <DialogContent>
            <CardDetails
              ip={selectedClient.ip}
              lastSeen={"N/A"}
              mac={selectedClient.mac}
              name={selectedClient.name}
              vendor={""}
              onClose={() => setIsClientDetailsOpen(false)}
            />
          </DialogContent>
        </Dialog>
      )}

      {showHelp && (
        <Dialog open={showHelp} onOpenChange={setShowHelp}>
          <DialogContent className="max-w-4xl max-h-4/5 overflow-y-auto bg-transparent backdrop-blur-sm">
            <DialogTitle>Log Table Help</DialogTitle>
            <DialogDescription>
              The log table contains a couple of columns which sometimes serves
              a multi-purpose. This help box will explain what each one means
              and can do.
            </DialogDescription>

            <li>
              <b>Timestamp</b>
              <br /> Specifies when the query was sent by the client.
            </li>

            <li>
              <b>Domain</b>
              <br /> Domain name the client has requested
            </li>

            <li>
              <b>IP(s)</b>
              <br /> Response given back to the client. This can contain
              multiple response types inside the same request, indicated by the
              '+N'. Hovering over will reveal all resolved IP's.
            </li>

            <li>
              <b>Client</b>
              <br /> Here the client hostname and IP will be shown. It is
              possible to click the client to show a modal about the client
              where further actions and information is available.
            </li>

            <li>
              <b>Status</b>
              <br /> This column will indicate multiple factors of the request
              and response.
              <li className="ml-4">
                <b>ok / blacklisted / cached</b> - The request was fully
                processed, blacklisted or found in cache. In all cases the
                client receives a response; only 'blacklisted' differs as the IP
                will always be '0.0.0.0'.
              </li>
              <li className="ml-4">
                <b>Response Status</b> - This specifies whether a request was
                sucessfully fulfilled, failed or otherwise. Most common types
                are:
                <ul className="ml-4">
                  <li>
                    <b>NoError - </b>
                    Request was sucesfully fulfilled without any error.
                  </li>
                  <li>
                    <b>NXDomain</b> - Either a blacklisted domain or it was not
                    found.
                  </li>
                </ul>
              </li>
            </li>

            <li>
              <b>Response</b>
              <br /> Time taken to fully process a request from once the request
              is received to once the server responds.
            </li>

            <li>
              <b>Type</b>
              <br /> Response type given back to the client. Most common types
              are:
              <li className="ml-4">
                <b>A</b> - The IPv4 address
              </li>
              <li className="ml-4">
                <b>AAAA</b> - The IPv6 address
              </li>
              <li className="ml-4">
                <b>CNAME</b> - A domain name alias
              </li>
            </li>

            <li>
              <b>Protocol</b>
              <br /> Protocol used while processing the request to an upstream
              server. Most common one is UDP, however TCP, TLS, dns-over-tcp
              (DoT) and dns-over-https (DoH) is also available.
            </li>

            <li>
              <b>Size</b>
              <br /> Response size in bytes given back to the client.
            </li>

            <li>
              <b>Action</b>
              <br /> Here it is possible to toggle the status of a domain name.
              For example if the domain is whitelisted, then it can be
              blacklisted and vice versa.
            </li>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
