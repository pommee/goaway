"use client";

import * as React from "react";
import {
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable
} from "@tanstack/react-table";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { columns, Queries } from "@/app/logs/columns";
import { useEffect, useState } from "react";
import { DeleteRequest, GetRequest } from "@/util";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { toast } from "sonner";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger
} from "@/components/ui/tooltip";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTrigger
} from "@/components/ui/dialog";
import {
  CaretDoubleLeft,
  CaretDoubleRight,
  CaretDown,
  CaretLeft,
  CaretRight,
  Warning
} from "@phosphor-icons/react";

type QueryResponse = {
  details: Queries[];
  draw: string;
  recordsFiltered: number;
  recordsTotal: number;
};

async function fetchQueries(
  page: number,
  pageSize: number,
  domainFilter: string = ""
): Promise<QueryResponse> {
  try {
    let url = `queries?page=${page}&pageSize=${pageSize}`;
    if (domainFilter) {
      url += `&search=${encodeURIComponent(domainFilter)}`;
    }

    const [, response] = await GetRequest(url);

    if (response?.details && Array.isArray(response.details)) {
      return {
        details: response.details.map((item) => ({
          ...item,
          client: {
            ip: item.client?.ip || "",
            name: item.client?.name || "",
            mac: item.client?.mac || ""
          },
          ip: Array.isArray(item.ip) ? item.ip : []
        })),
        draw: response.draw || "1",
        recordsFiltered: response.recordsFiltered || 0,
        recordsTotal: response.recordsTotal || 0
      };
    } else {
      console.error("Invalid response format", response);
      return {
        details: [],
        draw: "1",
        recordsFiltered: 0,
        recordsTotal: 0
      };
    }
  } catch (error) {
    console.error("Failed to fetch queries:", error);
    return { details: [], draw: "1", recordsFiltered: 0, recordsTotal: 0 };
  }
}

export function Logs() {
  const [queries, setQueries] = useState<Queries[]>([]);
  const [pageIndex, setPageIndex] = useState(0);
  const [pageSize, setPageSize] = useState(10);
  const [totalRecords, setTotalRecords] = useState(0);
  const [loading, setLoading] = useState(true);
  const [domainFilter, setDomainFilter] = useState("");
  const [wsConnected, setWsConnected] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});

  const totalPages = Math.ceil(totalRecords / pageSize);

  const debounce = (func, delay) => {
    let timeoutId;
    return (...args) => {
      if (timeoutId) clearTimeout(timeoutId);
      timeoutId = setTimeout(() => {
        func(...args);
      }, delay);
    };
  };

  const debouncedSetDomainFilter = React.useMemo(
    () =>
      debounce((value) => {
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

        const formattedQuery = {
          ...newQuery,
          client: {
            ip: newQuery.client?.ip || "",
            name: newQuery.client?.name || "",
            mac: newQuery.client?.mac || ""
          },
          ip: Array.isArray(newQuery.ip) ? newQuery.ip : []
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

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        setWsConnected(false);
      };

      ws.onclose = () => {
        console.log("WebSocket connection closed");
        setWsConnected(false);
      };
    };

    return () => {
      if (ws) {
        ws.close();
      }
    };
  }, [pageIndex, pageSize, domainFilter]);

  useEffect(() => {
    async function fetchData() {
      setLoading(true);

      const result = await fetchQueries(pageIndex + 1, pageSize, domainFilter);

      setQueries(result.details);
      setTotalRecords(result.recordsFiltered);
      setLoading(false);
    }

    fetchData();
  }, [pageIndex, pageSize, domainFilter]);

  const table = useReactTable({
    data: queries,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    manualPagination: true,
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
    }
  });

  async function clearLogs() {
    const [responseCode] = await DeleteRequest("queries");
    if (responseCode === 200) {
      toast.success("Logs cleared successfully!");
      setQueries([]);
      setTotalRecords(0);
      setIsModalOpen(false);
    }
  }

  return (
    <div className="w-full">
      <div className="flex justify-between items-center mb-4">
        <div className="flex items-center">
          {wsConnected && (
            <div className="flex items-center mr-4">
              <div className="w-3 h-3 bg-green-500 rounded-full mr-2 animate-pulse"></div>
              <span className="text-sm text-green-500">Live updates</span>
            </div>
          )}
        </div>
        <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
          <DialogTrigger asChild>
            <Button
              variant="outline"
              className="bg-zinc-800 border-none hover:bg-zinc-700 text-white"
            >
              Clear logs
            </Button>
          </DialogTrigger>
          <DialogContent className="bg-zinc-900 text-white border-zinc-800 md:w-auto max-w-none">
            <div className="flex justify-center mb-4">
              <Warning className="h-10 w-10 text-amber-500" />
            </div>
            <DialogDescription className="text-base">
              <div className="bg-amber-600 border-2 border-amber-800 rounded-md p-4 mt-2">
                <p className="text-white">
                  Are you sure you want to clear all logs? This is an
                  irreversible action!
                </p>
              </div>
            </DialogDescription>
            <Button
              variant="outline"
              className="bg-red-800 hover:bg-red-700 text-white"
              onClick={clearLogs}
            >
              Yes
            </Button>
          </DialogContent>
        </Dialog>
      </div>

      <div className="flex items-center py-4">
        <Input
          placeholder="Filter domain..."
          onChange={(event) => debouncedSetDomainFilter(event.target.value)}
          className="max-w-sm"
        />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto">
              Columns <CaretDown />
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
      <div className="rounded-md border">
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
                              {flexRender(
                                cell.column.columnDef.cell,
                                cell.getContext()
                              )}
                            </span>
                          </TooltipTrigger>
                          <TooltipContent>
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
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
        <div className="flex-1 text-sm text-muted-foreground">
          Displaying {table.getPreSelectedRowModel().rows.length} of{" "}
          {totalRecords} record(s).
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
              <SelectTrigger className="h-8 w-[70px]">
                <SelectValue placeholder={pageSize} />
              </SelectTrigger>
              <SelectContent side="top">
                {[5, 10, 20, 30, 50].map((size) => (
                  <SelectItem key={size} value={`${size}`}>
                    {size}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex w-[100px] items-center justify-center text-sm font-medium">
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
              <CaretDoubleLeft />
            </Button>
            <Button
              variant="outline"
              className="h-8 w-8 p-0"
              onClick={() => setPageIndex((prev) => Math.max(0, prev - 1))}
              disabled={pageIndex === 0 || loading}
            >
              <span className="sr-only">Go to previous page</span>
              <CaretLeft />
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
              <CaretRight />
            </Button>
            <Button
              variant="outline"
              className="hidden h-8 w-8 p-0 lg:flex"
              onClick={() => setPageIndex(totalPages - 1)}
              disabled={pageIndex >= totalPages - 1 || loading}
            >
              <span className="sr-only">Go to last page</span>
              <CaretDoubleRight />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
