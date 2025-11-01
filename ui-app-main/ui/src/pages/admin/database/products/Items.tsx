import { useEffect, useState } from "react";
// Eliminamos las importaciones no utilizadas
import InputField from "../../../../components/Input/InputField";
import Button from "../../../../components/Button/Button";
import SelectField from "../../../../components/Input/SelectField";
import FilterBar from "../../../../layout/FilterBar/FilterBar";
import { useWorkspaceFilters } from "../../../../hooks/useWorkspaceFilters";
import { Product } from "../../../../hooks/useProducts/types";
import useProducts from "../../../../hooks/useProducts";
import useCategories from "../../../../hooks/useCategories";

interface Row {
  id: number;
  name: string;
  unit: string;
  price: string;
  type: string;
  category: string;
}

export const units = [
  { id: 1, name: "Lts" },
  { id: 2, name: "Kg" },
];

export default function Items() {
  const { saveProducts, result, error } = useProducts();

  const [errorMessage, setErrorMessage] = useState<string>("");
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState<boolean>(false);
  const [initialRows, setInitialRows] = useState<string>("");
  const { categories, types, getCategories, getTypes } = useCategories();
  const {
    filters,
    selectedProject,
    projectId,
    selectedCustomer,
    selectedCampaignId,
  } = useWorkspaceFilters(["customer", "project", "campaign"]);
  const [rows, setRows] = useState<Row[]>(
    Array.from({ length: 5 }, (_, i) => ({
      id: i + 1,
      name: "",
      unit: "",
      price: "",
      type: "",
      category: "",
    }))
  );

  useEffect(() => {
    getCategories("");
    getTypes();

    // Set initial rows state
    setInitialRows(
      JSON.stringify(
        Array.from({ length: 5 }, (_, i) => ({
          id: i + 1,
          name: "",
          unit: "",
          price: "",
          type: "",
          category: "",
        }))
      )
    );

    // Add browser refresh/close protection
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasUnsavedChanges) {
        e.preventDefault();
        const message =
          "Hay cambios sin guardar. Si continúa, los cambios se perderán.";
        e.returnValue = message;
        return message;
      }
    };

    // Add protection for all link clicks in the app
    const handleLinkClick = (e: MouseEvent) => {
      if (!hasUnsavedChanges) return; // Si no hay cambios, permitir navegación normal

      // Buscar si el clic fue en un enlace (a) o en un elemento dentro de un enlace
      let target = e.target as HTMLElement;
      while (target && target.tagName !== "A" && target.tagName !== "BODY") {
        target = target.parentElement as HTMLElement;
      }

      // Si es un enlace y no es el botón de guardar, mostrar confirmación
      if (target && target.tagName === "A") {
        // Ignorar si el enlace es part of the form (guardar, agregar)
        if (target.classList.contains("ignore-protection")) return;

        const confirmed = window.confirm(
          "Hay cambios sin guardar. ¿Desea salir de todas formas?"
        );
        if (!confirmed) {
          e.preventDefault();
          e.stopPropagation();
        }
      }
    };

    window.addEventListener("beforeunload", handleBeforeUnload);
    document.addEventListener("click", handleLinkClick, true); // true para fase de captura

    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
      document.removeEventListener("click", handleLinkClick, true);
    };
  }, [getCategories, getTypes, hasUnsavedChanges]);

  function cleanForm() {
    const emptyRows = Array.from({ length: 5 }, (_, i) => ({
      id: i + 1,
      name: "",
      unit: "",
      price: "",
      type: "",
      category: "",
    }));

    setRows(emptyRows);
    setHasUnsavedChanges(false);
    setInitialRows(JSON.stringify(emptyRows));
  }

  useEffect(() => {
    if (result !== "") {
      cleanForm();
      setTimeout(() => {
        document
          .getElementById("main-scroll")
          ?.scrollTo({ top: 0, behavior: "smooth" });
      }, 0);
      setHasUnsavedChanges(false);
    }
    setErrorMessage("");
    setSuccessMessage(result);
  }, [result]);

  useEffect(() => {
    if (error) {
      setErrorMessage(error);
      setSuccessMessage(null);
    }
  }, [error]);

  const handleChange = (id: number, field: keyof Row, value: string) => {
    setRows((prev) => {
      const newRows = prev.map((row) =>
        row.id === id ? { ...row, [field]: value } : row
      );
      const hasChanges = JSON.stringify(newRows) !== initialRows;
      setHasUnsavedChanges(hasChanges);
      return newRows;
    });
  };

  function hasIncompleteRows(rows: Row[]) {
    return rows.some((row) => {
      const anyFilled =
        row.name || row.unit || row.price || row.type || row.category;
      const anyMissing =
        !row.name || !row.unit || !row.price || !row.type || !row.category;
      return anyFilled && anyMissing;
    });
  }

  const handleCreateSupply = () => {
    if (!projectId) {
      setErrorMessage(
        "Por favor, seleccione un proyecto y campaña antes de guardar."
      );
      return;
    }

    setErrorMessage("");
    if (hasIncompleteRows(rows)) {
      setErrorMessage(
        "Por favor, complete todos los campos del registro antes de guardar."
      );
      return;
    }

    const suppliesToSave: Product[] = rows
      .filter(
        (row) => row.name && row.unit && row.price && row.type && row.category
      )
      .map((row) => ({
        name: row.name,
        unit: Number(row.unit),
        price: Number(row.price),
        type: Number(row.type),
        category: Number(row.category),
      }));

    if (suppliesToSave.length === 0) {
      setErrorMessage(
        "Por favor, ingrese al menos un insumo antes de guardar."
      );
      return;
    }

    if (projectId) {
      saveProducts(suppliesToSave, projectId);
    }
  };

  const handleAddRow = () => {
    setRows((prev) => {
      const newRows = [
        ...prev,
        {
          id: prev.length + 1,
          name: "",
          unit: "",
          price: "",
          type: "",
          category: "",
        },
      ];
      setHasUnsavedChanges(true);
      return newRows;
    });
  };

  return (
    <div className="w-full mx-auto">
      <FilterBar filters={filters} />
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
        <div className="flex justify-between items-center">
          <h1 className="text-custom-text font-semibold text-xl leading-none">
            Agregar insumos
          </h1>
          <Button
            variant="primary"
            size="sm"
            className="text-sm font-medium flex items-center gap-1"
            href="/admin/database/items/list"
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
                d="M4 6h16M4 10h16M4 14h16M4 18h16"
              />
            </svg>
            Ver listado
          </Button>
        </div>
        <div className="mt-4">
          <div className="hidden sm:grid grid-cols-[1fr_0.5fr_0.5fr_1fr_1fr] gap-4 mb-2">
            <span className="font-medium">Insumo</span>
            <span className="font-medium">Unidad</span>
            <span className="font-medium">Precio</span>
            <span className="font-medium">Rubro</span>
            <span className="font-medium">Tipo/Clase</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-[1fr_0.5fr_0.5fr_1fr_1fr] gap-4">
            {rows.map((row, index) => {
              return (
                <div
                  key={index}
                  className="sm:contents border sm:border-0 p-4 sm:p-0 rounded-md sm:rounded-none mb-4 sm:mb-0 shadow-sm sm:shadow-none"
                >
                  <div className="sm:col-span-1">
                    <label className="sm:hidden text-sm text-gray-600">
                      Insumo
                    </label>
                    <InputField
                      label=""
                      name={`item-${index}`}
                      value={row.name}
                      onChange={(e) =>
                        handleChange(row.id, "name", e.target.value)
                      }
                      placeholder="Ingrese nombre"
                    />
                  </div>
                  <div className="sm:col-span-1">
                    <label className="sm:hidden text-sm text-gray-600">
                      Unidad
                    </label>
                    <SelectField
                      key={row.id}
                      label=""
                      name={`unit-${index}`}
                      value={row.unit}
                      onChange={(e) =>
                        handleChange(row.id, "unit", e.target.value)
                      }
                      options={units}
                    />
                  </div>
                  <div className="sm:col-span-1">
                    <label className="sm:hidden text-sm text-gray-600">
                      Precio
                    </label>
                    <InputField
                      label=""
                      name={`price-${index}`}
                      value={row.price}
                      type="text"
                      onChange={(e) => {
                        let value = e.target.value.replace(/,/g, ".");
                        if (/^\d*\.?\d{0,2}$/.test(value)) {
                          handleChange(row.id, "price", value);
                        }
                      }}
                      placeholder="u$s"
                    />
                  </div>
                  <div className="sm:col-span-1">
                    <label className="sm:hidden text-sm text-gray-600">
                      Rubro
                    </label>
                    <SelectField
                      key={`category-${row.id}`}
                      label=""
                      name={`category-${index}`}
                      value={row.category.toString()}
                      onChange={(e) => {
                        const cat = categories.find(
                          (cat) => cat.id === Number(e.target.value)
                        );
                        handleChange(row.id, "category", e.target.value);
                        handleChange(
                          row.id,
                          "type",
                          cat?.type_id?.toString() || ""
                        );
                      }}
                      options={categories}
                    />
                  </div>
                  <div className="sm:col-span-1">
                    <label className="sm:hidden text-sm text-gray-600">
                      Tipo / Clase
                    </label>
                    <SelectField
                      key={`type-${row.id}`}
                      label=""
                      name={`type-${index}`}
                      value={row.type.toString()}
                      disabled
                      onChange={() => {}}
                      options={types}
                    />
                  </div>
                </div>
              );
            })}
          </div>
          <Button
            variant="primary"
            size="sm"
            onClick={handleAddRow}
            className="text-blue-500 hover:underline mt-4"
          >
            + Agregar nuevo insumo
          </Button>
        </div>
      </div>
      <div className="flex gap-4 my-4 justify-end">
        <Button
          variant="outlineGray"
          className="text-base font-medium"
          onClick={() => {
            if (hasUnsavedChanges) {
              const confirmed = window.confirm(
                "Hay cambios sin guardar. ¿Desea cancelar de todas formas?"
              );
              if (confirmed) {
                cleanForm();
              }
            } else {
              cleanForm();
            }
          }}
        >
          Cancelar
        </Button>
        <Button
          variant="success"
          disabled={
            !selectedProject || !selectedCustomer || !selectedCampaignId
          }
          onClick={handleCreateSupply}
          className="text-base font-medium"
        >
          Guardar
        </Button>
      </div>
    </div>
  );
}
