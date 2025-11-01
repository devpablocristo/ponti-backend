import { useEffect, useMemo, useState } from "react";

import FilterBar from "../../../../layout/FilterBar/FilterBar";
import { useWorkspaceFilters } from "../../../../hooks/useWorkspaceFilters";
import useProducts from "../../../../hooks/useProducts";
import DataTable from "../../../../components/Table/DataTable";
import { Supply } from "../../../../hooks/useProducts/types";
import Button from "../../../../components/Button/Button";
import { Column } from "../../types";
import { BaseModal } from "../../../../components/Modal/BaseModal";
import InputField from "../../../../components/Input/InputField";
import SelectField from "../../../../components/Input/SelectField";
import { units } from "./Items";
import useCategories from "../../../../hooks/useCategories";
import APIClient from "../../../../restclient/apiInstance";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const columns: Column<Supply>[] = [
  { key: "id", header: "ID" },
  {
    key: "name",
    header: "Nombre",
    render: (value) => <strong className="text-blue-700">{value}</strong>,
  },
  {
    key: "price",
    header: "Precio",
    render: (value) => <strong>{value}</strong>,
  },
  {
    key: "category_name",
    header: "Rubro",
    render: (value) => value,
  },
  {
    key: "type_name",
    header: "Tipo",
    render: (value) => value,
  },
];

