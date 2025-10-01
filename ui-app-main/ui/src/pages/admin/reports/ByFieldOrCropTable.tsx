import { FieldCropReportData, RowToRender } from "../../../hooks/useReporting/types.ts";
import { cropColors } from "../colors";

const CropBadge = ({ cropName }: { cropName: string }) => (
  <span
    className={ `px-2 py-1 text-xs font-medium rounded-md ${ cropColors[cropName] || "bg-[#E5E7EB] text-[#000000] border border-[#000000]" }` }>
    { cropName }
  </span>
);

export const ByFieldOrCropTable = ({
                                     data,
                                     rows,
                                   }: {
  data: FieldCropReportData | null;
  rows?: RowToRender[];
}) => {
  if (!data) {
    return (
      <div className="p-4 text-sm text-gray-600 rounded-lg bg-gray-50">
        No hay datos disponibles
      </div>
    );
  }

  const getValue = (rowKey: string, columnId: string) => {
    const row = data.rows.find(r => r.key === rowKey);
    if (!row || !row.values[columnId]) {
      return 0;
    }
    return parseFloat(row.values[columnId].number) || 0;
  };

  const indicatorBackgroundColor = (value: number) => {
    if (value > 0) {
      return value < 6 ? "bg-[#FACA15]" : "bg-[#31C48D]";
    } else {
      return "bg-[#F98080]";
    }
  }

  return (
    <div className="overflow-x-auto flex justify-between">
      <table className="w-full text-sm bg-white border border-gray-300">
        <thead>
        <tr className="h-14">
          <th></th>
          { data.columns.map((row) => (
            <>
              <th className="w-2"></th>
              <th className="p-2">
                <p className="mb-2 uppercase font-medium">{ row.field_name }</p>
                <hr className="mb-3"/>
                <CropBadge cropName={ row.crop_name }/>
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
            ].filter(Boolean).join(" ") + " p-2 text-left w-1/5 border-t border-t-gray-300" }>
              { label }
            </th>
            { data.columns.map((column, index) => {
              const baseBg = rowIndex % 2 === 1 ? "bg-white" : "bg-[#EBF5FF]";
              const finalRowClasses = [
                !classNameRows.includes("text-") && "text-gray-600",
                !classNameRows.includes("bg-") && baseBg,
                !classNameRows.includes("font-") && "font-light",
                classNameRows,
              ].filter(Boolean).join(" ");

              const rowValue = getValue(key, column.id);

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
    </div>
  );
};