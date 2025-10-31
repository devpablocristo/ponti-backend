import { ClipboardList, Newspaper, Calendar, ChartNoAxesCombined } from "lucide-react";
import { DashboardData } from "../../../hooks/useDashboard/types";

interface OperationalIndicatorsProps {
  dashboard: DashboardData | null;
}

export default function OperationalIndicators({dashboard}: OperationalIndicatorsProps) {
  if (!dashboard || !dashboard.operational_indicators || !dashboard.operational_indicators.items) {
    return (
      <div className="p-4 border rounded-xl bg-white">
        <h2 className="font-semibold text-lg mb-4">Indicadores operativos</h2>
        <div className="p-4 text-sm text-gray-600 rounded-lg bg-gray-50">
          No operational data available
        </div>
      </div>
    );
  }

  const { operational_indicators } = dashboard;

  const getIconForType = (type: string) => {
    switch (type) {
      case "first_workorder":
      case "last_workorder":
        return <Newspaper className="text-[#A4CAFE] w-10 h-10 -scale-x-100" />;
      case "last_stock_count":
        return <ChartNoAxesCombined className="text-[#84E1BC] w-10 h-10" />;
      case "campaign_closing":
        return <ClipboardList className="text-[#F8B4B4] w-10 h-10" />;
      default:
        return <Newspaper className="text-gray-500 w-10 h-10 -scale-x-100" />;
    }
  };

  const formatDate = (dateString: string | null) => {
    if (!dateString) return "N/A";
    try {
      return new Date(dateString).toLocaleDateString('es-ES');
    } catch {
      return dateString;
    }
  };

  const getCodeDisplay = (item: any) => {
    if (item.workorder_id) return `N°${item.workorder_id}`;
    return null;
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 items-center">
      <h2 className="font-medium text-[#020617] font-sans text-lg">Indicadores operativos</h2>
      {operational_indicators.items.map((item, index) => (
        <div
          key={index}
          className={`bg-white items-center gap-4 px-6 border rounded-lg min-h-28 ${item.type == "campaign_closing" ? "py-4" : "py-6"}`}
        >
          <p className="text-xs font-medium">{item.title}</p>
          <div className="my-4">{getIconForType(item.type)}</div>
          <div className={`flex items-center text-xs text-[#6B7280] gap-2 mt-1 ${item.type == "campaign_closing" && "border rounded-lg py-2 px-3 mr-10"}`}>
            <Calendar className="w-4 h-4" />
            <span>{formatDate(item.date)}</span>
            {getCodeDisplay(item) && <span>- {getCodeDisplay(item)}</span>}
          </div>
        </div>
      ))}
    </div>
  );
}