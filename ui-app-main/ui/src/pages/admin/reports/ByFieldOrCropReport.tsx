import { useEffect, useState } from "react";
import { ExternalLink, LoaderCircle, SquareArrowOutUpRight } from "lucide-react";
import FilterBar from "../../../layout/FilterBar/FilterBar";
import { useWorkspaceFilters } from "../../../hooks/useWorkspaceFilters";
import SelectField from "../../../components/Input/SelectField";
import Button from "../../../components/Button/Button";
import { ByFieldOrCropTable } from "./ByFieldOrCropTable.tsx";
import useReporting from "../../../hooks/useReporting";
import { RowToRender } from "../../../hooks/useReporting/types.ts";
import { usePDF } from 'react-to-pdf';
import { formatNumberAr } from "../utils.ts";

const rowsToRender: RowToRender[] = [
  {
    label: "Superficie", key: "surface", valueFormat: {
      crop: (value) => `${ value } Has`,
    },
  },
  {
    label: "Producción", key: "production", valueFormat: {
      crop: (value) => `${ value } Tn`,
    },
  },
  {
    label: "Rendimiento",
    key: "yield",
    valueFormat: {
      crop: (value) => `${ value } Tn/Has`,
    },
    classNameRows: "text-black font-bold h-14",
    classNameHeader: "text-black font-bold h-14",
  },
  {
    label: "Precio bruto",
    key: "gross_price",
    valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Tn`,
    },
    classNameRows: "text-black font-bold",
    classNameHeader: "text-black font-bold",
  },
  {
    label: "Flete", key: "freight_cost", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Tn`,
    },
  },
  {
    label: "Gasto Comercial", key: "commercial_cost", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Tn`,
    },
  },
  {
    label: "Precio Neto",
    key: "net_price",
    valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Tn`,
    },
    classNameRows: "text-black font-bold",
    classNameHeader: "text-black font-bold",
  },
  {
    label: "Ingreso Neto",
    key: "net_income",
    valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Tn`,
    },
    classNameRows: "text-white bg-[#9CA3AF] font-bold",
    classNameHeader: "text-white bg-[#9CA3AF] font-bold",
  },
  {
    label: "Siembra", key: "labor_siembra", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Pulverizacion", key: "labor_pulverizacion", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Riego", key: "labor_riego", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Cosecha", key: "labor_cosecha", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Otras Labores", key: "labor_otras", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Total Labores",
    key: "labors_cost",
    valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
    classNameRows: "text-black bg-[#E5E7EB] font-bold h-12",
    classNameHeader: "text-black bg-[#E5E7EB] font-bold h-12",
  },
  {
    label: "Semillas", key: "supply_semillas", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Curasemillas", key: "supply_curasemillas", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Herbicidas", key: "supply_herbicidas", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Insecticidas", key: "supply_insecticidas", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Coadyuvantes", key: "supply_coadyuvantes", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Fertilizantes", key: "supply_fertilizantes", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Fungicidas", key: "supply_fungicidas", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Otros insumos", key: "supply_otros", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Total Insumos",
    key: "supplies_cost",
    valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
    classNameRows: "text-black bg-[#E5E7EB] font-bold",
    classNameHeader: "text-black bg-gray-[#E5E7EB] font-bold",
  },
  {
    label: "Total Costos Directos",
    key: "total_direct_costs",
    valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
    classNameRows: "text-black bg-[#FBD5D5] font-bold",
    classNameHeader: "text-black bg-[#FBD5D5] font-bold",
  },
  {
    label: "Margen bruto",
    key: "gross_margin",
    valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
    classNameRows: "text-white bg-[#9CA3AF] font-bold",
    classNameHeader: "text-white bg-[#9CA3AF] font-bold",
  },
  {
    label: "Arriendo", key: "lease", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Administración y Estructura", key: "admin", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
  },
  {
    label: "Resultado operativo",
    key: "operating_result",
    valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
    classNameRows: "text-white bg-black font-bold",
    classNameHeader: "text-white bg-black font-bold",
    showIndicator: true,
  },
  {
    label: "Total invertido",
    key: "total_invested",
    valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Ha`,
    },
    classNameRows: "text-black bg-[#FBD5D5] font-bold",
    classNameHeader: "text-black bg-[#FBD5D5] font-bold",
  },
  {
    label: "Renta del cultivo",
    key: "return_pct",
    valueFormat: {
      crop: (value) => `${ value }%`,
    },
    showIndicator: true,
  },
  {
    label: "Rinde de indiferencia total", key: "indifference_yield", valueFormat: {
      crop: (value) => `${ value } Tn/Ha`,
    },
  },
  {
    label: "Precio de indiferencia", key: "indifference_price", valueFormat: {
      crop: (value) => `${ formatNumberAr(value) } u$/Tn`,
    },
  },
];

