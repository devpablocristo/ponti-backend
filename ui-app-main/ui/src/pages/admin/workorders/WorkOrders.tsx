import React, { useEffect, useMemo, useState } from "react";
import { LoaderCircle } from "lucide-react";

import DataTable from "../../../components/Table/DataTable";
import {
  Metrics,
  OrdersData,
  WorkorderData,
} from "../../../hooks/useWorkOrders/types";
import useOrders from "../../../hooks/useWorkOrders";
import FilterBar from "../../../layout/FilterBar/FilterBar";
import { IndicatorCard } from "../../../components/Card/IndicatorCard";
import CreateOrder from "./CreateOrder";
import { useWorkspaceFilters } from "../../../hooks/useWorkspaceFilters";
import { BaseModal } from "../../../components/Modal/BaseModal";
import Button from "../../../components/Button/Button";
import UpdateOrder from "./UpdateOrder";
import { cropColors, laborColors } from "../../../pages/admin/colors";
import APIClient from "../../../restclient/apiInstance";
import { formatNumberAr } from "../utils";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

function OrdersHeader({
  ordersAmount,
  laborAmount,
  selectedColumns,
  setSelectedColumns,
  setVisibleColumns,
  allColumns,
}: {
  ordersAmount: number;
  laborAmount: number;
  selectedColumns: string[];
  setSelectedColumns: (columns: string[]) => void;
  setVisibleColumns: (columns: string[]) => void;
  allColumns: any[];
}) {
  const [showColumnsModal, setShowColumnsModal] = useState(false);

  return (
    <div className="flex justify-between items-center p-4 bg-white rounded-t-xl border-b border-gray-100">
      <div className="text-sm text-gray-900">
        <span className="font-semibold">Órdenes:</span> {ordersAmount}{" "}
        <span className="font-semibold ml-4">Labores:</span> {laborAmount}
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

function OrdersIndicators({
  metrics,
  processing,
}: {
  metrics: Metrics;
  processing: boolean;
}) {
  return (
    <div className="flex flex-col md:flex-row gap-4">
      <IndicatorCard
        title="Superficie ejecutada"
        value={processing ? "Cargando..." : formatNumberAr(metrics.surface_ha) + " Has"}
        color="gray"
        height="85px"
        width="200px"
      />
      <IndicatorCard
        title="Consumo en litros"
        value={processing ? "Cargando..." : formatNumberAr(metrics.liters) + " Lts"}
        color="gray"
        height="85px"
        width="200px"
      />
      <IndicatorCard
        title="Consumo en kilos"
        value={processing ? "Cargando..." : formatNumberAr(metrics.kilograms) + " Kg"}
        color="gray"
        height="85px"
        width="200px"
      />
      <IndicatorCard
        title="Costos directos"
        value={processing ? "Cargando..." : " u$" + formatNumberAr(metrics.direct_cost)}
        color="gray"
        height="85px"
        width="200px"
      />
    </div>
  );
}

export function WorkOrders() {
  const [selectedOrderId, setSelectedOrderId] = useState<number | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [drawerUpdateOpen, setDrawerUpdateOpen] = useState(false);
  const [orderToDuplicate, setOrderToDuplicate] =
    useState<WorkorderData | null>(null);

  const {
    getOrders,
    deleteOrder,
    getMetrics,
    metrics,
    processingMetrics,
    errorMetrics,
    orders,
    processing,
    error,
  } = useOrders();

  const columns: Column<OrdersData>[] = React.useMemo(
    () => [
      {
        key: "number",
        header: "N°",
        filterable: true,
        filterType: "select",
        filterOptions: [...new Set(orders.map((order) => order.number))].sort(
          (a, b) => Number(a) - Number(b)
        ),
        render: (value, data) => (
          <strong className="text-blue-700">
            <a
              onClick={() => {
                setSelectedOrderId(data.id);
                setDrawerUpdateOpen(true);
              }}
            >
              {value as string}
            </a>
          </strong>
        ),
      },
      {
        key: "project_name",
        header: "Proyecto",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(orders.map((order) => order.project_name)),
        ].sort(),
      },
      {
        key: "field_name",
        header: "Campo",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(orders.map((order) => order.field_name)),
        ].sort(),
      },
      {
        key: "lot_name",
        header: "Lote",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(orders.map((order) => order.lot_name)),
        ].sort(),
      },
      {
        key: "date",
        header: "Fecha",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(
            orders.map((order) => {
              if (!order.date) return "";
              const datePart = order.date.split("T")[0];
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
        key: "crop_name",
        header: "Cultivo",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(orders.map((order) => order.crop_name)),
        ].sort(),
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
        key: "labor_category_name",
        header: "Labor",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(orders.map((order) => order.labor_category_name)),
        ].sort(),
        render: (labor) => (
          <span
            className={`px-2 py-1 text-[14px] rounded-md ${
              laborColors[labor] || "bg-[#E5E7EB] text-[#000000] border border-[#000000]"
            }`}
          >
            {labor}
          </span>
        ),
      },
      {
        key: "type_name",
        header: "Tipo/Clase",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(orders.map((order) => order.type_name)),
        ].sort(),
      },
      {
        key: "contractor",
        header: "Contratista",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(orders.map((order) => order.contractor)),
        ].sort(),
      },
      {
        key: "surface_ha",
        header: "Superficie",
        filterable: false,
      },
      {
        key: "supply_name",
        header: "Insumo",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(orders.map((order) => order.supply_name)),
        ].sort(),
      },
      {
        key: "consumption",
        header: "Consumo",
        filterable: false,
      },
      {
        key: "category_name",
        header: "Rubro",
        filterable: true,
        filterType: "select",
        filterOptions: [
          ...new Set(orders.map((order) => order.category_name)),
        ].sort(),
      },
      { key: "dose", header: "Dosis", filterable: false, render: (value: any) => {
          return <strong>{value}</strong>;
        }},
      {
        key: "cost_per_ha",
        header: "Costo $/HA",
        filterable: false,
        render: (value: any) => {
          const num = Number(value);
          return <>{isNaN(num) ? "-" : `u$${formatNumberAr(num)}`}</>;
        },
      },
      { key: "unit_price", header: "Precio unidad", filterable: false },
      {
        key: "total_cost",
        header: "Total costo",
        filterable: false,
        render: (value: any) => {
          const num = Number(value);
          return <>{isNaN(num) ? "-" : `u$${formatNumberAr(num)}`}</>;
        },
      },
    ],
    [orders]
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

  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  const { projectId, selectedProject, selectedField, filters } =
    useWorkspaceFilters(["customer", "project", "campaign", "field"]);

  const [errorMessage, setErrorMessage] = useState("");

  const [isProcessing, setIsProcessing] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalConfig, setModalConfig] = useState({
    title: "",
    message: "",
    primaryButtonText: "",
    secondaryButtonText: "Cancelar",
    onConfirm: () => {},
  });

  const [columnsFilters, setColumnsFilters] = useState<Record<string, any>>({});

  useEffect(() => {
    if (!error) return;
    setErrorMessage(error);
  }, [error]);

  useEffect(() => {
    setColumnsToShow(columns);
  }, [columns]);

  useEffect(() => {
    if (!projectId && !selectedField) {
      setErrorMessage("Seleccione un proyecto o un campo para ver las ordenes");
      return;
    }
    setErrorMessage("");
    setVisibleColumns(columns.map((col) => col.key));

    let query = "";
    if (selectedField && selectedField.id !== 0) {
      query += `field_id=${selectedField.id}`;
    } else {
      query += `project_id=${projectId}`;
    }

    setCurrentPage(1);
    getOrders(query);
    getMetrics(query);
  }, [projectId, selectedField]);

  const handleOrderCreated = () => {
    let query = "";
    if (selectedField && selectedField.id !== 0) {
      query += `field_id=${selectedField.id}`;
    } else {
      query += `project_id=${projectId}`;
    }
    setCurrentPage(1);
    getOrders(query);
    getMetrics(query);
  };

  const handleOrderDuplicated = (order: WorkorderData) => {
    setSelectedOrderId(null);
    setDrawerUpdateOpen(false);
    setDrawerOpen(true);
    setOrderToDuplicate(order);
  };

  const handlePreFinish = (id: number) => {
    setModalConfig({
      title: "Confirmar eliminación",
      message: "¿Está seguro que desea eliminar la orden?",
      primaryButtonText: "Sí, eliminar",
      secondaryButtonText: "Cancelar",
      onConfirm: () => handleFinishConfirmed(id),
    });
    setIsModalOpen(true);
  };

  const handleFinishConfirmed = async (id: number) => {
    setIsProcessing(true);

    try {
      await deleteOrder(id);
    } catch (error) {
      setErrorMessage("Error al eliminar la orden.");
    } finally {
      setModalConfig({
        title: "Confirmación",
        message: "La orden ha sido eliminada.",
        primaryButtonText: "Volver",
        secondaryButtonText: "Volver",
        onConfirm: () => {
          window.location.href = "/admin/work-orders";
        },
      });
      setIsModalOpen(true);
      setIsProcessing(false);
    }
  };

  useEffect(() => {
    setColumnsToShow(
      allColumns.filter((col) => visibleColumns.includes(col.key))
    );
  }, [visibleColumns]);

  const handlePageChange = (newPage: number) => {
    setCurrentPage(newPage);
  };

  const filteredOrders = useMemo(() => {
    return orders.filter((order) => {
      return Object.entries(columnsFilters).every(([key, value]) => {
        if (!value) return true;
        return (
          String(order[key as keyof OrdersData]).toLowerCase() ===
          String(value).toLowerCase()
        );
      });
    });
  }, [orders, columnsFilters]);

  const paginatedOrders = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return filteredOrders.slice(startIndex, startIndex + itemsPerPage);
  }, [filteredOrders, currentPage, itemsPerPage]);

  const handleExport = async () => {
    if (!projectId) return;

    try {
      const response = await request.get<Blob>(
        `/workorders/export/${projectId}`,
        undefined,
        { responseType: "blob" }
      );

      const url = window.URL.createObjectURL(response);

      const link = document.createElement("a");
      link.href = url;
      link.download = `ordenes_${projectId}_${new Date().toISOString()}.xlsx`;
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
            label: "Exportar órdenes",
            icon: <svg width="14" height="13" viewBox="0 0 14 13" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M5.66675 2.49984H3.00008C2.64646 2.49984 2.30732 2.64031 2.05727 2.89036C1.80722 3.14041 1.66675 3.47955 1.66675 3.83317V10.4998C1.66675 10.8535 1.80722 11.1926 2.05727 11.4426C2.30732 11.6927 2.64646 11.8332 3.00008 11.8332H9.66675C10.0204 11.8332 10.3595 11.6927 10.6096 11.4426C10.8596 11.1926 11.0001 10.8535 11.0001 10.4998V7.83317M8.33341 1.1665H12.3334M12.3334 1.1665V5.1665M12.3334 1.1665L5.66675 7.83317" stroke="#547792" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            ,
            variant: "outlinePonti",
            isPrimary: true,
            disabled: !projectId,
            onClick: () => handleExport(),
          },
          {
            label: "+ Nueva orden",
            variant: "success",
            isPrimary: true,
            disabled: !projectId,
            onClick: () => {
              setDrawerOpen(true);
              setOrderToDuplicate(null);
            },
          },
        ]}
      />
      {errorMessage && (
        <div
          className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
          role="alert"
        >
          <span className="font-medium">Error!</span> {errorMessage}
        </div>
      )}
      {!processing && !errorMetrics && orders.length > 0 && (
        <div className="my-4">
          <OrdersIndicators metrics={metrics} processing={processingMetrics} />
        </div>
      )}
      <div className="mt-4 relative">
        {processing ||
          (isProcessing && (
            <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
              <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
            </div>
          ))}
        {selectedProject && (
          <CreateOrder
            drawerOpen={drawerOpen}
            setDrawerOpen={setDrawerOpen}
            projectId={projectId}
            orderToDuplicate={orderToDuplicate}
            selectedField={selectedField}
            onOrderCreated={handleOrderCreated}
          />
        )}
        {selectedOrderId && (
          <UpdateOrder
            orderId={selectedOrderId}
            drawerOpen={drawerUpdateOpen}
            setDrawerOpen={setDrawerUpdateOpen}
            onOrderUpdated={handleOrderCreated}
            onOrderDuplicated={handleOrderDuplicated}
          />
        )}
        <DataTable
          key={orders.length}
          data={paginatedOrders}
          filters={columnsFilters}
          onFilterChange={setColumnsFilters}
          columns={columnsToShow}
          onDelete={(item) => handlePreFinish(item.id)}
          enableFilters={true}
          headerComponent={
            <OrdersHeader
              ordersAmount={orders.length}
              laborAmount={orders.length}
              selectedColumns={selectedColumns}
              setSelectedColumns={setSelectedColumns}
              setVisibleColumns={setVisibleColumns}
              allColumns={allColumns}
            />
          }
          message="No hay ordenes disponibles"
          pagination={{
            page: currentPage,
            perPage: itemsPerPage,
            total: filteredOrders.length,
            onPageChange: handlePageChange,
          }}
        />
        <BaseModal
          isOpen={isModalOpen}
          isSaving={isProcessing}
          onClose={() => setIsModalOpen(false)}
          title={modalConfig.title}
          message={modalConfig.message}
          primaryButtonText={modalConfig.primaryButtonText}
          secondaryButtonText={modalConfig.secondaryButtonText}
          onPrimaryAction={() => {
            modalConfig.onConfirm();
            setIsModalOpen(false);
          }}
        >
          <div className="flex flex-col items-center gap-2">
            <p>{modalConfig.message}</p>
          </div>
        </BaseModal>
      </div>
    </div>
  );
}

type Column<T> = {
  key: keyof T;
  header: string;
  render?: (value: any, item: T) => React.ReactNode;
  filterable?: boolean;
  filterType?: "text" | "number" | "select" | "date";
  filterOptions?: string[];
};
