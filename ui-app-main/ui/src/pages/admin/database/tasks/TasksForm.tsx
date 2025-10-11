import { useEffect, useState } from "react";
import InputField from "../../../../components/Input/InputField";
import Button from "../../../../components/Button/Button";
import SelectField from "../../../../components/Input/SelectField";
import { useWorkspaceFilters } from "../../../../hooks/useWorkspaceFilters";
import FilterBar from "../../../../layout/FilterBar/FilterBar";
import useCategories from "../../../../hooks/useCategories";
import { TaskToSave } from "../../../../hooks/useTasks/types";
import useTask from "../../../../hooks/useTasks";
import { LoaderCircle } from "lucide-react";

interface Labor {
  id: number;
  name: string;
  category: string;
  price: string;
  contractor: string;
}

export default function TasksForm() {
  const { saveTasks, result, error, processing } = useTask();

  const [errorMessage, setErrorMessage] = useState<string>("");
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const { categories, getCategories } = useCategories();

  const { filters, projectId } = useWorkspaceFilters([
    "customer",
    "project",
    "campaign",
  ]);

  const [rows, setLabors] = useState<Labor[]>(
    Array.from({ length: 5 }, (_, i) => ({
      id: i,
      name: "",
      category: "",
      price: "",
      contractor: "",
    }))
  );

  useEffect(() => {
    getCategories("type_id=4");
  }, []);

  function cleanForm() {
    setLabors(
      Array.from({ length: 5 }, (_, i) => ({
        id: i,
        name: "",
        category: "",
        price: "",
        contractor: "",
      }))
    );
  }

  useEffect(() => {
    if (result !== "") {
      cleanForm();
      setTimeout(() => {
        document
          .getElementById("main-scroll")
          ?.scrollTo({ top: 0, behavior: "smooth" });
      }, 0);
    }
    setErrorMessage("");
    setSuccessMessage(result);
  }, [result]);

  useEffect(() => {
    if (error) {
      setErrorMessage(error);
    }
  }, [error]);

  useEffect(() => {
    if (errorMessage) {
      setSuccessMessage(null);
    }
  }, [errorMessage]);

  const handleChange = (id: number, field: keyof Labor, value: string) => {
    setLabors((prev) =>
      prev.map((labor) =>
        labor.id === id ? { ...labor, [field]: value } : labor
      )
    );
  };

  function hasIncompleteRows(rows: Labor[]) {
    return rows.some((row) => {
      const anyFilled = row.name || row.category || row.price || row.contractor;
      const anyMissing =
        !row.name || !row.category || !row.price || !row.contractor;
      return anyFilled && anyMissing;
    });
  }

  const handleCreateLabors = () => {
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

    const laborsToSave: TaskToSave[] = rows
      .filter((row) => row.name && row.category && row.price && row.contractor)
      .map((row) => ({
        name: row.name,
        category_id: Number(row.category),
        price: row.price,
        contractor_name: row.contractor,
      }));

    if (laborsToSave.length === 0) {
      setErrorMessage(
        "Por favor, ingrese al menos un insumo antes de guardar."
      );
      return;
    }

    saveTasks(laborsToSave, projectId);
  };

  return (
    <div className="w-full mx-auto">
      <FilterBar filters={filters} />
      <div className="w-full p-6 mt-4 bg-white rounded-lg shadow-md">
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
            Agregar labores
          </h1>
          <Button
            variant="primary"
            size="sm"
            className="text-sm font-medium flex items-center gap-1"
            href="/admin/database/tasks/list"
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
        {processing ? (
          <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
            <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
          </div>
        ) : (
          <div className="mt-4">
            <div className="hidden sm:grid grid-cols-[1fr_1fr_0.5fr_1fr] gap-4 mb-2">
              <span className="font-semibold">Labor</span>
              <span className="font-semibold">Rubro</span>
              <span className="font-semibold">Precio</span>
              <span className="font-semibold">Contratista</span>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-[1fr_1fr_0.5fr_1fr] gap-4">
              {rows.map((row, index) => (
                <div
                  key={index}
                  className="sm:contents border sm:border-0 p-4 sm:p-0 rounded-md sm:rounded-none mb-4 sm:mb-0 shadow-sm sm:shadow-none"
                >
                  <div className="sm:col-span-1">
                    <label className="sm:hidden text-sm text-gray-600">
                      Labor
                    </label>
                    <InputField
                      label=""
                      name={`labor-${index}`}
                      value={row.name}
                      onChange={(e) =>
                        handleChange(index, "name", e.target.value)
                      }
                      placeholder="nombre"
                    />
                  </div>
                  <div className="sm:col-span-1">
                    <label className="sm:hidden text-sm text-gray-600">
                      Rubro
                    </label>
                    <SelectField
                      key={row.id}
                      label=""
                      name={`category-${index}`}
                      value={row.category.toString()}
                      onChange={(e) =>
                        handleChange(row.id, "category", e.target.value)
                      }
                      options={categories}
                    />
                  </div>
                  <div className="sm:col-span-1">
                    <label className="sm:hidden text-sm text-gray-600">
                      Precio
                    </label>
                    <InputField
                      label=""
                      name={`precio-${index}`}
                      value={row.price}
                      onChange={(e) => {
                        let value = e.target.value.replace(/,/g, ".");
                        if (/^\d*\.?\d{0,2}$/.test(value)) {
                          handleChange(index, "price", value);
                        }
                      }}
                      placeholder="u$s"
                    />
                  </div>
                  <div className="sm:col-span-1">
                    <label className="sm:hidden text-sm text-gray-600">
                      Contratista
                    </label>
                    <InputField
                      label=""
                      name={`contratista-${index}`}
                      value={row.contractor}
                      onChange={(e) =>
                        handleChange(index, "contractor", e.target.value)
                      }
                      placeholder="nombre"
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
      <div className="flex justify-between flex-wrap gap-4 my-4">
        <div>
          <button className="text-blue-600 hover:underline mr-4">
            Importar labores
          </button>
          <button className="text-blue-600 hover:underline">
            Exportar labores
          </button>
        </div>
        <div className="flex gap-4 my-2 justify-end">
          <Button variant="outlineGray" className="text-base font-medium">
            Cancelar
          </Button>
          <Button
            onClick={handleCreateLabors}
            variant="success"
            className="text-base font-medium"
            disabled={processing}
          >
            Guardar
          </Button>
        </div>
      </div>
    </div>
  );
}
