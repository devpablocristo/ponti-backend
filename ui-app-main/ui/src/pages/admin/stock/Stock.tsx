import { useEffect, useMemo, useState } from "react";
import { LoaderCircle, Pencil, Check, AlertCircle } from "lucide-react";

import DataTable from "../../../components/Table/DataTable";
import useStock from "../../../hooks/useStock";
import FilterBar from "../../../layout/FilterBar/FilterBar";
import { IndicatorCard } from "../../../components/Card/IndicatorCard";
import { useWorkspaceFilters } from "../../../hooks/useWorkspaceFilters";
import { GetStockItems, Summary } from "../../../hooks/useStock/types";
import { BaseModal } from "../../../components/Modal/BaseModal";
import { Column } from "../types";
import SelectField from "../../../components/Input/SelectField";
import APIClient from "../../../restclient/apiInstance";
import { formatNumberAr } from "../utils";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const EditableCell = ({
  item,
  value,
  projectId,
}: {
  item: any;
  value: string | number;
  projectId: number | null;
}) => {
  const [editing, setEditing] = useState(false);
  const [editValue, setEditValue] = useState(value ?? "");
  const { updateStock, processingStock, errorStock, resultStock } = useStock();

  useEffect(() => {
    setEditValue(value ?? "");
  }, [value, item.id]);

  const save = async () => {
    if (editValue === "" || editValue === "0") {
      return;
    }
    if (projectId === null) {
      alert("Error al guardar");
      return;
    }
    updateStock(projectId, item.id, Number(editValue));
  };

  useEffect(() => {
    if (errorStock) {
      alert(errorStock);
      return;
    }
    if (resultStock) {
      setEditing(false);
      window.location.reload();
      return;
    }
  }, [errorStock, resultStock]);

  if (editing) {
    return (
      <div className="flex items-center gap-2">
        <input
          type="number"
          min="0"
          className="block w-full min-w-[80px] p-2 text-gray-900 border border-gray-300 rounded-lg bg-gray-50 text-sm focus:ring-blue-500 focus:border-blue-500"
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          disabled={processingStock}
          onBlur={save}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              save();
            }
            if (e.key === "Escape") {
              setEditing(false);
            }
          }}
        />
        {processingStock ? (
          <LoaderCircle className="animate-spin w-4 h-4 text-blue-500" />
        ) : (
          <button
            className="text-green-600 hover:text-green-800"
            onClick={save}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="flex items-center justify-between w-full min-w-[80px]">
      <input
        type="number"
        min="0"
        className="block w-full p-2 text-gray-800 border border-gray-300 rounded-lg bg-gray-100 text-sm"
        value={value}
        onChange={() => {}}
        disabled={true}
      />
      <button
        className="text-blue-600 hover:text-blue-800 flex items-center p-1"
        style={{ minWidth: 24, minHeight: 24 }}
        onClick={() => setEditing(true)}
        aria-label="Editar"
      >
        <Pencil className="h-3.5 w-3.5" />
      </button>
    </div>
  );
};

function CloseStockDate({
  date,
  onDateChange,
  enabledCloseStock,
  setEnabledCloseStock,
  disabledCloseStock,
}: {
  date: string;
  onDateChange: (date: string) => void;
  enabledCloseStock: boolean;
  setEnabledCloseStock: (enabled: boolean) => void;
  disabledCloseStock: boolean;
}) {
  const [internalDate, setInternalDate] = useState(date);

  useEffect(() => {
    setInternalDate(date);
  }, [date]);

  return (
    <div className="flex flex-col items-start gap-1">
      <label className="text-sm font-semibold mb-1">Cerrar stock a fecha</label>
      <div className="flex items-center gap-2">
        <input
          type="date"
          disabled={disabledCloseStock}
          value={internalDate}
          onChange={(e) => setInternalDate(e.target.value)}
          className="pl-10 pr-3 py-2 border border-gray-300 rounded-md bg-white text-gray-900 font-medium focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <input
          type="checkbox"
          checked={enabledCloseStock}
          onChange={() => {
            if (!enabledCloseStock && internalDate) {
              setEnabledCloseStock(true);
              onDateChange(internalDate);
            } else {
              setEnabledCloseStock(false);
            }
          }}
          className="accent-blue-600 w-5 h-5"
          disabled={disabledCloseStock}
        />
        <strong className="text-xs">Cerrar stock</strong>
      </div>
    </div>
  );
}

