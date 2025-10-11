import { InvestorContributionReportData, RowToRender } from "../../../hooks/useReporting/types.ts";

export const InvestorContributionTable = ({
                                            data,
                                            rows,
                                            tableType = "contributions",
                                          }: {
  data: InvestorContributionReportData | null;
  rows: RowToRender[];
  tableType?: "contributions" | "harvest";
}) => {
  if (!data) {
    return (
      <div className="p-4 text-sm text-gray-600 rounded-lg bg-gray-50">
        No hay datos disponibles
      </div>
    );
  }

  const investors = data.investor_headers;

  const getRowData = (key: string) => {
    if (key === "total" && tableType === "contributions") {
      return {
        total_usd: data.pre_harvest.total_usd,
        total_usd_ha: data.pre_harvest.total_us_ha,
        investors: data.pre_harvest.investors,
      };
    }

    if (tableType === "harvest") {
      const harvestRow = data.harvest.rows.find(row => row.key === key);
      return harvestRow ? {
        total_usd: harvestRow.total_usd,
        total_usd_ha: harvestRow.total_us_ha,
        investors: harvestRow.investors,
      } : null;
    }

    const contribution = data.contributions.find(c => c.key === key);
    return contribution ? {
      total_usd: contribution.total_usd,
      total_usd_ha: contribution.total_usd_ha,
      investors: contribution.investors,
    } : null;
  };

  const getInvestorValue = (rowKey: string, investorId: number) => {
    const rowData = getRowData(rowKey);
    if (!rowData) return 0;

    const investorData = rowData.investors.find(inv => inv.investor_id === investorId);
    return investorData?.amount_usd || 0;
  };

  const getInvestorSharePct = (rowKey: string, investorId: number) => {
    const rowData = getRowData(rowKey);
    if (!rowData) return 0;

    const investorData = rowData.investors.find(inv => inv.investor_id === investorId);
    return investorData?.share_pct || 0;
  };

  return (
    <div className="overflow-x-auto flex justify-between">
      <table className="bg-white w-full text-sm border border-gray-300">
        <thead>
        <tr className="h-14">
          <th className="bg-[#9CA3AF] text-white font-extralight text-left pl-4 text-xs">
            { tableType === "contributions" ? "TIPO/CLASE" : "RUBRO" }
          </th>
          <th className="bg-[#9CA3AF] text-white font-extralight text-left pl-4 text-xs">
            TOTAL INVERTIDO
          </th>
          <th className="bg-[#9CA3AF] text-white font-extralight text-left pl-4 text-xs">
            TOTAL U$ / HA
          </th>
          { investors.map((investor, index) => {
            const customBg = index % 2 === 0 ? "bg-[#6B7280] text-white" : "bg-[#4B5563] text-white";
            return (
              <th
                key={ investor.investor_id }
                className={ `${ customBg } p-2 h-14 uppercase font-extralight text-left pl-4 text-xs` }
              >
                { investor.investor_name }
                <span className="px-2 py-1 bg-[#374151] rounded-lg text-gray-300 ml-2 border border-[#6B7280]">
                    { investor.share_pct }%
                  </span>
              </th>
            );
          }) }
        </tr>
        </thead>
        <tbody>
        { rows.map(({
                      label,
                      key,
                      valueFormat,
                      classNameRows = "",
                      classNameHeader = "",
                    }, rowIndex) => {
          const rowData = getRowData(key);

          if (!rowData) {
            return null;
          }

          let baseBg = rowIndex % 2 === 1 ? "bg-[#EBF5FF]" : "bg-white";
          if (key === "total" || key === "totals") {
            baseBg = "bg-[#E5E7EB]";
          }

          const finalHeaderClasses = [
            !classNameHeader.includes("text-") && "text-black",
            !classNameHeader.includes("font-") && "font-bold",
            !classNameHeader.includes("bg-") && baseBg,
            classNameHeader,
          ].filter(Boolean).join(" ") + " pl-4 text-left border-t border-t-gray-300 h-14";

          const finalRowClasses = [
            !classNameRows.includes("text-") && "text-gray-600",
            !classNameRows.includes("bg-") && baseBg,
            !classNameRows.includes("font-") && "font-light",
            classNameRows,
          ].filter(Boolean).join(" ");

          return (
            <tr key={ key } className={ baseBg }>
              <th className={ finalHeaderClasses }>
                { label }
              </th>
              <td className={ `${ finalRowClasses } p-2 text-left border-t border-t-gray-300 pl-4` }>
                { valueFormat.totalInvested(rowData.total_usd) }
              </td>
              <td className={ `${ finalRowClasses } p-2 text-left border-t border-t-gray-300 pl-4` }>
                { valueFormat.totalPerHa(rowData.total_usd_ha) }
              </td>
              { investors.map((investor, index) => {
                const investorValue = getInvestorValue(key, investor.investor_id);
                const investorSharePct = getInvestorSharePct(key, investor.investor_id);

                let cellBg = baseBg;
                let cellTextColor = "text-gray-600";

                if (key === "total" || key === "totals") {
                  cellBg = index % 2 === 0 ? "bg-[#6B7280]" : "bg-[#4B5563]";
                  cellTextColor = "text-white";
                }

                return (
                  <td
                    key={ investor.investor_id }
                    className={ `${ cellBg } ${ cellTextColor } ${ (key === "total" || key === "totals") ? "font-bold" : "font-light" } p-2 text-left border-t border-t-gray-300 pl-4` }
                  >
                    { valueFormat.investor(investorValue, investorSharePct) }
                    {
                      (key === "total" || key === "totals") && (
                        <span
                          className="px-2 py-1 bg-[#374151] rounded-lg text-gray-300 ml-2 border border-[#6B7280] text-xs font-normal">
                          { investorSharePct }%
                        </span>
                      )
                    }
                  </td>
                );
              }) }
            </tr>
          );
        }) }

        { tableType === "contributions" && data.comparison && (
          <>
            <tr className="h-14">
              <td></td>
              <td></td>
              <td className="p-2 font-bold text-right text-black">
                Aporte acordado
              </td>
              { data.comparison.map((investor) => (
                <td key={ investor.investor_id } className="p-2 text-left pl-4 text-gray-700">
                  u${ investor.agreed_usd } - { investor.agreed_share_pct }%
                </td>
              )) }
            </tr>
            <tr className="h-14">
              <td></td>
              <td></td>
              <td className="p-2 font-bold text-right text-black border-t">
                Ajuste de aporte
              </td>
              { data.comparison.map((investor) => (
                <td key={ investor.investor_id } className="p-2 text-left pl-4 text-gray-700 border-t">
                  u${ investor.adjustment_usd }
                </td>
              )) }
            </tr>
          </>
        ) }

        { tableType === "harvest" && data.harvest && (
          <>
            <tr className="h-14">
              <td></td>
              <td></td>
              <td className="p-2 font-bold text-right text-black">
                Pago acordado
              </td>
              { data.harvest.footer_payment_agreed.map((investor) => (
                <td key={ investor.investor_id } className="p-2 text-left pl-4 text-gray-700">
                  u${ investor.amount_usd } - { investor.share_pct }%
                </td>
              )) }
            </tr>
            <tr className="h-14">
              <td></td>
              <td></td>
              <td className="p-2 font-bold text-right text-black border-t">
                Ajuste de pago
              </td>
              { data.harvest.footer_payment_adjustment.map((investor) => (
                <td key={ investor.investor_id } className="p-2 text-left pl-4 text-gray-700 border-t">
                  u${ investor.amount_usd }
                </td>
              )) }
            </tr>
          </>
        ) }
        </tbody>
      </table>
    </div>
  );
};