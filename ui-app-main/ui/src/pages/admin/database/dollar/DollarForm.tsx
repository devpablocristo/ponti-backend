import { useEffect, useState } from "react";
import InputField from "../../../../components/Input/InputField";
import Button from "../../../../components/Button/Button";
import { useWorkspaceFilters } from "../../../../hooks/useWorkspaceFilters";
import FilterBar from "../../../../layout/FilterBar/FilterBar";
import useDollar from "../../../../hooks/useDollar";
import { DollarData } from "../../../../hooks/useDollar/types";

interface DollarRow {
  month: string;
  initialValue?: string;
  finalValue?: string;
  averageValue?: string;
}

const months = [
  "Enero",
  "Febrero",
  "Marzo",
  "Abril",
  "Mayo",
  "Junio",
  "Julio",
  "Agosto",
  "Septiembre",
  "Octubre",
  "Noviembre",
  "Diciembre",
];

export default function DollarForm() {
  const [errorMessage, setErrorMessage] = useState<string>("");
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const { dollars, getDollarInfo, saveDollarInfo, processing, error, result } =
    useDollar();
  const { filters, projectId } = useWorkspaceFilters([
    "customer",
    "project",
    "campaign",
  ]);

  const [rows, setRows] = useState<DollarRow[]>(
    months.map((month) => ({
      month,
    }))
  );

  useEffect(() => {
    setRows(
      months.map((month) => ({
        month,
      }))
    );
    if (!projectId) return;
    getDollarInfo(projectId);
  }, [getDollarInfo, projectId]);

  useEffect(() => {
    if (dollars.length === 0) {
      setRows(
        months.map((month) => ({
          month,
        }))
      );
      return;
    }

    setRows(
      months.map((month) => ({
        month,
        initialValue: dollars.find((d: DollarData) => d.month === month)
          ?.start_value,
        finalValue: dollars.find((d: DollarData) => d.month === month)
          ?.end_value,
        averageValue: dollars.find((d: DollarData) => d.month === month)
          ?.average_value,
      }))
    );
  }, [dollars]);

  useEffect(() => {
    if (result !== "") {
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
      setSuccessMessage(null);
    }
  }, [error]);

  const handleSaveDollarValues = () => {
    if (!projectId) {
      setErrorMessage("Por favor, seleccione un proyecto y campaña.");
      return;
    }

    const dollarData = rows.map((row) => ({
      month: row.month,
      start_value: row.initialValue || "0",
      end_value: row.finalValue || "0",
      average_value: row.averageValue || "0",
    }));

    saveDollarInfo(dollarData, projectId);
  };

  const handleChange = (
    index: number,
    field: "initialValue" | "finalValue",
    value: string
  ) => {
    setRows((prevRows) => {
      const updatedRows = [...prevRows];
      const row = { ...updatedRows[index], [field]: value };
      const i = parseFloat(row.initialValue || "");
      const f = parseFloat(row.finalValue || "");
      row.averageValue = !isNaN(i) && !isNaN(f) ? ((i + f) / 2).toFixed(2) : "";
      updatedRows[index] = row;
      return updatedRows;
    });
  };

  return (
    <div className="w-full mx-auto">
      <FilterBar filters={filters} />
      <div className="mt-4 p-6 w-full mx-auto bg-white rounded-lg shadow-md">
        {processing && (
          <div className="flex items-center justify-center w-full h-full">
            <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-gray-900"></div>
          </div>
        )}
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
        <h1 className="text-custom-text font-semibold text-xl leading-none">
          Valor promedio del dólar según{" "}
          <a
            href="https://www.bna.com.ar/"
            target="_blank"
            className="text-blue-600 hover:underline"
            rel="noopener noreferrer"
          >
            BNA
          </a>
        </h1>
        <div className="mt-1">
          <div className="w-full px-8 sm:px-16 md:px-32 py-4">
            <div>
              <div className="hidden sm:grid grid-cols-[1.2fr_1fr_1fr_1fr] gap-4 mb-2">
                <span className="font-medium text-right w-full block">Mes</span>
                <span className="font-medium">Valor inicial</span>
                <span className="font-medium">Valor final</span>
                <span className="font-medium">Valor promedio</span>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-[1.2fr_1fr_1fr_1fr] gap-4">
                {rows.map((row, index) => (
                  <div
                    key={index}
                    className="sm:contents border sm:border-0 p-4 sm:p-0 rounded-md sm:rounded-none mb-4 sm:mb-0 shadow-sm sm:shadow-none"
                  >
                    <div className="sm:col-span-1 flex justify-end">
                      <label className="sm:hidden text-sm text-gray-600">
                        Mes
                      </label>
                      <span>{row.month}</span>
                    </div>
                    <div className="sm:col-span-1 text-right">
                      <label className="sm:hidden text-sm text-gray-600">
                        Valor inicial $
                      </label>
                      <InputField
                        label=""
                        name={`initial-${index}`}
                        value={row.initialValue || ""}
                        type="text"
                        size="sm"
                        onChange={(e) => {
                          let value = e.target.value.replace(/,/g, ".");
                          if (/^\d*\.?\d{0,2}$/.test(value)) {
                            handleChange(index, "initialValue", value);
                          }
                        }}
                      />
                    </div>
                    <div className="sm:col-span-1 text-right">
                      <label className="sm:hidden text-sm text-gray-600">
                        Valor final $
                      </label>
                      <InputField
                        label=""
                        name={`final-${index}`}
                        value={row.finalValue || ""}
                        type="text"
                        size="sm"
                        onChange={(e) => {
                          let value = e.target.value.replace(/,/g, ".");
                          if (/^\d*\.?\d{0,2}$/.test(value)) {
                            handleChange(index, "finalValue", value);
                          }
                        }}
                      />
                    </div>
                    <div className="sm:col-span-1">
                      <label className="sm:hidden text-sm text-gray-600">
                        Valor promedio $
                      </label>
                      <InputField
                        label=""
                        name={`average-${index}`}
                        value={row.averageValue || ""}
                        type="text"
                        disabled
                        size="sm"
                        onChange={() => {}}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
      <div className="flex gap-4 my-4 justify-end">
        <Button variant="outlineGray" className="text-base font-medium">
          Cancelar
        </Button>
        <Button
          variant="success"
          disabled={processing}
          onClick={handleSaveDollarValues}
          className="text-base font-medium"
        >
          Guardar
        </Button>
      </div>
    </div>
  );
}
