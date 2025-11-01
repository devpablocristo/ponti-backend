import React, { useEffect, useState, useRef } from "react";
import {
  Edit,
  Copy,
  Filter,
  FilterX,
  ArrowUp,
  ArrowDown,
  ArrowUpDown,
} from "lucide-react";

type FilterType = "text" | "number" | "select" | "date";

type Column<T> = {
  key: keyof T;
  header: string;
  render?: (value: T[keyof T], item: T) => React.ReactNode;
  filterable?: boolean;
  filterType?: FilterType;
  filterOptions?: string[];
  sortable?: boolean;
};

type DataTableProps<T> = {
  data: T[];
  filters?: Record<string, any>;
  onFilterChange?: (filters: Record<string, any>) => void;
  columns: Column<T>[];
  headerComponent?: React.ReactNode;
  expandableRowRender?: (item: T) => React.ReactNode;
  onEdit?: (item: T) => void;
  onCopy?: (item: T) => void;
  onDelete?: (item: T) => void;
  className?: string;
  pagination?: {
    page: number;
    perPage: number;
    total: number;
    onPageChange: (page: number) => void;
  };
  message?: string;
  enableFilters?: boolean;
};

function getPaginationRange(
  totalPages: number,
  currentPage: number
): (number | null)[] {
  const delta = 1;
  const range: (number | null)[] = [];
  const rangeWithDots: (number | null)[] = [];
  let l: number | undefined = undefined;

  for (let i = 1; i <= totalPages; i++) {
    if (
      i === 1 ||
      i === totalPages ||
      (i >= currentPage - delta && i <= currentPage + delta)
    ) {
      range.push(i);
    }
  }

  for (let i = 0; i < range.length; i++) {
    if (typeof l === "number") {
      if ((range[i] as number) - l === 2) {
        rangeWithDots.push(l + 1);
      } else if ((range[i] as number) - l > 2) {
        rangeWithDots.push(null);
      }
    }
    rangeWithDots.push(range[i]);
    l = range[i] as number;
  }
  return rangeWithDots;
}

