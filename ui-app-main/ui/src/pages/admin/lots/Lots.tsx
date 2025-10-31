import React, { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { LoaderCircle, Pencil } from "lucide-react";

import DataTable from "../../../components/Table/DataTable";
import { BaseModal } from "../../../components/Modal/BaseModal";
import {
  LotKPIs,
  LotsData,
  LotsDataUpdate,
} from "../../../hooks/useLots/types";
import useLots from "../../../hooks/useLots";
import FilterBar from "../../../layout/FilterBar/FilterBar";
import { IndicatorCard } from "../../../components/Card/IndicatorCard";
import { useWorkspaceFilters } from "../../../hooks/useWorkspaceFilters";
import InputField from "../../../components/Input/InputField";
import Button from "../../../components/Button/Button";
import SelectField from "../../../components/Input/SelectField";
import { cropColors } from "../colors";
import APIClient from "../../../restclient/apiInstance";
import { formatNumberAr } from "../utils";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const EditableCell = ({
  item,
  value,
  onSuccessEdit,
}: {
  item: any;
  value: string;
  onSuccessEdit: () => void;
}) => {
  const [editing, setEditing] = React.useState(false);
  const [editValue, setEditValue] = React.useState(value ?? "");
  const { updateTons, processingTons, errorTons, resultTons } = useLots();

  useEffect(() => {
    setEditValue(value ?? "");
  }, [value, item.id]);

  const save = async () => {
    if (editValue === "" || editValue === "0") {
      return;
    }
    updateTons(item.id, Number(editValue));
  };

  useEffect(() => {
    if (errorTons) {
      alert(errorTons);
      return;
    }
    if (resultTons) {
      setEditing(false);
      onSuccessEdit();
      return;
    }
  }, [errorTons, resultTons]);

  if (editing) {
    return (
      <div className="flex items-center gap-2">
        <input
          type="number"
          min="0"
          className="block w-full min-w-[80px] p-2 text-gray-900 border border-gray-300 rounded-lg bg-gray-50 text-sm focus:ring-blue-500 focus:border-blue-500"
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          disabled={processingTons}
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
        {processingTons ? (
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
      <span className="truncate text-right w-full pr-2">{editValue}</span>
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

function LotsIndicators({
  kpis,
  processing,
  error,
}: {
  kpis: LotKPIs;
  processing: boolean;
  error: string | null;
}) {
  return (
    <div className="flex gap-4">
      {processing ? (
        <LoaderCircle className="animate-spin w-4 h-4 text-blue-500" />
      ) : error ? (
        <span className="text-red-500">{error}</span>
      ) : (
        <>
          <IndicatorCard
            title="Superficie sembrada"
            value={formatNumberAr(kpis.seeded_area) + " Has"}
            color="gray"
            height="85px"
            width="200px"
          />
          <IndicatorCard
            title="Superficie cosechada"
            value={formatNumberAr(kpis.harvested_area) + " Has"}
            color="gray"
            height="85px"
            width="200px"
          />
          <IndicatorCard
            title="Toneladas x hectárea"
            value={formatNumberAr(kpis.yield_tn_per_ha) + " Tn"}
            color="green"
            height="85px"
            width="200px"
          />
          <IndicatorCard
            title="Costo x hectárea"
            value={formatNumberAr(kpis.cost_per_hectare) + " u$"}
            color="red"
            height="85px"
            width="200px"
          />
          <IndicatorCard
            title="Superficie total"
            value={formatNumberAr(kpis.superficie_total) + " Has"}
            color="red"
            height="85px"
            width="200px"
          />
        </>
      )}
    </div>
  );
}

function LotsHeader({
  fieldsAmount,
  lotsAmount,
  selectedColumns,
  setSelectedColumns,
  setVisibleColumns,
  columns,
  harvestColumns,
  commercializationColumns,
  allColumns,
}: {
  fieldsAmount: number;
  lotsAmount: number;
  selectedColumns: string[];
  setSelectedColumns: (columns: string[]) => void;
  setVisibleColumns: (columns: string[]) => void;
  columns: any[];
  harvestColumns: any[];
  commercializationColumns: any[];
  allColumns: any[];
}) {
  const [active, setActive] = useState("Siembra");
  const [showColumnsModal, setShowColumnsModal] = useState(false);

  const buttonBase =
    "px-4 py-2 text-sm font-medium border border-gray-200 focus:z-10 focus:ring-2 focus:outline-none focus:ring-0 rounded-none";
  const activeClass = "bg-[#547792] text-white";
  const inactiveClass =
    "bg-white text-gray-900 hover:bg-gray-100 hover:text-blue-700 dark:bg-gray-800 dark:border-gray-700";

  return (
    <div className="flex justify-between items-center p-4 bg-white rounded-t-xl border-b border-gray-100">
      <div className="text-sm text-gray-900">
        Campos: <span className="font-semibold mr-2">{fieldsAmount}{" "}</span> 
        Lotes: <span className="font-semibold">{lotsAmount}</span>
      </div>

      <div className="inline-flex rounded-md shadow-xs" role="group">
        <button
          type="button"
          className={`rounded-s-lg ${buttonBase} ${
            active === "Siembra" ? activeClass : inactiveClass
          }`}
          tabIndex={0}
          onClick={() => {
            setActive("Siembra");
            setVisibleColumns(columns.map((col) => col.key));
            setSelectedColumns(columns.map((col) => col.key));
          }}
        >
          Siembra
        </button>
        <button
          type="button"
          className={`${buttonBase} border-l-0 border-r-0 ${
            active === "Cosecha" ? activeClass : inactiveClass
          }`}
          tabIndex={0}
          onClick={() => {
            setActive("Cosecha");
            setVisibleColumns(harvestColumns.map((col) => col.key));
            setSelectedColumns(harvestColumns.map((col) => col.key));
          }}
        >
          Cosecha
        </button>
        <button
          type="button"
          className={`rounded-e-lg ${buttonBase} ${
            active === "Comercialización" ? activeClass : inactiveClass
          }`}
          tabIndex={0}
          onClick={() => {
            setActive("Comercialización");
            setVisibleColumns(commercializationColumns.map((col) => col.key));
            setSelectedColumns(commercializationColumns.map((col) => col.key));
          }}
        >
          Comercialización
        </button>
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

function Drawer({
  open,
  onClose,
  children,
}: {
  open: boolean;
  onClose: () => void;
  children: React.ReactNode;
}) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-40 flex">
      <div
        className="fixed inset-0 bg-black bg-opacity-30 transition-opacity"
        onClick={onClose}
      />
      <div className="ml-auto h-full w-full max-w-xl bg-white shadow-xl p-8 overflow-y-auto relative animate-slide-in-right">
        <button
          className="absolute top-4 right-4 text-gray-400 hover:text-gray-600"
          onClick={onClose}
          aria-label="Cerrar"
        >
          <svg
            width="24"
            height="24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
        {children}
      </div>
      <style>{`
        @keyframes slide-in-right {
          from { transform: translateX(100%); opacity: 0; }
          to { transform: translateX(0); opacity: 1; }
        }
        .animate-slide-in-right {
          animation: slide-in-right 0.25s cubic-bezier(0.4,0,0.2,1);
        }
      `}</style>
    </div>
  );
}

export function Lots() {
  const navigate = useNavigate();
  const [lot, setLot] = useState<LotsDataUpdate | null>(null);
  const [successMessage, setSuccessMessage] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  const columns: Column<LotsData>[] = [
    {
      key: "project_name",
      header: "Proyecto",
      render: (value, data) => (
        <strong className="text-blue-700">
          <a href={`/admin/database/customers/${data.project_id}`}>
            {value as string}
          </a>
        </strong>
      ),
    },
    { key: "field_name", header: "Campo" },
    { key: "lot_name", header: "Lote" },
    {
      key: "previous_crop",
      header: "Cultivo Ant.",
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
      key: "current_crop",
      header: "Cultivo Act.",
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
      key: "variety",
      header: "Variedad",
      render: (value) => <b>{value}</b>,
    },
    { key: "sowed_area", header: "Sup. total" },
    {
      key: "dates",
      header: "Fecha Siembra",
      render: (value) => {
        if (value?.length > 0) {
          for (let i = value.length - 1; i >= 0; i--) {
            if (value[i].sowing_date) {
              return <b>{value[i].sowing_date}</b>;
            }
          }
        }
        return "";
      },
    },
    {
      key: "cost_per_hectare",
      header: "Costo U$ /HA",
      render: (value) => "$" + value,
    },
  ];

  const harvestColumns: Column<LotsData>[] = [
    ...columns,
    { key: "harvested_area", header: "Sup. Cosecha" },
    {
      key: "harvest_date",
      header: "Fecha Cosecha",
      render: (value) => {
        if (value?.length > 0) {
          for (let i = value.length - 1; i >= 0; i--) {
            if (value[i].harvest_date) {
              return <b>{value[i].harvest_date}</b>;
            }
          }
        }
        return "";
      },
    },
    {
      key: "tons",
      header: "Toneladas",
      render: (value, item) => (
        <EditableCell item={item} value={value} onSuccessEdit={onSuccessEdit} />
      ),
    },
    {
      key: "yield",
      header: "Rendimiento",
      render: (value) => value + " Tn/Has",
    },
  ];

  const commercializationColumns: Column<LotsData>[] = [
    ...harvestColumns,
    {
      key: "net_income",
      header: "Ingreso Neto",
      render: (value) => "$" + value,
    },
    { key: "rent", header: "Arriendo", render: (value) => "$" + value },
    {
      key: "admin_cost",
      header: "Adm. Proyecto",
      render: (value) => "$" + value,
    },
    {
      key: "total_assets",
      header: "Activo Total",
      render: (value) => "$" + value,
    },
    {
      key: "operating_result",
      header: "Resultado Operativo",
      render: (value) => "$" + value,
    },
  ];

  const [columnsToShow, setColumnsToShow] = useState(columns);

  const allColumnsMap = new Map();
  [...columns, ...harvestColumns, ...commercializationColumns].forEach(
    (col) => {
      allColumnsMap.set(col.key, col);
    }
  );
  const allColumns = Array.from(allColumnsMap.values());

  const [selectedColumns, setSelectedColumns] = useState(
    allColumns.map((col) => col.key)
  );

  const [visibleColumns, setVisibleColumns] = useState(selectedColumns);

  useEffect(() => {
    setColumnsToShow(
      allColumns.filter((col) => visibleColumns.includes(col.key))
    );
  }, [visibleColumns]);

  useEffect(() => {
    setVisibleColumns(columns.map((col) => col.key));
  }, []);

  const {
    getLots,
    getLotsKpis,
    lots,
    crops,
    getCrops,
    updateLot,
    updateLotError,
    result,
    processing,
    error,
    kpis,
    processingKpis,
    errorKpis,
  } = useLots();

  useEffect(() => {
    getCrops();
  }, [getCrops]);

  const {
    selectedCustomer,
    projectId,
    selectedCampaignId,
    selectedField,
    fields,
    filters,
    seasons,
  } = useWorkspaceFilters(["customer", "project", "campaign", "field"]);

  const [drawerOpen, setDrawerOpen] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (!selectedCustomer || !projectId || !selectedCampaignId) {
      setMessage("Seleccione un proyecto, campaña y campo para ver resultados");
      return;
    }
    setMessage("");
    if (selectedField) {
      setCurrentPage(1);
      getLots(`field_id=${selectedField.id}`);
      getLotsKpis(`field_id=${selectedField.id}`);
    } else {
      setCurrentPage(1);
      getLots(`project_id=${projectId}`);
      getLotsKpis(`project_id=${projectId}`);
    }
  }, [getLots, selectedCustomer, projectId, selectedCampaignId, selectedField]);

  useEffect(() => {
    if (result !== "") {
      setSuccessMessage(result);
      if (selectedField) {
        setCurrentPage(1);
        getLots(`field_id=${selectedField.id}`);
        getLotsKpis(`field_id=${selectedField.id}`);
      } else if (projectId) {
        setCurrentPage(1);
        getLots(`project_id=${projectId}`);
        getLotsKpis(`project_id=${projectId}`);
      }
    }
  }, [result]);

  useEffect(() => {
    if (drawerOpen && lot) {
      const newLot = lots.find((l) => l.id === lot.id);
      if (newLot) {
        setLot(newLot);
      }
    }
  }, [lots]);

  useEffect(() => {
    if (updateLotError && updateLotError !== "") {
      setErrorMessage(updateLotError);
      setSuccessMessage("");
    }
  }, [updateLotError]);

  const onSuccessEdit = () => {
    if (selectedField) {
      setCurrentPage(1);
      getLots(`field_id=${selectedField.id}`);
      getLotsKpis(`field_id=${selectedField.id}`);
    } else if (projectId) {
      setCurrentPage(1);
      getLots(`project_id=${projectId}`);
      getLotsKpis(`project_id=${projectId}`);
    }
  };

  const handlePageChange = (newPage: number) => {
    setCurrentPage(newPage);
  };

  const paginatedLots = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return lots.slice(startIndex, startIndex + itemsPerPage);
  }, [lots, currentPage, itemsPerPage]);

  const handleCreateLot = () => {
    if (selectedField && projectId && selectedCustomer && selectedCampaignId) {
      navigate(`/admin/database/customers/${selectedField.project_id}`);
      return;
    }
  };

  function handleLotChange<K extends keyof LotsDataUpdate>(
    key: K,
    value: LotsDataUpdate[K]
  ) {
    setLot((prev) => ({
      ...prev,
      id: prev?.id || 0,
      lot_name: prev?.lot_name || "",
      field_id: prev?.field_id || 0,
      previous_crop_id: prev?.previous_crop_id || 0,
      current_crop_id: prev?.current_crop_id || 0,
      variety: prev?.variety || "",
      sowed_area: prev?.sowed_area || "",
      dates: prev?.dates || [],
      season: prev?.season || "",
      [key]: value,
      updated_at: prev?.updated_at || new Date().toISOString(),
    }));
  }

  const handleSave = () => {
    if (lot) {
      const lotToUpdate = { ...lot };
      if (lotToUpdate.dates) {
        const invalidDate = lotToUpdate.dates.find(
          (date) =>
            date &&
            date.harvest_date &&
            (!date.sowing_date || date.sowing_date === "")
        );
        if (invalidDate) {
          setErrorMessage(
            "Si hay fecha de cosecha, debe cargar también la fecha de siembra."
          );
          return;
        }
      }

      if (lotToUpdate.sowed_area === "" || lotToUpdate.sowed_area === "0") {
        setErrorMessage("Area de siembra obligatoria");
        return;
      }
      setSuccessMessage("");
      setErrorMessage("");
      updateLot(lotToUpdate);
    }
  };

  const handleExport = async () => {
    if (!projectId) return;

    try {
      const response = await request.get<Blob>(
        `/lots/export/${projectId}`,
        undefined,
        { responseType: "blob" }
      );

      const url = window.URL.createObjectURL(response);

      const link = document.createElement("a");
      link.href = url;
      link.download = `lotes_${projectId}_${new Date().toISOString()}.xlsx`;
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
            label: "Exportar lotes",
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
            label: "+ Nuevo lote",
            variant: "success",
            isPrimary: true,
            disabled:
              !projectId ||
              !selectedCampaignId ||
              !selectedCustomer ||
              !selectedField,
            onClick: handleCreateLot,
          },
        ]}
      />
      {message && (
        <div
          className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
          role="alert"
        >
          <span className="font-medium">{message}</span>
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
      {!message && !error && (
        <div className="my-4">
          <LotsIndicators
            kpis={kpis}
            processing={processingKpis}
            error={errorKpis}
          />
        </div>
      )}
      <div className="mt-4 relative">
        {processing && (
          <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
            <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
          </div>
        )}
        <Drawer open={drawerOpen} onClose={() => setDrawerOpen(false)}>
          <div className="flex flex-col h-full">
            <h2 className="text-lg font-semibold mb-4">
              {lot?.project_name} ({selectedField?.name}: {lot?.lot_name})
            </h2>
            {processing ? (
              <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
                <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
              </div>
            ) : (
              <>
                <form className="space-y-4 flex-1">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="w-48">
                      <InputField
                        label="Fecha de siembra"
                        name="sowingDate"
                        type="date"
                        value={lot?.dates?.[0]?.sowing_date || ""}
                        onChange={(e) => {
                          const newDates = [...(lot?.dates || [])];
                          newDates[0] = {
                            ...newDates[0],
                            sowing_date: e.target.value,
                            sequence: 1,
                            harvest_date: newDates[0]?.harvest_date || "",
                          };
                          handleLotChange("dates", newDates);
                        }}
                        size="sm"
                      />
                    </div>
                    <div className="w-48">
                      <InputField
                        label="Fecha de cosecha"
                        placeholder="yyyy-dd-mm"
                        name="harvestDate"
                        type="date"
                        value={lot?.dates?.[0]?.harvest_date || ""}
                        onChange={(e) => {
                          const newDates = [...(lot?.dates || [])];
                          newDates[0] = {
                            ...newDates[0],
                            harvest_date: e.target.value,
                            sequence: 1,
                            sowing_date: newDates[0]?.sowing_date || "",
                          };
                          handleLotChange("dates", newDates);
                        }}
                        size="sm"
                      />
                    </div>
                    <div className="w-48">
                      <InputField
                        label=""
                        placeholder="yyyy-dd-mm"
                        name="sowingDate"
                        type="date"
                        value={lot?.dates?.[1]?.sowing_date || ""}
                        onChange={(e) => {
                          const newDates = [...(lot?.dates || [])];
                          newDates[1] = {
                            ...newDates[1],
                            sowing_date: e.target.value,
                            sequence: 2,
                            harvest_date: newDates[1]?.harvest_date || "",
                          };
                          handleLotChange("dates", newDates);
                        }}
                        size="sm"
                      />
                    </div>
                    <div className="w-48">
                      <InputField
                        label=""
                        placeholder="yyyy-dd-mm"
                        name="harvestDate"
                        type="date"
                        value={lot?.dates?.[1]?.harvest_date || ""}
                        onChange={(e) => {
                          const newDates = [...(lot?.dates || [])];
                          newDates[1] = {
                            ...newDates[1],
                            harvest_date: e.target.value,
                            sequence: 2,
                            sowing_date: newDates[1]?.sowing_date || "",
                          };
                          handleLotChange("dates", newDates);
                        }}
                        size="sm"
                      />
                    </div>
                    <div className="w-48">
                      <InputField
                        label=""
                        placeholder="yyyy-dd-mm"
                        name="sowingDate"
                        type="date"
                        value={lot?.dates?.[2]?.sowing_date || ""}
                        onChange={(e) => {
                          const newDates = [...(lot?.dates || [])];
                          newDates[2] = {
                            ...newDates[2],
                            sowing_date: e.target.value,
                            sequence: 3,
                            harvest_date: newDates[2]?.harvest_date || "",
                          };
                          handleLotChange("dates", newDates);
                        }}
                        size="sm"
                      />
                    </div>
                    <div className="w-48">
                      <InputField
                        label=""
                        placeholder="yyyy-dd-mm"
                        name="harvestDate"
                        type="date"
                        value={lot?.dates?.[2]?.harvest_date || ""}
                        onChange={(e) => {
                          const newDates = [...(lot?.dates || [])];
                          newDates[2] = {
                            ...newDates[2],
                            harvest_date: e.target.value,
                            sequence: 3,
                            sowing_date: newDates[2]?.sowing_date || "",
                          };
                          handleLotChange("dates", newDates);
                        }}
                        size="sm"
                      />
                    </div>
                    <div className="col-span-2">
                      <hr />
                    </div>
                    <div>
                      <InputField
                        label="Lote"
                        name="lotName"
                        type="text"
                        value={lot?.lot_name || ""}
                        onChange={(e) =>
                          handleLotChange("lot_name", e.target.value)
                        }
                        size="sm"
                      />
                    </div>
                    <div>
                      <InputField
                        label="Hectáreas"
                        name="hectares"
                        type="number"
                        value={lot?.sowed_area || ""}
                        onChange={(e) =>
                          handleLotChange("sowed_area", e.target.value)
                        }
                        size="sm"
                      />
                    </div>
                    <div>
                      <SelectField
                        label="Cultivo anterior"
                        placeholder="Seleccione cultivo"
                        name="previousCrop"
                        options={crops}
                        value={String(lot?.previous_crop_id || "")}
                        onChange={(e) =>
                          handleLotChange(
                            "previous_crop_id",
                            Number(e.target.value)
                          )
                        }
                        fullWidth
                        size="sm"
                      />
                    </div>
                    <div>
                      <SelectField
                        label="Cultivo actual"
                        placeholder="Seleccione cultivo"
                        name="currentCrop"
                        options={crops}
                        value={String(lot?.current_crop_id || "")}
                        onChange={(e) =>
                          handleLotChange(
                            "current_crop_id",
                            Number(e.target.value)
                          )
                        }
                        size="sm"
                        fullWidth
                      />
                    </div>
                    <div>
                      <SelectField
                        label="Periodo"
                        name="season"
                        value={lot?.season || ""}
                        onChange={(e) =>
                          handleLotChange("season", e.target.value)
                        }
                        options={seasons}
                        size="sm"
                        fullWidth
                      />
                    </div>
                    <div>
                      <InputField
                        label="Variedad"
                        name="variety"
                        type="text"
                        value={lot?.variety || ""}
                        onChange={(e) =>
                          handleLotChange("variety", e.target.value)
                        }
                        size="sm"
                      />
                    </div>
                  </div>
                  {errorMessage && errorMessage !== "" && (
                    <div
                      className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
                      role="alert"
                    >
                      <span className="font-medium">Error!</span> {errorMessage}
                      <button
                        type="button"
                        className="ms-auto -mx-1 -my-1 bg-red-50 text-red-500 rounded-lg focus:ring-2 focus:ring-red-400 p-1.5 hover:bg-red-200 inline-flex items-center justify-center h-8 w-8 dark:bg-gray-800 dark:text-red-400 dark:hover:bg-gray-700"
                        aria-label="Close"
                        onClick={() => setErrorMessage("")}
                      >
                        <span className="sr-only">Close</span>
                        <svg
                          className="w-2 h-2"
                          aria-hidden="true"
                          xmlns="http://www.w3.org/2000/svg"
                          fill="none"
                          viewBox="0 0 14 14"
                        >
                          <path
                            stroke="currentColor"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth="2"
                            d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
                          />
                        </svg>
                      </button>
                    </div>
                  )}
                  {successMessage && successMessage !== "" && (
                    <div
                      className="flex items-center p-4 mb-4 text-sm text-green-800 rounded-lg bg-green-50 dark:bg-gray-800 dark:text-green-400"
                      role="alert"
                    >
                      <svg
                        className="shrink-0 inline w-4 h-4 me-3"
                        aria-hidden="true"
                        xmlns="http://www.w3.org/2000/svg"
                        fill="currentColor"
                        viewBox="0 0 20 20"
                      >
                        <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z" />
                      </svg>
                      <span className="sr-only">Info</span>
                      <div>
                        <span className="font-medium">{successMessage}</span>
                      </div>
                      <button
                        type="button"
                        className="ms-auto -mx-1.5 -my-1.5 bg-green-50 text-green-500 rounded-lg focus:ring-2 focus:ring-green-400 p-1.5 hover:bg-green-200 inline-flex items-center justify-center h-8 w-8 dark:bg-gray-800 dark:text-green-400 dark:hover:bg-gray-700"
                        data-dismiss-target="#alert-3"
                        aria-label="Close"
                        onClick={() => setSuccessMessage("")}
                      >
                        <span className="sr-only">Close</span>
                        <svg
                          className="w-3 h-3"
                          aria-hidden="true"
                          xmlns="http://www.w3.org/2000/svg"
                          fill="none"
                          viewBox="0 0 14 14"
                        >
                          <path
                            stroke="currentColor"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth="2"
                            d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
                          />
                        </svg>
                      </button>
                    </div>
                  )}
                </form>
                <div className="flex justify-end gap-2 mt-auto pt-6 bg-white">
                  <Button
                    variant="outlineGray"
                    className="text-base font-medium"
                    onClick={() => setDrawerOpen(false)}
                  >
                    Cancelar
                  </Button>
                  <Button
                    variant="success"
                    className="text-base font-medium"
                    onClick={handleSave}
                  >
                    Guardar
                  </Button>
                </div>
              </>
            )}
          </div>
        </Drawer>
        {!message && !error && (
          <DataTable
            data={paginatedLots}
            columns={columnsToShow}
            headerComponent={
              <LotsHeader
                fieldsAmount={fields.length}
                lotsAmount={lots.length}
                selectedColumns={selectedColumns}
                setSelectedColumns={setSelectedColumns}
                setVisibleColumns={setVisibleColumns}
                columns={columns}
                harvestColumns={harvestColumns}
                commercializationColumns={commercializationColumns}
                allColumns={allColumns}
              />
            }
            onEdit={(item) => {
              setLot(item);
              setSuccessMessage("");
              setDrawerOpen(true);
            }}
            message="No hay lotes disponibles"
            pagination={{
              page: currentPage,
              perPage: itemsPerPage,
              total: lots.length,
              onPageChange: handlePageChange,
            }}
          />
        )}
      </div>
    </div>
  );
}

type Column<T> = {
  key: keyof T;
  header: string;
  render?: (value: any, item: T) => React.ReactNode;
};

export default Lots;