export default function ListItems() {
  const {
    getSupplies,
    error,
    supplies,
    updateSupply,
    deleteSupply,
    result,
    processing,
    errorUpdate,
    resultUpdate,
  } = useProducts();
  const { categories, types, getCategories, getTypes } = useCategories();

  const [errorMessage, setErrorMessage] = useState<string>("");

  const [modalOpen, setModalOpen] = useState(false);
  const [item, setItem] = useState<Supply | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  const { filters, projectId } = useWorkspaceFilters([
    "customer",
    "project",
    "campaign",
  ]);

  useEffect(() => {
    if (projectId) {
      getSupplies(projectId);
      getCategories("");
      getTypes();
    }
  }, [projectId]);

  useEffect(() => {
    if (result && projectId) {
      getSupplies(projectId);
    }
  }, [result, projectId]);

  useEffect(() => {
    if (resultUpdate && projectId) {
      alert(resultUpdate);
      getSupplies(projectId);
    }
  }, [resultUpdate, projectId]);

  useEffect(() => {
    if (error) {
      setErrorMessage(error);
    }
  }, [error]);

  useEffect(() => {
    if (errorUpdate) {
      setErrorMessage(errorUpdate);
    }
  }, [errorUpdate]);

  const handleDelete = (id: number) => {
    if (window.confirm("¿Está seguro que desea eliminar el insumo?")) {
      setErrorMessage("");
      deleteSupply(id);

      setTimeout(() => {
        const totalAfterDelete = supplies.length - 1;
        const lastPage = Math.max(
          1,
          Math.ceil(totalAfterDelete / itemsPerPage)
        );
        if (currentPage > lastPage) {
          setCurrentPage(lastPage);
        }
      }, 200);
    }
  };

  const handleEdit = (item: Supply) => {
    setItem(item);
    setModalOpen(true);
  };

  const handleSave = () => {
    if (processing) return;
    if (item && projectId) {
      updateSupply(projectId, item);
      setModalOpen(false);
    }
  };

  const paginatedSupplies = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return supplies.slice(startIndex, startIndex + itemsPerPage);
  }, [supplies, currentPage, itemsPerPage]);

  const handleExport = async () => {
    if (!projectId) return;

    try {
      const response = await request.get<Blob>(
        `/supply_movements/database-export/${projectId}`,
        undefined,
        { responseType: "blob" }
      );

      const url = window.URL.createObjectURL(response);

      const link = document.createElement("a");
      link.href = url;
      link.download = `insumosbd_${projectId}_${new Date().toISOString()}.xlsx`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <div className="w-full mx-auto">
      <FilterBar filters={filters} actions={[
          {
            label: "Exportar insumos",
            icon: <svg width="14" height="13" viewBox="0 0 14 13" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M5.66675 2.49984H3.00008C2.64646 2.49984 2.30732 2.64031 2.05727 2.89036C1.80722 3.14041 1.66675 3.47955 1.66675 3.83317V10.4998C1.66675 10.8535 1.80722 11.1926 2.05727 11.4426C2.30732 11.6927 2.64646 11.8332 3.00008 11.8332H9.66675C10.0204 11.8332 10.3595 11.6927 10.6096 11.4426C10.8596 11.1926 11.0001 10.8535 11.0001 10.4998V7.83317M8.33341 1.1665H12.3334M12.3334 1.1665V5.1665M12.3334 1.1665L5.66675 7.83317" stroke="#547792" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
            ,
            variant: "outlinePonti",
            isPrimary: true,
            disabled: !projectId,
            onClick: () => handleExport(),
          }
        ]}/>
      <div className="p-6 w-full mt-4 mx-auto bg-white rounded-lg shadow-md">
        {errorMessage && (
          <div
            id="alert-2"
            className="flex items-center p-4 mb-4 text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
            role="alert"
          >
            <svg
              className="shrink-0 w-4 h-4"
              aria-hidden="true"
              xmlns="http://www.w3.org/2000/svg"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z" />
            </svg>
            <span className="sr-only">Error</span>
            <div className="ms-3 text-sm font-medium">{errorMessage}</div>
            <button
              type="button"
              className="ms-auto -mx-1.5 -my-1.5 bg-red-50 text-red-500 rounded-lg focus:ring-2 focus:ring-red-400 p-1.5 hover:bg-red-200 inline-flex items-center justify-center h-8 w-8"
              onClick={() => setErrorMessage("")}
              aria-label="Close"
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
        <div className="flex justify-between items-center">
          <h1 className="text-custom-text font-semibold text-xl leading-none">
            Lista de insumos del proyecto
          </h1>
          <Button
            variant="secondary"
            size="sm"
            className="text-sm font-medium flex items-center gap-1"
            href="/admin/database/items"
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
                d="M10 19l-7-7m0 0l7-7m-7 7h18"
              />
            </svg>
            Volver
          </Button>
        </div>
        <div className="mt-4">
          <BaseModal
            isOpen={modalOpen}
            onClose={() => {
              setModalOpen(false);
              setItem(null);
            }}
            title={`Edicion de insumo ${item?.name || ""}`}
            primaryButtonText="Guardar"
            onPrimaryAction={handleSave}
          >
            <div className="flex flex-col gap-1">
              <InputField
                label="Nombre del insumo"
                placeholder="Nombre del insumo"
                name="name"
                type="text"
                value={item?.name || ""}
                onChange={(e) => {
                  if (!item) return;
                  setItem({ ...item, name: e.target.value });
                }}
              />
              <div className="flex gap-3">
                <div className="flex-1">
                  <SelectField
                    label="Unidad"
                    name={`unit-${item?.name || ""}`}
                    value={item?.unit_id?.toString() || ""}
                    onChange={(e) => {
                      if (!item) return;
                      setItem({ ...item, unit_id: parseInt(e.target.value) });
                    }}
                    options={units}
                  />
                </div>
                <div className="flex-2">
                  <InputField
                    label="Precio"
                    placeholder="Precio"
                    name="price"
                    type="text"
                    value={item?.price || ""}
                    onChange={(e) => {
                      if (!item) return;
                      let value = e.target.value.replace(/,/g, ".");
                      if (/^\d*\.?\d{0,3}$/.test(value)) {
                        setItem({ ...item, price: value} );
                      }
                    }}
                  />
                </div>
              </div>
              <SelectField
                label="Rubro"
                name={`category-${item?.name || ""}`}
                value={item?.category_id?.toString() || ""}
                onChange={(e) => {
                  if (!item) return;
                  const category = parseInt(e.target.value);
                  const cat = categories.find((cat) => cat.id === category);

                  setItem({
                    ...item,
                    category_id: category,
                    type_id: cat?.type_id || 0,
                  });
                }}
                options={categories}
              />
              <SelectField
                label=""
                name={`type`}
                value={item?.type_id?.toString() || ""}
                disabled
                onChange={(e) => {
                  if (!item) return;
                  setItem({ ...item, type_id: parseInt(e.target.value) });
                }}
                options={types}
              />
            </div>
          </BaseModal>
          <DataTable
            data={paginatedSupplies}
            columns={columns}
            onDelete={(item) => handleDelete(item.id)}
            onEdit={(item) => handleEdit(item)}
            message="No hay insumos cargados en el proyecto"
            pagination={{
              page: currentPage,
              perPage: itemsPerPage,
              total: supplies.length,
              onPageChange: (newPage: number) => setCurrentPage(newPage),
            }}
          />
        </div>
      </div>
    </div>
  );
}
