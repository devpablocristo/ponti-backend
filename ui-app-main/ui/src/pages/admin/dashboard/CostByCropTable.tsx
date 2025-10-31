import { DashboardData } from "../../../hooks/useDashboard/types";
import { cropColors } from "../colors.ts";

interface CostByCropTableProps {
  dashboard: DashboardData | null;
}

export function CostByCropTable({ dashboard }: CostByCropTableProps) {
  if (!dashboard || !dashboard.crop_incidence.items?.length) {
    return (
      <div className="bg-white rounded-xl border p-4 w-full">
        <h3 className="font-semibold text-[#020617] font-sans mb-4">
          Incidencia de costos por cultivo
        </h3>
        <div className="p-4 text-sm text-gray-600 rounded-lg bg-gray-50">
          No crop data available
        </div>
      </div>
    );
  }

  const { crop_incidence } = dashboard;
  const crops = crop_incidence.items.sort((a, b) => a.crop_id - b.crop_id);

  const getCropBackgroundClass = (cropName: string) => {
    if (!cropName || !cropColors[cropName]) return "bg-gray-50";
    const match = cropColors[cropName].match(/bg-\[[^\]]+]/);
    return match ? match[0] : "bg-gray-50";
  };

  return (
    <div className="bg-white rounded-xl border p-4 w-full">
      <h3 className="font-medium text-[#020617] font-sans mb-4 text-xl">
        Incidencia de costos por cultivo
      </h3>

      <div className="grid grid-cols-4 text-sm font-semibold mb-2">
        <div></div>
        <div className="bg-[#E5E7EB] text-gray-900 mx-1 rounded h-[43px] text-center flex justify-center items-center">
          Superficie Has
        </div>
        <div className="bg-[#E5E7EB] text-gray-900 mx-1 rounded text-center flex justify-center items-center">
          % Rotación
        </div>
        <div className="bg-[#E5E7EB] text-gray-900 mx-1 rounded text-center flex justify-center items-center">
          Costo u$/Ha
        </div>
      </div>

      {crops.map((crop) => (
        <div
          key={crop.crop_id}
          className="grid grid-cols-4 text-sm my-2"
        >
          <div className={`${getCropBackgroundClass(crop.name)} h-[45px] font-semibold content-center pl-5 rounded-l-[5px]`}>{crop.name}</div>
          <div className={`${getCropBackgroundClass(crop.name)} h-[45px] text-center content-center`}>{crop.hectares} Has</div>
          <div className={`${getCropBackgroundClass(crop.name)} h-[45px] text-center content-center mr-1 rounded-r-[5px]`}>{crop.incidence_pct}%</div>
          <div className={`${getCropBackgroundClass(crop.name)} text-center content-center ml-1 rounded-[5px]`}>{crop.cost_per_ha_usd} u$/Ha</div>
        </div>
      ))}

      <div className="grid grid-cols-4 text-sm font-semibold text-slate-700 mt-1 rounded">
        <div className="h-[45px]"></div>
        <div className="bg-[#E5E7EB] h-[45px] text-center content-center rounded-l-[5px]">{crop_incidence.total.hectares} Has</div>
        <div className="bg-[#E5E7EB] h-[45px] text-center content-center">
          {crops.reduce((sum, crop) => sum + parseFloat(crop.incidence_pct), 0)}%
        </div>
        <div className="bg-[#E5E7EB] h-[45px] text-center content-center rounded-r-[5px]">{crop_incidence.total.avg_cost_per_ha_usd} u$/Ha</div>
      </div>
    </div>
  );
}