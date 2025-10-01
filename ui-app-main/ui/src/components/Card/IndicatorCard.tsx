interface IndicatorCardProps {
  title: string;
  value: string;
  subtext?: string;
  icon?: React.ReactNode;
  color?: "green" | "red" | "gray" | "default";
  height?: string;
  width?: string;
  className?: string;
}

const colorMap = {
  green: "text-custom-green",
  red: "text-red-600",
  gray: "text-gray-500",
  default: "text-gray-700",
};

export function IndicatorCard({
  title,
  value,
  subtext,
  icon,
  color = "default",
  height = "130px",
  width = "170px",
  className = "",
}: IndicatorCardProps) {
  return (
    <div
      className={`bg-white rounded-xl border border-gray-300 px-4 py-2 min-h-[${height}] min-w-[${width}] flex flex-col justify-between shadow-sm ${className}`}
    >
      <div className="text-[14px] font-medium text-gray-950">{title}</div>
      <div className={`text-[24px] font-light ${colorMap[color]}`}>{value}</div>
      {subtext && (
        <div
          className={`text-sm ${colorMap[color]} mt-2 flex items-center gap-1}`}
        >
          {icon}
          {subtext}
        </div>
      )}
    </div>
  );
}
