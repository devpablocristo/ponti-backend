import { DashboardData } from "../../../hooks/useDashboard/types";
import { formatNumberAr } from "../utils";

interface ManagementBalanceTableProps {
  dashboard: DashboardData | null;
}

export default function ManagementBalanceTable({dashboard}: ManagementBalanceTableProps) {
  if (!dashboard || !dashboard.management_balance) {
    return (
      <div className="bg-white rounded-xl border p-4 w-full">
        <h3 className="font-medium text-[#020617] font-sans mb-4 text-xl">
          Balance de gestión
        </h3>
        <div className="p-4 text-sm text-gray-600 rounded-lg bg-gray-50">
          No data available
        </div>
      </div>
    );
  }

  const { management_balance } = dashboard;
  const { items } = management_balance;

  // Filter items by category - direct costs first
  const directCostItems = items.filter(item =>
    ['SEED', 'SUPPLIES', 'LABORS'].includes(item.category)
  );

  const otherItems = items.filter(item =>
    ['LEASE', 'ADMIN'].includes(item.category)
  );

  const formatCurrency = (value: string | undefined) => {
    if (!value) return "---";
    return `u$ ${formatNumberAr(value)}`;
  };

  return (
    <div className="bg-white rounded-xl border p-4 w-full">
      <h3 className="font-medium text-[#020617] font-sans mb-4 text-xl">
        Balance de gestión
      </h3>

      <div className="grid grid-cols-4 text-sm font-semibold mb-1">
        <div></div>
        <div className="ml-[16px]">
          <span className="bg-[#F98080] text-gray-900 py-2 px-2 rounded h-[43px] w-[111px] text-center flex justify-center items-center text-sm">
            Ejecutados
          </span>
        </div>
        <div className="ml-[16px]">
          <span className="bg-[#FBD5D5] text-gray-900 py-2 px-2 rounded text-sm h-[43px] w-[111px] text-center flex justify-center items-center">
            Invertidos
          </span>
        </div>
        <div className="ml-[16px]">
          <span className="bg-[#E5E7EB] text-gray-900 py-2 px-2 rounded text-sm h-[43px] w-[111px] text-center flex justify-center items-center">
            Stock
          </span>
        </div>
      </div>

      <div className="grid grid-cols-4 text-sm px-2 bg-gray-100 font-semibold text-gray-900 h-[45px] rounded-lg">
        <div className="flex items-center">Costos directos</div>
        <div className="text-black bg-[#F98080] rounded text-center flex items-center justify-center w-[111px] ml-[12px]">
          {formatCurrency(
            directCostItems.reduce((sum, item) => sum + parseFloat(item.executed_usd || "0"), 0).toString()
          )}
        </div>
        <div className="text-[#F05252] text-center flex items-center justify-center w-[111px] bg-[#FBD5D5] rounded ml-[17px]">
          {formatCurrency(
            directCostItems.reduce((sum, item) => sum + parseFloat(item.invested_usd || "0"), 0).toString()
          )}
        </div>
        <div className="text-center flex items-center justify-center w-[112px] bg-[#E5E7EB] rounded ml-[20px]">
          {formatCurrency(
            directCostItems.reduce((sum, item) => sum + parseFloat(item.stock_usd || "0"), 0).toString()
          )}
        </div>
      </div>

      {directCostItems
        .sort((a, b) => a.order - b.order)
        .map((item, idx) => (
          <div
            key={idx}
            className="grid grid-cols-4 text-sm pl-2 text-gray-800 h-[45px] border-b border-[#F3F4F6]"
          >
            <div className="flex items-center">{item.label}</div>
            <div className="text-black text-center flex items-center justify-center ml-2">
              {formatCurrency(item.executed_usd)}
            </div>
            <div className="text-black bg-[#FDF2F2] h-[45px] w-[111px] text-center flex items-center justify-center ml-[13px]">
              {formatCurrency(item.invested_usd)}
            </div>
            <div className="text-center bg-[#F9FAFB] flex items-center justify-center ml-[14px]">
              {formatCurrency(item.stock_usd)}
            </div>
          </div>
        ))}

      {otherItems
        .sort((a, b) => a.order - b.order)
        .map((item, idx) => (
          <div
            key={idx}
            className="grid grid-cols-4 text-sm px-2 bg-gray-100 font-semibold text-gray-900 h-[45px] rounded-lg mb-0.5"
          >
            <div className="flex items-center">{item.label}</div>
            <div className="text-black text-center flex items-center justify-center h-[45px] w-[111px] ml-2">
              {formatCurrency(item.executed_usd)}
            </div>
            <div className="text-[#F05252] text-center flex items-center justify-center ml-5">
              {formatCurrency(item.invested_usd)}
            </div>
            <div className="text-center flex items-center justify-center ml-5">
              {formatCurrency(item.stock_usd)}
            </div>
          </div>
        ))}

      {/* Totals Row */}
      {/*<div className="grid grid-cols-4 text-sm py-2 px-2 bg-gray-200 font-bold text-gray-900 border-t-2 border-gray-300 h-[45px]">*/}
      {/*  <div>Total</div>*/}
      {/*  <div className="text-red-500 text-center">*/}
      {/*    {formatCurrency(totals.executed_usd)}*/}
      {/*  </div>*/}
      {/*  <div className="text-blue-600 text-center">*/}
      {/*    {formatCurrency(totals.invested_usd)}*/}
      {/*  </div>*/}
      {/*  <div className="text-center">*/}
      {/*    {formatCurrency(totals.stock_usd)}*/}
      {/*  </div>*/}
      {/*</div>*/}
    </div>
  );
}