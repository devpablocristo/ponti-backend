import { useEffect } from "react";
import { LoaderCircle, ExternalLink, SquareArrowOutUpRight } from "lucide-react";
import FilterBar from "../../../layout/FilterBar/FilterBar";
import { useWorkspaceFilters } from "../../../hooks/useWorkspaceFilters";
import Button from "../../../components/Button/Button";
import InputField from "../../../components/Input/InputField.tsx";
import useReporting from "../../../hooks/useReporting";
import { InvestorContributionTable } from "./InvestorContributionTable.tsx";
import { RowToRender } from "../../../hooks/useReporting/types.ts";
import { usePDF } from "react-to-pdf";
import { formatNumberAr } from "../utils.ts";

const contributionRowsToRender: RowToRender[] = [
  {
    label: "Agroquímicos",
    key: "agrochemicals",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value, percentage) => `u$${ formatNumberAr(value) } - ${ percentage }%`,
    },
  },
  {
    label: "Semilla",
    key: "seeds",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value, percentage) => `u$${ formatNumberAr(value) } - ${ percentage }%`,
    },
  },
  {
    label: "Labores grales",
    key: "general_labors",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value, percentage) => `u$${ formatNumberAr(value) } - ${ percentage }%`,
    },
  },
  {
    label: "Siembra",
    key: "sowing",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value, percentage) => `u$${ formatNumberAr(value) } - ${ percentage }%`,
    },
  },
  {
    label: "Riego",
    key: "irrigation",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value, percentage) => `u$${ formatNumberAr(value) } - ${ percentage }%`,
    },
  },
  {
    label: "Arriendo",
    key: "capitalizable_lease",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value, percentage) => `u$${ formatNumberAr(value) } - ${ percentage }%`,
    },
  },
  {
    label: "Administración y estructura",
    key: "administration_structure",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value, percentage) => `u$${ formatNumberAr(value) } - ${ percentage }%`,
    },
  },
  {
    label: "Totales",
    key: "total",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value) => `u$${ formatNumberAr(value) }`,
    },
    classNameRows: "text-black bg-[#E5E7EB] font-bold",
    classNameHeader: "text-black bg-[#E5E7EB] font-bold",
  },
];

const harvestRowsToRender: RowToRender[] = [
  {
    label: "Cosecha",
    key: "harvest",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value, percentage) => `u$${ formatNumberAr(value) } - ${ percentage }%`,
    },
  },
  {
    label: "Totales",
    key: "totals",
    valueFormat: {
      totalInvested: (value) => `u$${ formatNumberAr(value) }`,
      totalPerHa: (value) => `${ formatNumberAr(value) } u$/Ha`,
      investor: (value) => `u$${ formatNumberAr(value) }`,
    },
    classNameRows: "text-black bg-[#E5E7EB] font-bold",
    classNameHeader: "text-black bg-[#E5E7EB] font-bold",
  },
];

export function InvestorContributionReport() {
  const { filters, projectId, selectedCampaignId, loading } =
    useWorkspaceFilters(["project", "campaign"]);

  const {
    investorContributionReportingData: reportingData,
    processing,
    error,
    getInvestorContributionReportingData,
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
    getInvestorContributionReportingData(buildQueryParams());
  }, []);

  const timestamp = new Date().toISOString().slice(0, 19).replace(/:/g, '-');
  const { toPDF, targetRef } = usePDF({ filename: `informe-aporte-inversor-${ timestamp }.pdf` });

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
            onClick: () => getInvestorContributionReportingData(buildQueryParams()),
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
          <div ref={ targetRef }>
            <div className="rounded-xl border py-6 px-2">
              <div className="border-b mb-5" style={ { borderColor: "#D1D5DB" } }/>
              <div className="flex justify-between items-center gap-8 mb-4">
                <h2 className="text-2xl font-semibold text-gray-900 w-4/12">
                  Aporte por inversor
                </h2>
                <div className="flex gap-8">
                  <InputField
                    label="Superficie"
                    name="surface"
                    size="sm"
                    value=""
                    placeholder={ `${ reportingData?.general.surface_total_ha !== undefined ? formatNumberAr(reportingData.general.surface_total_ha) : "" } Has` }
                    className="flex items-center w-[175px] gap-2 text-nowrap"
                    inputClassName="h-12 -mt-2"
                    onChange={ () => {
                    } }
                    disabled
                  />
                  <InputField
                    label="Arriendo / Ha"
                    name="rental-per-ha"
                    size="sm"
                    value=""
                    placeholder={ `u$ ${ reportingData?.general.lease_fixed_usd !== undefined ? formatNumberAr(reportingData.general.lease_fixed_usd) : "" }` }
                    className="flex items-center w-[200px] gap-2 text-nowrap"
                    inputClassName="h-12 -mt-2"
                    onChange={ () => {
                    } }
                    disabled
                  />
                  <InputField
                    label="Admin. proyecto / Ha"
                    name="admin-project-per-ha"
                    size="sm"
                    value=""
                    placeholder={ `u$ ${ reportingData?.general.admin_per_ha_usd !== undefined ? formatNumberAr(reportingData.general.admin_per_ha_usd) : "" }` }
                    className="flex items-center w-[250px] gap-2 text-nowrap"
                    inputClassName="h-12 -mt-2"
                    onChange={ () => {
                    } }
                    disabled
                  />
                </div>
              </div>
              <div
                className="border-b mt-2 mb-6"
                style={ { borderColor: "#D1D5DB" } }
              />
              <h2 className="font-semibold text-xl mb-4">Aportes pre cosecha</h2>
              { processing ? (
                <div className="flex items-center justify-center h-48">
                  <LoaderCircle className="w-8 h-8 text-blue-500 animate-spin"/>
                </div>
              ) : (
                <InvestorContributionTable
                  data={ reportingData }
                  rows={ contributionRowsToRender }
                  tableType="contributions"
                />
              ) }
            </div>

            <div className="rounded-xl border py-6 px-2 mb-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">
                Pagos de cosecha
              </h2>
              { processing ? (
                <div className="flex items-center justify-center h-48">
                  <LoaderCircle className="w-8 h-8 text-blue-500 animate-spin"/>
                </div>
              ) : (
                <InvestorContributionTable
                  data={ reportingData }
                  rows={ harvestRowsToRender }
                  tableType="harvest"
                />
              ) }
            </div>
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

export default InvestorContributionReport;