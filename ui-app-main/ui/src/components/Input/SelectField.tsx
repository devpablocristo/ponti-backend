import { ChevronDown } from "lucide-react";

type SelectFieldProps = {
  label: string;
  name: string;
  placeholder?: string;
  options: { id: number; name: string }[];
  value: string;
  onChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  className?: string;
  size?: "sm" | "md";
  fullWidth?: boolean;
  disabled?: boolean;
};

const SelectField: React.FC<SelectFieldProps> = ({
  label,
  name,
  placeholder = "",
  options,
  value,
  onChange,
  className = "",
  size = "md",
  fullWidth = false,
  disabled = false,
}) => {
  const sizeClasses = size === "sm" ? "text-sm py-2" : "text-base py-3";

  return (
    <div className={`${fullWidth ? "w-full" : ""}`}>
      {label !== "" && (
        <label className="block mb-1 text-sm font-medium text-gray-900 leading-[1.5]">
          {label}
        </label>
      )}
      <div className="relative">
        <select
          name={name}
          value={value}
          disabled={disabled}
          onChange={onChange}
          className={`bg-gray-50 border border-gray-400 text-gray-900 rounded-lg focus:ring-custom-btn focus:border-custom-btn block px-4 w-full appearance-none ${sizeClasses} ${className}`}
        >
          <option value="" disabled>
            {placeholder ? placeholder : "Seleccionar..."}
          </option>
          {options.map((opt) => (
            <option key={opt.id} value={opt.id}>
              {opt.name}
            </option>
          ))}
        </select>
        <ChevronDown
          size={16}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 pointer-events-none"
        />
      </div>
    </div>
  );
};

export default SelectField;