function ItemsIndicators({
  summary,
  selectedDate,
  onDateChange,
  enabledCloseStock,
  setEnabledCloseStock,
  disabledCloseStock,
}: {
  summary: Summary;
  selectedDate: string;
  onDateChange: (date: string) => void;
  enabledCloseStock: boolean;
  setEnabledCloseStock: (enabled: boolean) => void;
  disabledCloseStock: boolean;
}) {
  return (
    <div className="flex flex-col md:flex-row gap-4">
      <IndicatorCard
        title="Total insumos invertidos Kg"
        value={formatNumberAr(summary.total_kg) + " Kg"}
        color="gray"
        height="85px"
        width="220px"
      />
      <IndicatorCard
        title="Total insumos invertido Lts"
        value={formatNumberAr(summary.total_lt) + " Lts"}
        color="gray"
        height="85px"
        width="220px"
      />
      <IndicatorCard
        title="Total u$ / Neto"
        value={formatNumberAr(summary.total_usd) + " u$"}
        color="gray"
        height="85px"
        width="220px"
      />

      <div className="flex-1 flex justify-end">
        <CloseStockDate
          date={selectedDate}
          onDateChange={onDateChange}
          enabledCloseStock={enabledCloseStock}
          setEnabledCloseStock={setEnabledCloseStock}
          disabledCloseStock={disabledCloseStock}
        />
      </div>
    </div>
  );
}

