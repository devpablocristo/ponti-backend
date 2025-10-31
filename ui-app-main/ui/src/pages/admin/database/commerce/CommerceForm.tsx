import { useEffect, useState } from "react";
import InputField from "../../../../components/Input/InputField";
import Button from "../../../../components/Button/Button";
import FilterBar from "../../../../layout/FilterBar/FilterBar";
import { useWorkspaceFilters } from "../../../../hooks/useWorkspaceFilters";
import useProjects from "../../../../hooks/useDatabase/projects";
import useCommerce from "../../../../hooks/useCommerce";
import { LoaderCircle } from "lucide-react";

interface Commerce {
  id: number;
  cropId: number;
  cropName: string;
  boardPrice: string;
  freightCost: string;
  commercialCost: string;
  netWorth: string;
}

export default function CommerceForm() {
  const [errorMessage, setErrorMessage] = useState<string>("");
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [projectCropList, setProjectCropList] = useState<
    { id: number; name: string }[]
  >([]);

  const {
    getProject,
    selectedProject,
    processing: processingProjects,
  } = useProjects();

  const {
    saveCommerceInfo,
    getCommerceInfo,
    processing,
    error,
    result,
    commerceInfoList,
  } = useCommerce();
  const [rows, setRows] = useState<Commerce[]>([]);

  const { filters, projectId } = useWorkspaceFilters([
    "customer",
    "project",
    "campaign",
  ]);

  useEffect(() => {
    if (!projectId) {
      setErrorMessage("Por favor, seleccione un proyecto y campaña.");
      return;
    }

    getProject(projectId);
  }, [projectId]);

  useEffect(() => {
    if (!projectId) {
      setErrorMessage("Por favor, seleccione un proyecto y campaña.");
      return;
    }

    getCommerceInfo(projectId);
    setErrorMessage("");
  }, [projectId]);

  useEffect(() => {
    if (!selectedProject) return;
    const crops = selectedProject.fields.flatMap((field) =>
      field.lots.map((lot) => ({
        id: lot.current_crop_id,
        name: lot.current_crop_name || "",
      }))
    );

    const uniqueCrops = crops.filter(
      (crop, index, self) => index === self.findIndex((c) => c.id === crop.id)
    );

    setProjectCropList(uniqueCrops);
  }, [selectedProject]);

  useEffect(() => {
    if (projectCropList.length && rows.length === 0) {
      setRows(
        projectCropList.map((crop) => {
          const found = commerceInfoList.find((ci) => ci.crop_id === crop.id);
          return {
            id: found?.id || 0,
            cropId: crop.id,
            cropName: crop.name,
            boardPrice: found?.board_price || "",
            freightCost: found?.freight_cost || "",
            commercialCost: found?.commercial_cost || "",
            netWorth: found?.net_price || "",
          };
        })
      );
    }
  }, [projectCropList, commerceInfoList]);

  useEffect(() => {
    if (result !== "") {
      window.location.reload();
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

  const handleChange = (index: number, field: string, value: string) => {
    setRows((prevRows) => {
      const newRows = [...prevRows];
      newRows[index] = { ...newRows[index], [field]: value };
      return newRows;
    });
  };

  function hasIncompleteRows(rows: Commerce[]) {
    const hasPartial = rows.some(
      (item) =>
        !item.cropId ||
        !item.boardPrice ||
        !item.freightCost ||
        !item.commercialCost
    );

    if (hasPartial) {
      return true;
    }

    return false;
  }

  const handleSaveCommerceValues = () => {
    if (!projectId) {
      setErrorMessage("Por favor, seleccione un proyecto y campaña.");
      return;
    }

    setErrorMessage("");
    const itemsWithAnyValue = rows.filter(
      (item) => item.boardPrice || item.freightCost || item.commercialCost
    );

    if (hasIncompleteRows(itemsWithAnyValue)) {
      setErrorMessage(
        "Por favor, complete todos los campos del registro antes de guardar."
      );
      return;
    }

    if (itemsWithAnyValue.length === 0) {
      setErrorMessage(
        "Por favor, ingrese al menos un cultivo antes de guardar."
      );
      return;
    }

    const commerceData = itemsWithAnyValue.map((row) => ({
      id: row.id,
      crop_id: row.cropId,
      board_price: row.boardPrice,
      freight_cost: row.freightCost,
      commercial_cost: row.commercialCost,
    }));

    saveCommerceInfo(commerceData, projectId);
  };

  return (
    <div className="w-full mx-auto">
      <FilterBar filters={filters} />
      <div className="mt-4 p-6 w-full mx-auto bg-white rounded-lg shadow-md">
        <h1 className="text-custom-text font-semibold text-xl leading-none">
          Datos de comercialización por cultivo
        </h1>
        {errorMessage && (
          <div
            id="alert-2"
            className="flex items-center p-4 my-2 text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
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
            className="flex items-center p-4 my-2 text-sm text-green-800 rounded-lg bg-green-50 dark:bg-gray-800 dark:text-green-400"
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
        {processingProjects || processing ? (
          <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
            <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
          </div>
        ) : (
          <div className="mt-1">
            <div className="w-full px-4 py-6">
              <div>
                <div className="hidden sm:grid grid-cols-[0.5fr_0.5fr_0.5fr_0.5fr_0.5fr] gap-4 mb-2">
                  <span className="font-medium">Cultivo</span>
                  <span className="font-medium">Precio pizarra</span>
                  <span className="font-medium">Costo flete</span>
                  <span className="font-medium">Gastos comerciales</span>
                  <span className="font-medium">Precio neto</span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-[0.5fr_0.5fr_0.5fr_0.5fr_0.5fr] gap-4">
                  {rows.map((crop, index) => (
                    <div
                      key={index}
                      className="sm:contents border sm:border-0 p-4 sm:p-0 rounded-md sm:rounded-none mb-4 sm:mb-0 shadow-sm sm:shadow-none"
                    >
                      <div className="sm:col-span-1">
                        <label className="sm:hidden text-sm text-gray-600">
                          Cultivo
                        </label>
                        <InputField
                          label=""
                          name={`crop-${index}`}
                          value={crop.cropName}
                          onChange={() => {}}
                          disabled
                        />
                      </div>
                      <div className="sm:col-span-1">
                        <label className="sm:hidden text-sm text-gray-600">
                          Precio pizarra
                        </label>
                        <InputField
                          label=""
                          name={`price-${index}`}
                          value={crop.boardPrice}
                          onChange={(e) => {
                            let value = e.target.value.replace(/,/g, ".");
                            if (/^\d*\.?\d{0,2}$/.test(value)) {
                              handleChange(index, "boardPrice", value);
                            }
                          }}
                          placeholder="u$s"
                        />
                      </div>
                      <div className="sm:col-span-1">
                        <label className="sm:hidden text-sm text-gray-600">
                          Costo flete
                        </label>
                        <InputField
                          label=""
                          name={`cost-${index}`}
                          value={crop.freightCost}
                          onChange={(e) => {
                            let value = e.target.value.replace(/,/g, ".");
                            if (/^\d*\.?\d{0,2}$/.test(value)) {
                              handleChange(index, "freightCost", value);
                            }
                          }}
                          placeholder="u$s"
                        />
                      </div>
                      <div className="sm:col-span-1">
                        <label className="sm:hidden text-sm text-gray-600">
                          Gastos comerciales %
                        </label>
                        <InputField
                          label=""
                          name={`expenses-${index}`}
                          value={crop.commercialCost}
                          onChange={(e) => {
                            let value = e.target.value.replace(/,/g, ".");
                            if (/^\d*\.?\d{0,2}$/.test(value)) {
                              handleChange(index, "commercialCost", value);
                            }
                          }}
                          placeholder="%"
                        />
                      </div>
                      <div className="sm:col-span-1">
                        <label className="sm:hidden text-sm text-gray-600">
                          Precio neto
                        </label>
                        <InputField
                          label=""
                          name={`netWorth-${index}`}
                          value={crop.netWorth}
                          onChange={() => {}}
                          disabled
                          placeholder="u$s"
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
      <div className="flex gap-4 my-4 justify-end">
        <Button variant="outlineGray" className="text-base font-medium">
          Cancelar
        </Button>
        <Button
          onClick={handleSaveCommerceValues}
          disabled={processing || processingProjects || !selectedProject}
          variant="success"
          className="text-base font-medium"
        >
          Guardar
        </Button>
      </div>
    </div>
  );
}
