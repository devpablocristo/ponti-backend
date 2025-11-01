import { JSX, useEffect, useState, useMemo } from "react";
import {
  LoaderCircle,
  ClockIcon,
  CheckIcon,
  FileTextIcon,
  FileXIcon,
} from "lucide-react";

import useTasks from "../../../hooks/useTasks";
import DataTable from "../../../components/Table/DataTable";
import { InvoiceData, Metrics, TaskData } from "../../../hooks/useTasks/types";
import FilterBar from "../../../layout/FilterBar/FilterBar";
import { IndicatorCard } from "../../../components/Card/IndicatorCard";
import { useWorkspaceFilters } from "../../../hooks/useWorkspaceFilters";
import { BaseModal } from "../../../components/Modal/BaseModal";
import Button from "../../../components/Button/Button";
import InputField from "../../../components/Input/InputField";
import SelectField from "../../../components/Input/SelectField";
import { cropColors, laborColors } from "../../../pages/admin/colors";
import { Column } from "../../../pages/admin/types";
import APIClient from "../../../restclient/apiInstance";
import { formatNumberAr } from "../utils";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const statusConfig: Record<string, { classes: string; icon: JSX.Element }> = {
  Pendiente: {
    classes: "bg-red-100 text-red-700",
    icon: <ClockIcon className="w-4 h-4" />,
  },
  Pagada: {
    classes: "bg-green-100 text-green-700",
    icon: <CheckIcon className="w-4 h-4" />,
  },
  Facturada: {
    classes: "bg-blue-100 text-blue-700",
    icon: <FileTextIcon className="w-4 h-4" />,
  },
  NoFacturada: {
    classes: "bg-gray-100 text-gray-700",
    icon: <FileXIcon className="w-4 h-4" />,
  },
};

const emptyStatus = "NoFacturada";

const statusOptions = [
  { id: 1, name: "Pendiente" },
  { id: 2, name: "Pagada" },
  { id: 3, name: "Facturada" },
];

function TaskHeader({
  taskAmount,
  selectedColumns,
  setSelectedColumns,
  setVisibleColumns,
  allColumns,
}: {
  taskAmount: number;
  selectedColumns: string[];
  setSelectedColumns: (columns: string[]) => void;
  setVisibleColumns: (columns: string[]) => void;
  allColumns: any[];
}) {
  const [showColumnsModal, setShowColumnsModal] = useState(false);

  return (
    <div className="flex justify-between items-center p-4 bg-white rounded-t-xl border-b border-gray-100">
      <div className="text-sm text-gray-900">
        <span className="font-semibold">Tareas:</span> {taskAmount}
      </div>
      <Button
        variant="light"
        size="sm"
        iconLeft={
          <svg
            className="w-4 h-4 mr-2"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M3 4a1 1 0 011-1h2a1 1 0 011 1v16a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM17 4a1 1 0 011-1h2a1 1 0 011 1v16a1 1 0 01-1 1h-2a1 1 0 01-1-1V4zM10 10h4M10 14h4"
            />
          </svg>
        }
        onClick={() => setShowColumnsModal(true)}
      >
        Configurar columnas
      </Button>
      <BaseModal
        isOpen={showColumnsModal}
        onClose={() => setShowColumnsModal(false)}
        title=""
        primaryButtonText="Aplicar"
        primaryButtonColor="bg-blue-600 hover:bg-blue-800 focus:ring-blue-300 dark:focus:ring-blue-800"
        onPrimaryAction={() => {
          setVisibleColumns(selectedColumns);
          setShowColumnsModal(false);
        }}
        secondaryButtonText="Cancelar"
        onSecondaryAction={() => setShowColumnsModal(false)}
      >
        <h3 className="text-lg font-semibold mb-4">Columnas</h3>
        <div className="grid grid-cols-2 md:grid-cols-3 gap-2 max-h-72 overflow-y-auto px-2 mt-4">
          {allColumns.map((col) => (
            <label
              key={col.key}
              className="flex items-center text-sm font-medium text-gray-700 gap-2"
            >
              <input
                type="checkbox"
                checked={selectedColumns.includes(col.key)}
                onChange={(e) => {
                  if (e.target.checked) {
                    setSelectedColumns([...selectedColumns, col.key]);
                  } else {
                    setSelectedColumns(
                      selectedColumns.filter((k) => k !== col.key)
                    );
                  }
                }}
                className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
              />
              {col.header}
            </label>
          ))}
        </div>
      </BaseModal>
    </div>
  );
}