export function Stock() {
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedDate, setSelectedDate] = useState<string>("");
  const [columnsFilters, setColumnsFilters] = useState<Record<string, any>>({});
  const [disabledCloseStock, setDisabledCloseStock] = useState(false);
  const [enabledCloseStock, setEnabledCloseStock] = useState(false);
  const [stockPeriods, setStockPeriods] = useState<
    {
      id: number;
      name: string;
    }[]
  >([{ id: 0, name: "Activo" }]);

  const [period, setPeriod] = useState("0");

  const { projectId, filters, selectedCustomer, selectedCampaignId } =
    useWorkspaceFilters(["customer", "project", "campaign", "field"]);

  const {
    getStock,
    stock,
    summary,
    processing,
    error,
    closeStock,
    processingCloseStock,
    errorCloseStock,
    resultCloseStock,
    getPeriods,
    errorPeriods,
    periods,
  } = useStock();

  const columns: Column<GetStockItems>[] = useMemo(
    () => [
      {
        key: "supply_name",
        header: "Insumo",
        render: (value) => <strong>{value}</strong>,
        filterable: true,
        filterType: "select",
        filterOptions: [...new Set(stock.map((stock) => stock.supply_name))],
      },
      {
        key: "class_type",
        header: "Rubro",
        filterable: true,
        filterType: "select",
        filterOptions: [...new Set(stock.map((stock) => stock.class_type))],
      },
      {
        key: "investor_name",
        header: "Inversor",
        filterable: true,
        filterType: "select",
        filterOptions: [...new Set(stock.map((stock) => stock.investor_name))],
      },
      {
        key: "entry_stock",
        filterable: false,
        header: "Ingresados",
        render: (value, item) => {
          const unit = item.supply_unit_id === 1 ? "Lt" : "Kg";
          return formatNumberAr(value) + unit;
        },
      },
      {
        key: "consumed",
        filterable: false,
        header: "Consumidos",
        render: (value, item) => {
          const unit = item.supply_unit_id === 1 ? "Lt" : "Kg";
          return formatNumberAr(value) + unit;
        },
      },
      {
        key: "stock_units",
        filterable: false,
        header: "Stock",
        render: (value, item) => {
          const unit = item.supply_unit_id === 1 ? "Lt" : "Kg";
          return formatNumberAr(value) + unit;
        },
      },
      {
        key: "real_stock_units",
        filterable: false,
        header: "Stock real",
        render: (value, item) => (
          <EditableCell item={item} value={value} projectId={projectId} />
        ),
      },
      {
        key: "stock_difference",
        filterable: false,
        header: "Diferencia",
        render: (diff) => {
          const value = Number(diff);
          if (value === 0) {
            return diff;
          }
          if (value > 0) {
            return (
              <span className="inline-flex items-center gap-1 px-2 py-1 text-[14px] rounded-md bg-green-100 text-green-900">
                <Check className="w-4 h-4" />
                {value}
              </span>
            );
          }
          return (
            <span className="inline-flex items-center gap-1 px-2 py-1 text-[14px] rounded-md bg-red-100 text-red-900">
              <AlertCircle className="w-4 h-4" />
              {value}
            </span>
          );
        },
      },
      {
        key: "close_date",
        filterable: false,
        header: "Fecha de cierre",
        render: (dateString) => {
          if (!dateString) return " - ";
          const datePart = dateString.split("T")[0];
          const [year, month, day] = datePart.split("-").map(Number);
          const dayStr = String(day).padStart(2, "0");
          const monthStr = String(month).padStart(2, "0");
          return <strong>{`${dayStr}/${monthStr}/${year}`}</strong>;
        },
      },
      {
        key: "supply_unit_price",
        header: "Precio U.",
        filterable: false,
        render: (value) => "u$ " + formatNumberAr(value),
      },
      {
        key: "total_usd",
        header: "Total u$",
        filterable: false,
        render: (value) => {
          const num = Number(value);
          return (
            <strong>
              {isNaN(num)
                ? "-"
                : `u$ ${num.toLocaleString("es-AR", {
                    minimumFractionDigits: 0,
                    maximumFractionDigits: 3,
                  })}`}
            </strong>
          );
        },
      },
    ],
    [projectId, stock]
  );

  useEffect(() => {
    if (!projectId || !selectedCustomer || !selectedCampaignId) {
      return;
    }

    getStock(projectId, "");
    getPeriods(projectId);
    setDisabledCloseStock(false);
    setSelectedDate("");
  }, [getStock, projectId, selectedCustomer, selectedCampaignId]);

  useEffect(() => {
    if (periods && periods.length > 0) {
      setStockPeriods((prev) => [
        ...prev,
        ...periods
          .filter((p) => !prev.some((item) => item.name === p))
          .map((p, idx) => ({
            id: prev.length + idx,
            name: p,
          })),
      ]);
    }
  }, [periods]);

  useEffect(() => {
    if (!projectId) return;
    const periodNumber = Number(period);
    if (periodNumber === 0) {
      getStock(projectId, "");
      setDisabledCloseStock(false);
      setSelectedDate("");
      return;
    }
    getStock(projectId, stockPeriods[periodNumber]?.name || "");
    setSelectedDate(stockPeriods[periodNumber]?.name || "");
    setDisabledCloseStock(true);
  }, [period, stockPeriods]);

  useEffect(() => {
    if (errorCloseStock) {
      alert(errorCloseStock);
    }
  }, [errorCloseStock]);

  useEffect(() => {
    if (resultCloseStock && projectId) {
      alert(resultCloseStock);
      getStock(projectId, "");
      getPeriods(projectId);
      setEnabledCloseStock(false);
      setDisabledCloseStock(false);
      setSelectedDate("");
    }
  }, [resultCloseStock]);

  const handlePageChange = (newPage: number) => {
    setCurrentPage(newPage);
  };

  const handleDateChange = (date: string) => {
    setSelectedDate(date);
    setIsModalOpen(true);
  };

  const handleCloseStock = () => {
    if (projectId === null) {
      return;
    }
    closeStock(projectId, selectedDate);
    setIsModalOpen(false);
  };

  const filteredStock = useMemo(() => {
    return stock.filter((m) => {
      return Object.entries(columnsFilters).every(([key, value]) => {
        if (!value) return true;
        return String(m[key as keyof GetStockItems])
          .toLowerCase()
          .includes(String(value).toLowerCase());
      });
    });
  }, [stock, columnsFilters]);

  const paginatedStock = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return filteredStock.slice(startIndex, startIndex + itemsPerPage);
  }, [filteredStock, currentPage, itemsPerPage]);

  const handleExport = async () => {
    if (!projectId) return;

    try {
      const response = await request.get<Blob>(
        `/stock/export/${projectId}`,
        undefined,
        { responseType: "blob" }
      );

      const url = window.URL.createObjectURL(response);

      const link = document.createElement("a");
      link.href = url;
      link.download = `stock_${projectId}_${new Date().toISOString()}.xlsx`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <div>
      <FilterBar
        filters={filters}
        actions={[
          {
            label: "Exportar stock",
            icon: <svg width="14" height="13" viewBox="0 0 14 13" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M5.66675 2.49984H3.00008C2.64646 2.49984 2.30732 2.64031 2.05727 2.89036C1.80722 3.14041 1.66675 3.47955 1.66675 3.83317V10.4998C1.66675 10.8535 1.80722 11.1926 2.05727 11.4426C2.30732 11.6927 2.64646 11.8332 3.00008 11.8332H9.66675C10.0204 11.8332 10.3595 11.6927 10.6096 11.4426C10.8596 11.1926 11.0001 10.8535 11.0001 10.4998V7.83317M8.33341 1.1665H12.3334M12.3334 1.1665V5.1665M12.3334 1.1665L5.66675 7.83317" stroke="#547792" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            ,
            variant: "outlinePonti",
            isPrimary: true,
            disabled: !projectId,
            onClick: () => handleExport(),
          },
        ]}
      />
      {!error && projectId && selectedCustomer && selectedCampaignId && (
        <div className="my-4">
          <ItemsIndicators
            summary={summary}
            selectedDate={selectedDate}
            disabledCloseStock={disabledCloseStock}
            onDateChange={handleDateChange}
            enabledCloseStock={enabledCloseStock}
            setEnabledCloseStock={setEnabledCloseStock}
          />
        </div>
      )}
      <div className="mt-4 relative">
        {processing && (
          <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
            <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
          </div>
        )}

        {error && (
          <div
            className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
            role="alert"
          >
            <span className="font-medium">Error!</span> {error}
          </div>
        )}
        {stockPeriods && stockPeriods.length > 0 && (
          <SelectField
            label="Periodo (fecha de cierre)"
            name="period"
            options={stockPeriods}
            className="max-w-56 mb-3"
            value={period}
            size="sm"
            onChange={(e) => setPeriod(e.target.value)}
          />
        )}
        {errorPeriods && <div>{errorPeriods}</div>}
        {projectId && selectedCustomer && selectedCampaignId && (
          <DataTable
            data={paginatedStock}
            columns={columns}
            message="No hay stock disponible"
            filters={columnsFilters}
            onFilterChange={setColumnsFilters}
            enableFilters={true}
            pagination={{
              page: currentPage,
              perPage: itemsPerPage,
              total: filteredStock.length,
              onPageChange: handlePageChange,
            }}
          />
        )}
        <BaseModal
          isOpen={isModalOpen}
          onClose={() => {
            setIsModalOpen(false);
            setEnabledCloseStock(false);
          }}
          isSaving={processingCloseStock}
          title={"Se va a cerrar el stock"}
          message={`¿Está seguro que desea cerrar el stock a la fecha ${selectedDate
            .split("-")
            .reverse()
            .join("/")}?`}
          primaryButtonText={"Sí, cerrar"}
          secondaryButtonText={"Cancelar"}
          onPrimaryAction={() => {
            handleCloseStock();
          }}
          onSecondaryAction={() => {
            setIsModalOpen(false);
            setEnabledCloseStock(false);
          }}
        >
          <div className="flex flex-col items-center gap-2">
            <p>{`¿Está seguro que desea cerrar el stock a la fecha ${selectedDate
              .split("-")
              .reverse()
              .join("/")}?`}</p>
          </div>
        </BaseModal>
      </div>
    </div>
  );
}
