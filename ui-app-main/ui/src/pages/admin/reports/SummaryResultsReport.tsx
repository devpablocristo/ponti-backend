import { useEffect, useState } from "react";
import { ExternalLink, LoaderCircle, SquareArrowOutUpRight } from "lucide-react";
import FilterBar from "../../../layout/FilterBar/FilterBar";
import { useWorkspaceFilters } from "../../../hooks/useWorkspaceFilters";
import SelectField from "../../../components/Input/SelectField";
import Button from "../../../components/Button/Button";
import useReporting from "../../../hooks/useReporting";
import { SummaryResultsTable } from "./SummaryResultsTable.tsx";
import { RowToRender } from "../../../hooks/useReporting/types.ts";
import { usePDF } from "react-to-pdf";

const rowsToRender: RowToRender[] = [
  {
    label: "Superficie",
    key: "surface_ha",
    valueFormat: {
      crop: (value) => `${ value } Has`,
    },
    classNameHeader: "text-black font-semibold h-14",
  },
  {
    label: "Ingreso Neto",
    key: "net_income_usd",
    valueFormat: {
      crop: (value) => `u$ ${ value }`,
    },
    classNameRows: "text-white font-semibold bg-[#9CA3AF]",
    classNameHeader: "text-white font-semibold bg-[#9CA3AF] h-14",
  },
  {
    label: "Costos Directos",
    key: "direct_costs_usd",
    valueFormat: {
      crop: (value) => `u$ ${ value }`,
    },
    classNameHeader: "text-black font-semibold h-14",
  },
  {
    label: "Arriendo",
    key: "rent_usd",
    valueFormat: {
      crop: (value) => `u$ ${ value }`,
    },
    classNameHeader: "text-black font-semibold h-14",
  },
  {
    label: "Estructura",
    key: "structure_usd",
    valueFormat: {
      crop: (value) => `u$ ${ value }`,
    },
    classNameHeader: "text-black font-semibold h-14",
  },
  {
    label: "Total invertido",
    key: "total_invested_usd",
    valueFormat: {
      crop: (value) => `u$ ${ value }`,
    },
    classNameRows: "text-black font-bold bg-[#FBD5D5]",
    classNameHeader: "text-black font-semibold bg-[#FBD5D5] h-14",
  },
  {
    label: "Resultado operativo",
    key: "operating_result_usd",
    valueFormat: {
      crop: (value) => `u$ ${ value }`,
    },
    classNameRows: "text-white font-bold bg-black",
    classNameHeader: "text-white font-semibold bg-black h-14",
    showIndicator: true,
  },
  {
    label: "Renta del cultivo",
    key: "crop_return_pct",
    valueFormat: {
      crop: (value) => `${ value }%`,
    },
    classNameRows: "text-black font-bold",
    classNameHeader: "text-black font-semibold h-14",
    showIndicator: true,
  },
];

export function SummaryResultsReport() {
  const [selectedCrop, setSelectedCrop] = useState<string>("0");

  const { filters, projectId, selectedCampaignId, loading } =
    useWorkspaceFilters(["project", "campaign"]);

  const {
    summaryResultsReportingData: reportingData,
    processing,
    error,
    getSummaryResultsReportingData,
  } = useReporting();

  const buildQueryParams = () => {
    const params: Record<string, string> = {};

    if (projectId) {
      params.project_id = String(projectId);
    }
    if (selectedCampaignId) {
      params.campaign_id = String(selectedCampaignId);
    }

    return new URLSearchParams(params).toString();
  };

  useEffect(() => {
    getSummaryResultsReportingData(buildQueryParams());
  }, []);

  const filteredData = selectedCrop === "0"
    ? reportingData
    : reportingData ? {
      ...reportingData,
      crops: reportingData.crops.filter((crop) => crop.crop_id.toString() === selectedCrop),
    } : null;

  const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
  const { toPDF, targetRef } = usePDF({ filename: `informe-resumen-resultados-${ timestamp }.pdf` });

  return (
    <div className="relative">
      { (loading.projects || loading.campaigns || processing) && (
        <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
          <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin"/>
        </div>
      ) }

      <FilterBar
        filters={ filters }
        actions={ [
          {
            label: "Generar informe",
            variant: "success",
            disabled: processing,
            onClick: () => getSummaryResultsReportingData(buildQueryParams()),
          },
          {
            label: "Exportar informe",
            variant: "outlineGreen",
            icon: <SquareArrowOutUpRight className="h-3.5 w-3.5 stroke-[3px]"/>,
            disabled: processing,
            onClick: toPDF,
          },
        ] }
      />

      { error && (
        <div
          className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50"
          role="alert"
        >
          <span className="font-medium">{ error }</span>
        </div>
      ) }

      { !error && (
        <>
          <div className="border py-6 px-2" ref={ targetRef }>
            <div
              className="border-b mb-5"
              style={ { borderColor: "#D1D5DB" } }
            />
            <div className="flex mb-3 items-center">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">
                Resumen de resultados
              </h2>
              <div className="mb-4 ml-8 w-[200px]">
                <SelectField
                  label="Cultivo"
                  name="summaryView"
                  value={ selectedCrop }
                  onChange={ (e) => setSelectedCrop(e.target.value) }
                  options={ reportingData ? [
                    { id: 0, name: "Todos" },
                    ...reportingData.crops.map((crop) => ({
                      id: crop.crop_id,
                      name: crop.crop_name,
                    })),
                  ] : [] }
                  size="sm"
                  fullWidth
                />
              </div>
            </div>
            <div
              className="border-b mb-5"
              style={ { borderColor: "#D1D5DB" } }
            />
            <SummaryResultsTable
              data={ filteredData }
              rows={ rowsToRender }
            />
          </div>
          <div className="flex justify-end mr-2 mb-6">
            <Button
              variant="outlineGreen"
              className="gap-2"
              onClick={ toPDF }
              disabled={ processing }
            >
              <ExternalLink></ExternalLink>
              Exportar informe
            </Button>
          </div>
        </>
      ) }
    </div>
  );
}

export default SummaryResultsReport;