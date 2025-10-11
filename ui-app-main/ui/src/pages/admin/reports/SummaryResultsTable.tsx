import { RowToRender, SummaryResultsReportData } from "../../../hooks/useReporting/types.ts";
import { cropColors } from "../colors";

const CropBadge = ({ cropName }: { cropName: string }) => (
  <span
    className={ `px-2 py-1 text-xs font-medium rounded-md ${ cropColors[cropName] || "bg-[#E5E7EB] text-[#000000] border border-[#000000]" }` }>
    { cropName }
  </span>
);

export const SummaryResultsTable = ({
                                      data,
                                      rows,
                                    }: {
  data: SummaryResultsReportData | null;
  rows?: RowToRender[];
}) => {
  if (!data) {
    return (
      <div className="p-4 text-sm text-gray-600 rounded-lg bg-gray-50">
        No hay datos disponibles
      </div>
    );
  }

  const getValue = (rowKey: string, cropId: number) => {
    const crop = data.crops.find(c => c.crop_id === cropId);
    if (!crop) return 0;

    return parseFloat(crop[rowKey as keyof typeof crop] as string) || 0;
  };


  const indicatorBackgroundColor = (value: string | number) => {
    const numericValue = Number(value);
    if (numericValue > 0) {
      return numericValue < 6 ? "bg-[#FACA15]" : "bg-[#31C48D]";
    } else {
      return "bg-[#F98080]";
    }
  }

  return (
    <div className="overflow-x-auto flex justify-between">
      <table
        className="w-11/12 text-sm text-nowrap bg-white border border-gray-300">
        <thead>
        <tr className="h-14">
          <th></th>
          { data.crops.map((crop) => (
            <>
              <th className="w-2"></th>
              <th className="p-2" key={ crop.crop_id }>
                <CropBadge cropName={ crop.crop_name }/>
              </th>
            </>
          )) }
        </tr>
        </thead>
        <tbody>
        { rows?.map(({
                       label,
                       key,
                       valueFormat,
                       classNameRows = "",
                       classNameHeader = "",
                       showIndicator,
                     }, rowIndex) => (
          <tr key={ key } className={ classNameHeader }>
            <th className={ [
              !classNameHeader.includes("text-") && "text-gray-600",
              !classNameHeader.includes("font-") && "font-light",
              classNameHeader,
            ].filter(Boolean).join(" ") + " p-2 text-left border-t border-t-gray-300" }>
              { label }
            </th>
            { data.crops.map((crop, index) => {
              const baseBg = rowIndex % 2 === 1 ? "bg-white" : "bg-[#EBF5FF]";
              const finalRowClasses = [
                !classNameRows.includes("text-") && "text-gray-600",
                !classNameRows.includes("bg-") && baseBg,
                !classNameRows.includes("font-") && "font-light",
                classNameRows,
              ].filter(Boolean).join(" ");

              const rowValue = getValue(key, crop.crop_id);

              return (
                <>
                  <td className="w-2 bg-white"></td>
                  { showIndicator ? (
                    <td key={ index }
                        className={ `${ finalRowClasses } p-1 text-center flex justify-center items-center h-14 border-t border-t-gray-300` }>
                      { valueFormat.crop(rowValue) }
                      <div className="relative h-6 w-6 ml-1">
                        <div
                          className={ `absolute inset-0 rounded-full opacity-30 ${ indicatorBackgroundColor(rowValue) }` }></div>
                        <div
                          className={ `absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-3 h-3 rounded-full ${ indicatorBackgroundColor(rowValue) }` }>
                        </div>
                      </div>
                    </td>
                  ) : (
                    <td key={ index }
                        className={ `${ finalRowClasses } p-1 text-center border-t h-3 border-t-gray-300` }>
                      { valueFormat.crop(rowValue) }
                    </td>
                  ) }
                </>
              );
            }) }
          </tr>
        )) }
        </tbody>
      </table>
      <div className="flex gap-4 ml-4 justify-around">
        <table className="text-sm border-2 border-gray-400 text-nowrap">
          <thead>
          <tr>
            <th className="h-14 text-xs text-[#6B7280] bg-white text-nowrap px-2">GRAL CAMPOS</th>
          </tr>
          </thead>
          <tbody>
          <tr>
            <td className="p-1 text-center text-[#6B7280] font-bold h-14 bg-[#D1D5DB]">
              { data.totals.total_surface_ha } Has
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center text-white font-bold h-14 bg-[#9CA3AF]">
              u$ { data.totals.total_net_income_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 bg-[#D1D5DB]">
              u$ { data.totals.total_direct_costs_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 bg-white">
              u$ { data.totals.total_rent_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 bg-[#D1D5DB]">
              u$ { data.totals.total_structure_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 bg-[#FBD5D5]">
              u$ { data.totals.total_invested_project_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center h-14 bg-black text-white font-bold flex justify-center items-center">
              u$ { data.totals.total_operating_result_usd }
              <div className="relative w-6 h-6 ml-1">
                <div
                  className={ `absolute inset-0 rounded-full opacity-30 ${ indicatorBackgroundColor(data.totals.total_operating_result_usd) }` }></div>
                <div
                  className={ `absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-3 h-3 rounded-full ${ indicatorBackgroundColor(data.totals.total_operating_result_usd) }` }>
                </div>
              </div>
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 flex justify-center items-center bg-white">
              { parseFloat(data.totals.project_return_pct) }%
              <div className="relative h-6 w-6 ml-1">
                <div
                  className={ `absolute inset-0 rounded-full opacity-30 ${ indicatorBackgroundColor(data.totals.project_return_pct) }` }></div>
                <div
                  className={ `absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-3 h-3 rounded-full ${ indicatorBackgroundColor(data.totals.project_return_pct) }` }>
                </div>
              </div>
            </td>
          </tr>
          </tbody>
        </table>
        <table className="text-sm border-2 border-gray-400 text-nowrap">
          <thead>
          <tr>
            <>
              <th className="h-14 text-xs text-[#6B7280] bg-white text-nowrap px-2">GRAL CULTIVOS</th>
            </>
          </tr>
          </thead>
          <tbody>
          <tr>
            <td className="p-1 text-center text-[#6B7280] font-bold h-14 bg-[#D1D5DB]">
              { data.general_crops.total_surface_ha } Has
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center text-white font-bold h-14 bg-[#9CA3AF]">
              u$ { data.general_crops.total_net_income_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 bg-[#D1D5DB]">
              u$ { data.general_crops.total_direct_costs_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 bg-white">
              u$ { data.general_crops.total_rent_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 bg-[#D1D5DB]">
              u$ { data.general_crops.total_structure_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 bg-[#FBD5D5]">
              u$ { data.general_crops.total_invested_project_usd }
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center h-14 bg-black text-white font-bold flex justify-center items-center">
              u$ { data.general_crops.total_operating_result_usd }
              <div className="relative h-6 w-6 ml-1">
                <div
                  className={ `absolute inset-0 rounded-full opacity-30 ${ indicatorBackgroundColor(data.general_crops.total_operating_result_usd) }` }></div>
                <div
                  className={ `absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-3 h-3 rounded-full ${ indicatorBackgroundColor(data.general_crops.total_operating_result_usd) }` }>
                </div>
              </div>
            </td>
          </tr>
          <tr>
            <td className="p-1 text-center font-bold h-14 flex justify-center items-center bg-white">
              { parseFloat(data.general_crops.project_return_pct) }%
              <div className="relative h-6 w-6 ml-1">
                <div
                  className={ `absolute inset-0 rounded-full opacity-30 ${ indicatorBackgroundColor(data.general_crops.project_return_pct) }` }></div>
                <div
                  className={ `absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-3 h-3 rounded-full ${ indicatorBackgroundColor(data.general_crops.project_return_pct) }` }>
                </div>
              </div>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
};