const DataTable = <T,>({
  data,
  filters,
  onFilterChange,
  columns,
  headerComponent,
  expandableRowRender,
  onEdit,
  onDelete,
  onCopy,
  className,
  pagination,
  message = "No hay proyectos disponibles",
  enableFilters = false,
}: DataTableProps<T>) => {
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc" | null>(
    null
  );

  const toggleSort = (key: string) => {
    if (sortKey !== key) {
      setSortKey(key);
      setSortDirection("asc");
    } else if (sortDirection === "asc") {
      setSortDirection("desc");
    } else if (sortDirection === "desc") {
      setSortKey(null);
      setSortDirection(null);
    } else {
      setSortDirection("asc");
    }
  };

  const sortedData = React.useMemo(() => {
    if (!sortKey || !sortDirection) return data;
    const col = columns.find((c) => String(c.key) === sortKey);
    if (!col) return data;
    return [...data].sort((a, b) => {
      const aValue = a[col.key];
      const bValue = b[col.key];
      if (aValue == null && bValue == null) return 0;
      if (aValue == null) return sortDirection === "asc" ? -1 : 1;
      if (bValue == null) return sortDirection === "asc" ? 1 : -1;

      const aNum = Number(aValue);
      const bNum = Number(bValue);
      const aIsNum = !isNaN(aNum) && aValue !== "";
      const bIsNum = !isNaN(bNum) && bValue !== "";

      if (aIsNum && bIsNum) {
        return sortDirection === "asc" ? aNum - bNum : bNum - aNum;
      }

      return sortDirection === "asc"
        ? String(aValue).localeCompare(String(bValue))
        : String(bValue).localeCompare(String(aValue));
    });
  }, [data, sortKey, sortDirection, columns]);

  const [expandedRow, setExpandedRow] = useState<number | null>(null);

  const [activeFilter, setActiveFilter] = useState<string | null>(null);
  const filterRefs = useRef<Record<string, HTMLDivElement | null>>({});

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        activeFilter &&
        !Object.values(filterRefs.current).some(
          (ref) => ref && ref.contains(event.target as Node)
        )
      ) {
        setActiveFilter(null);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [activeFilter]);

  const toggleRow = (index: number) => {
    setExpandedRow(expandedRow === index ? null : index);
  };

  const toggleFilter = (key: string) => {
    if (activeFilter === key) {
      setActiveFilter(null);
    } else {
      setActiveFilter(key);
    }
  };

  const handleFilterChange = (key: string, value: any) => {
    onFilterChange?.((prev: Record<string, any>) => ({
      ...prev,
      [key]: value,
    }));
  };

  const clearFilter = (key: string) => {
    onFilterChange?.((prev: Record<string, any>) => {
      const newFilters = { ...prev };
      delete newFilters[key];
      return newFilters;
    });
  };

  return (
    <div
      className={`relative overflow-x-auto shadow-md sm:rounded-lg rounded-xl border border-gray-100 overflow-hidden ${className}`}
    >
      {headerComponent && <div>{headerComponent}</div>}
      <div className="overflow-auto flex-1 w-full min-h-[250px] bg-white">
        <table className="w-full text-sm text-left text-gray-700">
          <thead className="text-xs text-white font-bold uppercase bg-custom-table-header">
            <tr>
              {expandableRowRender && <th className="w-8 p-2"></th>}
              {columns.map((column) => {
                const isFilterActive =
                  filters?.[String(column.key)] &&
                  filters[String(column.key)] !== "";
                const isSorted = sortKey === String(column.key);
                return (
                  <th key={String(column.key)} className="p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-1">
                        {/* Ícono de ordenamiento */}
                        {column.sortable !== false && (
                          <button
                            onClick={() => toggleSort(String(column.key))}
                            className={`mr-1 focus:outline-none ${
                              isSorted
                                ? "text-blue-600"
                                : "text-gray-300 hover:text-blue-400"
                            }`}
                            title={
                              isSorted
                                ? sortDirection === "asc"
                                  ? "Orden ascendente"
                                  : "Orden descendente"
                                : "Ordenar"
                            }
                          >
                            {isSorted ? (
                              sortDirection === "asc" ? (
                                <ArrowUp size={16} />
                              ) : (
                                <ArrowDown size={16} />
                              )
                            ) : (
                              <ArrowUpDown size={16} />
                            )}
                          </button>
                        )}
                        {column.header}
                      </div>
                      {enableFilters && column.filterable !== false && (
                        <div
                          className="relative"
                          ref={(el) => {
                            filterRefs.current[String(column.key)] = el;
                          }}
                        >
                          <button
                            onClick={() => toggleFilter(String(column.key))}
                            className={`ml-1 relative ${
                              activeFilter === String(column.key) ||
                              isFilterActive
                                ? "text-blue-500"
                                : "text-gray-200 hover:text-blue-600"
                            } focus:outline-none`}
                            title={
                              isFilterActive
                                ? `Filtro: ${filters[String(column.key)]}`
                                : "Filtrar"
                            }
                          >
                            {activeFilter === String(column.key) ? (
                              <FilterX size={14} />
                            ) : (
                              <Filter size={14} />
                            )}
                            {isFilterActive && (
                              <span className="absolute top-0 right-0 block h-2 w-2 rounded-full bg-blue-500 border border-white"></span>
                            )}
                          </button>

                          {activeFilter === String(column.key) && (
                            <div
                              className={`absolute z-[9999] ${
                                columns.indexOf(column) === 0
                                  ? "left-0"
                                  : "right-0"
                              } mt-2 w-48 bg-white rounded-md shadow-lg p-2 border border-gray-200`}
                            >
                              <div className="p-2">
                                <label className="block text-gray-700 text-xs mb-1">
                                  Filter
                                </label>

                                {column.filterType === "select" &&
                                column.filterOptions ? (
                                  <select
                                    className="border border-gray-300 rounded px-2 py-1 w-full text-xs text-gray-700"
                                    value={filters?.[String(column.key)] || ""}
                                    onChange={(e) =>
                                      handleFilterChange(
                                        String(column.key),
                                        e.target.value
                                      )
                                    }
                                  >
                                    <option value="">All</option>
                                    {column.filterOptions.map((option) => (
                                      <option key={option} value={option}>
                                        {option}
                                      </option>
                                    ))}
                                  </select>
                                ) : column.filterType === "date" ? (
                                  <input
                                    type="date"
                                    className="border border-gray-300 rounded px-2 py-1 w-full text-xs text-gray-700"
                                    value={filters?.[String(column.key)] || ""}
                                    onChange={(e) =>
                                      handleFilterChange(
                                        String(column.key),
                                        e.target.value
                                      )
                                    }
                                  />
                                ) : column.filterType === "number" ? (
                                  <input
                                    type="number"
                                    className="border border-gray-300 rounded px-2 py-1 w-full text-xs text-gray-700"
                                    value={filters?.[String(column.key)] || ""}
                                    onChange={(e) =>
                                      handleFilterChange(
                                        String(column.key),
                                        e.target.value
                                      )
                                    }
                                  />
                                ) : (
                                  <input
                                    type="text"
                                    className="border border-gray-300 rounded px-2 py-1 w-full text-xs text-gray-700"
                                    placeholder="Search..."
                                    value={filters?.[String(column.key)] || ""}
                                    onChange={(e) =>
                                      handleFilterChange(
                                        String(column.key),
                                        e.target.value
                                      )
                                    }
                                  />
                                )}

                                <div className="flex justify-between mt-2">
                                  <button
                                    className="bg-gray-200 text-xs rounded px-2 py-1 text-gray-700 hover:bg-gray-300"
                                    onClick={() =>
                                      clearFilter(String(column.key))
                                    }
                                  >
                                    Clear
                                  </button>
                                  <button
                                    className="bg-blue-500 text-xs rounded px-2 py-1 text-white hover:bg-blue-600"
                                    onClick={() => setActiveFilter(null)}
                                  >
                                    Apply
                                  </button>
                                </div>
                              </div>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  </th>
                );
              })}
              {(onEdit || onDelete || onCopy) && (
                <th className="p-4 text-center"></th>
              )}
            </tr>
          </thead>
          <tbody>
            {sortedData.length > 0 ? (
              sortedData.map((item, index) => (
                <React.Fragment key={index}>
                  <tr
                    key={index}
                    className={`border-t border-gray-100 text-gray-900 ${
                      index % 2 === 0 ? "bg-white" : "bg-[#EBF5FF]"
                    }`}
                  >
                    {expandableRowRender && (
                      <td
                        className="px-4 py-3 cursor-pointer"
                        onClick={(e) => {
                          e.stopPropagation();
                          toggleRow(index);
                        }}
                      >
                        <svg
                          className={`w-4 h-4 transition-transform duration-200 ${
                            expandedRow === index ? "rotate-90" : ""
                          }`}
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                          xmlns="http://www.w3.org/2000/svg"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M9 5l7 7-7 7"
                          />
                        </svg>
                      </td>
                    )}
                    {columns.map((column) => (
                      <td
                        key={String(column.key)}
                        className="px-6 py-4 min-w-[100px] max-w-[180px] whitespace-nowrap truncate"
                        title={String(item[column.key])}
                      >
                        {column.render
                          ? column.render(item[column.key], item)
                          : String(item[column.key])}
                      </td>
                    ))}
                    {(onEdit || onDelete || onCopy) && (
                      <td className="px-6 py-4 text-center">
                        <div className="flex justify-center space-x-2">
                          {onEdit && (
                            <button
                              onClick={() => onEdit(item)}
                              className="font-medium text-gray-800 hover:text-gray-900 hover:underline mr-3"
                              title="Editar proyecto"
                            >
                              <Edit size={16} />
                            </button>
                          )}
                          {onCopy && (
                            <button
                              onClick={() => onCopy(item)}
                              className="flex items-center gap-1 text-blue-700 hover:text-blue-900"
                              title="Duplicar proyecto"
                            >
                              <Copy size={16} />
                            </button>
                          )}
                          {onDelete && (
                            <button
                              onClick={() => onDelete(item)}
                              className="font-medium text-red-600 dark:text-red-500 hover:underline"
                            >
                              <svg
                                className="w-5 h-5"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                                xmlns="http://www.w3.org/2000/svg"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth="2"
                                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                ></path>
                              </svg>
                            </button>
                          )}
                        </div>
                      </td>
                    )}
                  </tr>
                  {expandableRowRender && expandedRow === index && (
                    <tr className="bg-white">
                      <td
                        colSpan={
                          columns.length + 1 + (onEdit || onDelete ? 1 : 0)
                        }
                      >
                        <div className="p-4">{expandableRowRender(item)}</div>
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              ))
            ) : (
              <tr>
                <td colSpan={columns.length + 1 + (onEdit || onDelete ? 1 : 0)}>
                  <div className="p-4 text-center">{message}</div>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="bg-white z-10 border-t sticky bottom-0">
        {pagination && data.length > 0 && (
          <nav
            className="bg-white flex flex-col md:flex-row justify-between items-start md:items-center space-y-3 md:space-y-0 p-4"
            aria-label="Table navigation"
          >
            <span className="text-sm font-normal text-gray-500">
              Mostrar
              <span className="mx-1 font-semibold text-gray-900">
                {(pagination.page - 1) * pagination.perPage + 1}-
                {Math.min(
                  pagination.page * pagination.perPage,
                  pagination.total
                )}
              </span>
              de
              <span className="ml-1 font-semibold text-gray-900">
                {pagination.total}
              </span>
            </span>
            <ul className="inline-flex items-stretch -space-x-px">
              <li>
                <button
                  onClick={() => pagination.onPageChange(pagination.page - 1)}
                  disabled={pagination.page === 1}
                  className="flex items-center justify-center h-full py-1.5 px-3 ml-0 text-gray-500 bg-white rounded-l-lg border border-gray-300 hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <svg
                    className="w-5 h-5"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
                      clipRule="evenodd"
                    />
                  </svg>
                </button>
              </li>
              {getPaginationRange(
                Math.ceil(pagination.total / pagination.perPage),
                pagination.page
              ).map((page, idx) =>
                page === null ? (
                  <li
                    key={`ellipsis-${idx}`}
                    className="flex items-center justify-center text-gray-400 px-2 select-none"
                  >
                    ...
                  </li>
                ) : (
                  <li key={page}>
                    <button
                      onClick={() => pagination.onPageChange(page)}
                      className={`flex items-center justify-center text-sm py-2 px-3 border ${
                        pagination.page === page
                          ? "bg-gray-200 text-black font-bold"
                          : "bg-white text-gray-600 hover:bg-gray-100"
                      }`}
                    >
                      {page}
                    </button>
                  </li>
                )
              )}
              <li>
                <button
                  onClick={() => pagination.onPageChange(pagination.page + 1)}
                  disabled={
                    pagination.page ===
                    Math.ceil(pagination.total / pagination.perPage)
                  }
                  className="flex items-center justify-center h-full py-1.5 px-3 text-gray-500 bg-white rounded-r-lg border border-gray-300 hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <svg
                    className="w-5 h-5"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
                      clipRule="evenodd"
                    />
                  </svg>
                </button>
              </li>
            </ul>
          </nav>
        )}
      </div>
    </div>
  );
};

export default DataTable;