export function ByFieldOrCropReport() {
  const [selectedField, setSelectedField] = useState<string>("0");
  const [selectedCrop, setSelectedCrop] = useState<string>("0");

  const { filters, projectId, selectedCampaignId, loading } =
    useWorkspaceFilters(["project", "campaign"]);

  const { fieldCropReportingData: reportingData, processing, error, getFieldCropReportingData } = useReporting();

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
    getFieldCropReportingData(buildQueryParams());
  }, []);

  const filteredData = () => {
    const filteredByField = selectedField === "0"
      ? reportingData
      : reportingData ? {
        ...reportingData,
        columns: reportingData.columns.filter((col) => col.field_id.toString() === selectedField),
      } : null;

    return selectedCrop === "0"
      ? filteredByField
      : filteredByField ? {
        ...filteredByField,
        columns: filteredByField.columns.filter((col) => col.crop_id.toString() === selectedCrop),
      } : null;
  }

  const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
  const { toPDF, targetRef } = usePDF({ filename: `informe-campo-cultivo-${ timestamp }.pdf` });

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
            onClick: () => getFieldCropReportingData(buildQueryParams()),
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
          <div className="rounded-xl border py-6 px-2" ref={ targetRef }>
            <div className="border-b mb-4" style={ { borderColor: "#D1D5DB" } }/>
            <div className="flex items-center gap-8 mb-4">
              <h2 className="text-xl font-semibold text-gray-900">
                Informe por Campo/Cultivo
              </h2>
              <div className="flex gap-4">
                <div className="w-48">
                  <SelectField
                    label="Campo"
                    name="field"
                    value={ selectedField }
                    onChange={ (e) => setSelectedField(e.target.value) }
                    options={ reportingData ? reportingData.columns.reduce((acc, crop) => {
                        if (acc.findIndex(f => f.id === crop.field_id) === -1) {
                          acc.push({ id: crop.field_id, name: crop.field_name });
                        }
                        return acc;
                      }, [{ id: 0, name: "Todos" }])
                      : [] }
                    size="sm"
                    fullWidth
                  />
                </div>
                <div className="w-48">
                  <SelectField
                    label="Cultivo"
                    name="crop"
                    value={ selectedCrop }
                    onChange={ (e) => setSelectedCrop(e.target.value) }
                    options={ reportingData ? [
                      { id: 0, name: "Todos" },
                      ...reportingData.columns.map((crop) => ({
                        id: crop.crop_id,
                        name: crop.crop_name,
                      })),
                    ] : [] }
                    size="sm"
                    fullWidth
                  />
                </div>
              </div>
            </div>
            <div
              className="border-b mt-2 mb-6"
              style={ { borderColor: "#D1D5DB" } }
            />

            { processing ? (
              <div className="flex items-center justify-center h-48">
                <LoaderCircle className="w-8 h-8 text-blue-500 animate-spin"/>
              </div>
            ) : (
              <ByFieldOrCropTable
                data={ filteredData() }
                rows={ rowsToRender }
              />
            ) }
          </div>
          <div className="flex justify-end mr-2 mb-6">
            <Button
              variant="outlineGreen"
              className="gap-2"
              onClick={ toPDF }
              disabled={ processing }
            >
              <ExternalLink/>
              Exportar informe
            </Button>
          </div>
        </>
      ) }
    </div>
  );
}

export default ByFieldOrCropReport;