function TasksIndicators({
  metrics,
  processing,
}: {
  metrics: Metrics;
  processing: boolean;
}) {
  return (
    <div className="flex flex-col md:flex-row gap-4">
      <IndicatorCard
        title="Superficie total"
        value={
          processing
            ? "Cargando..."
            : formatNumberAr(metrics.surface_ha) + " Has"
        }
        color="gray"
        height="85px"
        width="200px"
      />
      <IndicatorCard
        title="Costo promedio / Ha"
        value={
          processing
            ? "Cargando..."
            : " u$" +
              formatNumberAr(metrics.avg_cost_per_ha)
        }
        color="gray"
        height="85px"
        width="200px"
      />
      <IndicatorCard
        title="Total u$ / Neto"
        value={
          processing
            ? "Cargando..."
            : " u$" +
              formatNumberAr(metrics.net_total_cost)
        }
        color="gray"
        height="85px"
        width="200px"
      />
    </div>
  );
}

export function Tasks() {
  const {
    getTasks,
    tasks,
    metrics,
    getMetrics,
    processingMetrics,
    processing,
    error,
    errorMetrics,
    pageInfo,
    updateInvoice,
    createInvoice,
    processingInvoice,
    errorInvoice,
    resultInvoice,
  } = useTasks();

  const [currentPage, setCurrentPage] = useState(1);
  const [taskFilters, setTaskFilters] = useState<Record<string, any>>({});
  const [invoice, setInvoice] = useState<InvoiceData>({
    workorder_id: 0,
    invoice_id: 0,
    invoice_number: "",
    invoice_company: "",
    invoice_date: "",
    invoice_status: "",
  });
  const [showInvoiceModal, setShowInvoiceModal] = useState(false);
  const [resultInvoiceMessage, setResultInvoiceMessage] = useState<
    string | null
  >(null);
  const itemsPerPage = 10;
  const [errorInvoiceMessage, setErrorInvoiceMessage] = useState<string | null>(
    null
  );

  const { filters, projectId, selectedField } = useWorkspaceFilters([
    "customer",
    "project",
    "campaign",
    "field",
  ]);

  useEffect(() => {
    if (!projectId) return;

    let query = "";
    if (selectedField) {
      query += `?field_id=${selectedField.id}`;
    }
    setVisibleColumns(columns.map((col) => col.key));
    setCurrentPage(1);
    getTasks(projectId, query);
    getMetrics(projectId, query);
  }, [projectId, selectedField]);

  useEffect(() => {
    if (resultInvoice && projectId) {
      setResultInvoiceMessage(resultInvoice);
      let query = "";
      if (selectedField) {
        query += `?field_id=${selectedField.id}`;
      }
      setCurrentPage(1);
      getTasks(projectId, query);
      getMetrics(projectId, query);
    }
  }, [resultInvoice, projectId, selectedField]);

  useEffect(() => {
    if (errorInvoice) {
      setResultInvoiceMessage(null);
      setErrorInvoiceMessage(errorInvoice);
    }
  }, [errorInvoice]);

  const columns: Column<TaskData>[] = useMemo(
    () => [
      {
        key: "workorder_number",
        header: "OT N°",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(tasks.map((task) => task.workorder_number)),
        ].sort((a, b) => Number(a) - Number(b)),
        render: (value) => <strong className="text-gray-900">{value}</strong>,
      },
      {
        key: "date",
        header: "Fecha",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(
            tasks.map((task) => {
              const datePart = task.date.split("T")[0];
              const [year, month, day] = datePart.split("-").map(Number);
              const dayStr = String(day).padStart(2, "0");
              const monthStr = String(month).padStart(2, "0");
              return `${dayStr}/${monthStr}/${year}`;
            })
          ),
        ],
        render: (dateString) => {
          if (!dateString) return "";
          const datePart = dateString.split("T")[0];
          const [year, month, day] = datePart.split("-").map(Number);
          const dayStr = String(day).padStart(2, "0");
          const monthStr = String(month).padStart(2, "0");
          return `${dayStr}/${monthStr}/${year}`;
        },
      },
      {
        key: "field_name",
        header: "Campo",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(tasks.map((task) => task.field_name)),
        ].sort(),
      },
      {
        key: "crop_name",
        header: "Cultivo",
        filterable: true,
        filterType: "select",
        filterOptions: [...new Set(tasks.map((task) => task.crop_name))].sort(),
        render: (crop) => (
          <span
            className={`px-2 py-1 text-[14px] rounded-md ${
              cropColors[crop] || "bg-[#E5E7EB] text-[#000000] border border-[#000000]"
            }`}
          >
            {crop}
          </span>
        ),
      },
      {
        key: "contractor",
        header: "Contratista",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(tasks.map((task) => task.contractor)),
        ].sort(),
      },
      {
        key: "category_name",
        header: "Labor",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(tasks.map((task) => task.category_name)),
        ].sort(),
        render: (crop) => (
          <span
            className={`px-2 py-1 text-[14px] rounded-md ${
              laborColors[crop] || "bg-green-200 text-green-800"
            }`}
          >
            {crop}
          </span>
        ),
      },
      {
        key: "surface_ha",
        header: "Superficie",
        filterable: false,
        render: (value) => (
          <strong className="text-gray-900">{formatNumberAr(value)} Has</strong>
        ),
      },
      {
        key: "cost_ha",
        header: "Costo $/Ha",
        filterable: false,
        render: (value) => <strong className="text-gray-700">${formatNumberAr(value)}</strong>,
      },
      {
        key: "net_total",
        header: "Total $ Neto",
        filterable: false,
        render: (value) => (
          <strong className="text-gray-700">${formatNumberAr(value)}</strong>
        ),
      },
      {
        key: "total_iva",
        header: "Total $ IVA",
        filterable: false,
        render: (value) => (
          <strong className="text-gray-700">${formatNumberAr(value)}</strong>
        ),
      },
      {
        key: "investor_name",
        header: "Inversor",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(tasks.map((task) => task.investor_name)),
        ].sort(),
      },
      {
        key: "usd_avg_value",
        header: "u$ Prom",
        filterable: false,
        render: (value) => (
          <>u${formatNumberAr(value)}</>
        ),
      },
      {
        key: "usd_cost_ha",
        header: "Costo U$/Ha",
        filterable: false,
        render: (value) => (
          <>u${formatNumberAr(value)}</>
        ),
      },
      {
        key: "usd_net_total",
        header: "Total u$ Neto",
        filterable: false,
        render: (value) => (
          <>u${formatNumberAr(value)}</>
        ),
      },
      {
        key: "invoice_number",
        header: "N° Factura",
        filterable: false,
        render: (value) => (
          <input
            type="text"
            className="block w-full min-w-[80px] py-1 px-2 text-gray-900 border border-gray-300 rounded-lg bg-gray-50 text-sm disabled:opacity-50"
            value={value}
            disabled={true}
          />
        ),
      },
      {
        key: "invoice_company",
        header: "Empresa",
        filterable: false,
        render: (value) => (
          <input
            type="text"
            className="block w-full min-w-[80px] py-1 px-2 text-gray-900 border border-gray-300 rounded-lg bg-gray-50 text-sm disabled:opacity-50"
            value={value}
            disabled={true}
          />
        ),
      },
      {
        key: "invoice_date",
        header: "Fecha",
        filterable: false,
        render: (dateString) => {
          if (
            !dateString ||
            dateString === "0001-01-01T00:00:00Z" ||
            dateString.startsWith("0001-01-01")
          ) {
            return (
              <input
                type="text"
                className="block w-full min-w-[80px] py-1 px-2 text-gray-900 border border-gray-300 rounded-lg bg-gray-50 text-sm disabled:opacity-50"
                value=""
                disabled={true}
              />
            );
          }
          const datePart = dateString.split("T")[0];
          const [year, month, day] = datePart.split("-").map(Number);
          const dayStr = String(day).padStart(2, "0");
          const monthStr = String(month).padStart(2, "0");
          return `${dayStr}/${monthStr}/${year}`;
        },
      },
      {
        key: "invoice_status",
        header: "Estado",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(
            tasks.map((task) => {
              if (!task.invoice_status) {
                return emptyStatus;
              }
              return task.invoice_status;
            })
          ),
        ],
        render: (status) => {
          if (!status) {
            status = emptyStatus;
          }

          const config = statusConfig[status] || {
            classes: "bg-gray-100 text-gray-700",
            icon: null,
          };

          return (
            <span
              className={`inline-flex items-center gap-1 px-2 py-1 text-[14px] rounded-xl ${config.classes}`}
            >
              {config.icon}
              {status}
            </span>
          );
        },
      },
    ],
    [tasks]
  );

  const allColumnsMap = new Map();
  [...columns].forEach((col) => {
    allColumnsMap.set(col.key, col);
  });
  const allColumns = Array.from(allColumnsMap.values());

  const [columnsToShow, setColumnsToShow] = useState(columns);
  const [selectedColumns, setSelectedColumns] = useState(
    allColumns.map((col) => col.key)
  );
  const [visibleColumns, setVisibleColumns] = useState(selectedColumns);

  const filteredTasks = useMemo(() => {
    return tasks.filter((task) => {
      return Object.entries(taskFilters).every(([key, value]) => {
        if (!value) return true;

        if (key === "invoice_status") {
          if (value === emptyStatus) {
            return (
              !task.invoice_status ||
              task.invoice_status === "" ||
              task.invoice_status === value
            );
          }
          return task.invoice_status === value;
        }

        return (
          String(task[key as keyof TaskData]).toLowerCase() ===
          String(value).toLowerCase()
        );
      });
    });
  }, [tasks, taskFilters]);

  const paginatedTasks = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return filteredTasks.slice(startIndex, startIndex + itemsPerPage);
  }, [filteredTasks, currentPage, itemsPerPage]);

  useEffect(() => {
    setColumnsToShow(columns);
  }, [columns]);

  useEffect(() => {
    setColumnsToShow(
      allColumns.filter((col) => visibleColumns.includes(col.key))
    );
  }, [visibleColumns]);

  const handlePageChange = (newPage: number) => {
    setCurrentPage(newPage);
  };

  const isValidDate = (dateString: string) => {
    if (!dateString) return false;
    if (/^(00|0000)[\/\-](00|00)[\/\-](0000|00)$/.test(dateString))
      return false;
    const date = new Date(dateString);
    return date instanceof Date && !isNaN(date.getTime());
  };

  const handleExport = async () => {
    if (!projectId) return;

    try {
      const response = await request.get<Blob>(
        `/labors/export/${projectId}`,
        undefined,
        { responseType: "blob" }
      );

      const url = window.URL.createObjectURL(response);

      const link = document.createElement("a");
      link.href = url;
      link.download = `labores_${projectId}_${new Date().toISOString()}.xlsx`;
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
            label: "Exportar labores",
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
      <div className="my-4">
        {errorMetrics ? (
          <div className="text-red-500">{errorMetrics}</div>
        ) : (
          <TasksIndicators metrics={metrics} processing={processingMetrics} />
        )}
      </div>

      <div className="mt-4 relative">
        {processing && (
          <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
            <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
          </div>
        )}
        <DataTable
          key={tasks.length}
          data={paginatedTasks}
          columns={columnsToShow}
          filters={taskFilters}
          onFilterChange={setTaskFilters}
          className={`${processing ? "pointer-events-none opacity-60" : ""}`}
          enableFilters={true}
          message="No hay tareas disponibles"
          onEdit={(item) => {
            setResultInvoiceMessage(null);
            setErrorInvoiceMessage(null);

            if (item.invoice_id === 0) {
              setInvoice({
                workorder_id: item.workorder_id,
                invoice_id: 0,
                invoice_number: "",
                invoice_company: "",
                invoice_date: "",
                invoice_status: "",
              });
              setShowInvoiceModal(true);
              return;
            }

            const statusOption = statusOptions.find(
              (opt) => opt.name === item.invoice_status
            );

            setInvoice({
              workorder_id: item.workorder_id,
              invoice_id: item.invoice_id,
              invoice_number: item.invoice_number,
              invoice_company: item.invoice_company,
              invoice_date: item.invoice_date
                ? item.invoice_date.split("T")[0]
                : "",
              invoice_status: statusOption ? statusOption.id.toString() : "",
            });
            setShowInvoiceModal(true);
          }}
          headerComponent={
            <TaskHeader
              taskAmount={tasks.length}
              selectedColumns={selectedColumns}
              setSelectedColumns={setSelectedColumns}
              setVisibleColumns={setVisibleColumns}
              allColumns={allColumns}
            />
          }
          pagination={
            pageInfo
              ? {
                  page: currentPage,
                  perPage: itemsPerPage,
                  total: filteredTasks.length,
                  onPageChange: handlePageChange,
                }
              : undefined
          }
        />
        <BaseModal
          isOpen={showInvoiceModal}
          onClose={() => {
            setShowInvoiceModal(false);
          }}
          title="Cargar Factura"
          onPrimaryAction={() => {
            if (processingInvoice) return;

            const statusText = statusOptions.find(
              (opt) => opt.id.toString() === invoice.invoice_status
            )?.name;
            if (!statusText) return;

            setResultInvoiceMessage(null);
            setErrorInvoiceMessage(null);

            if (!isValidDate(invoice.invoice_date)) {
              setErrorInvoiceMessage("La fecha ingresada no es válida.");
              return;
            }

            const invoiceData = {
              ...invoice,
              invoice_status: statusText,
            };

            if (invoiceData.invoice_id > 0) {
              updateInvoice(invoiceData.invoice_id, invoiceData);
            } else {
              createInvoice(invoiceData);
            }
          }}
          onSecondaryAction={() => {
            if (processingInvoice) return;
            setShowInvoiceModal(false);
          }}
          primaryButtonText="Cargar"
          secondaryButtonText="Cancelar"
        >
          <div className="flex gap-2 mt-2">
            {processingInvoice && (
              <LoaderCircle className="w-5 h-5 text-blue-600 animate-spin" />
            )}
            <InputField
              label="Ingrese N° Factura"
              placeholder="N°"
              name="invoiceNumber"
              value={invoice.invoice_number}
              disabled={processingInvoice}
              onChange={(e) =>
                setInvoice({ ...invoice, invoice_number: e.target.value })
              }
            />
            <InputField
              label="Fecha"
              type="date"
              placeholder="Fecha"
              name="invoiceDate"
              disabled={processingInvoice}
              value={invoice.invoice_date}
              onChange={(e) =>
                setInvoice({ ...invoice, invoice_date: e.target.value })
              }
            />
          </div>
          <div className="flex gap-2 mt-2">
            <InputField
              label="Ingrese Nombre Empresa"
              placeholder="Empresa"
              name="companyName"
              disabled={processingInvoice}
              value={invoice.invoice_company}
              onChange={(e) =>
                setInvoice({ ...invoice, invoice_company: e.target.value })
              }
            />
            <SelectField
              label="Estado"
              placeholder="Estado"
              name="status"
              disabled={processingInvoice}
              value={invoice.invoice_status}
              options={statusOptions}
              onChange={(e) => {
                setInvoice({
                  ...invoice,
                  invoice_status: e.target.value,
                });
              }}
            />
          </div>
          {errorInvoiceMessage && (
            <div
              className="p-4 mt-4 text-sm text-red-800 rounded-lg bg-red-50"
              role="alert"
            >
              <span className="font-medium">Error!</span> {errorInvoiceMessage}
            </div>
          )}
          {resultInvoiceMessage && (
            <div
              className="p-4 mt-4 text-sm text-green-800 rounded-lg bg-green-50"
              role="alert"
            >
              <span className="font-medium">Exito!</span> {resultInvoiceMessage}
            </div>
          )}
        </BaseModal>
        {error && (
          <div
            className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50"
            role="alert"
          >
            <span className="font-medium">Error!</span> {error}
          </div>
        )}
      </div>
    </div>
  );
}
