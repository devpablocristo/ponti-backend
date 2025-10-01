import { useEffect, useState } from "react";
import { useDebounce } from "use-debounce";
import InputField from "../../components/Input/InputField";

import useFields from "../../hooks/useFields";
import { Entity } from "../../hooks/useDatabase/options/types";
import { Check, ChevronDown } from "lucide-react";

interface Props {
  onSelect: (field: Entity | undefined) => void;
  field: Entity | undefined;
  placeholder?: string;
  label?: string;
}

const FieldSearch: React.FC<Props> = ({
  onSelect,
  placeholder = "Buscar por nombre de campo",
  label = "",
  field,
}) => {
  const [inputValue, setInputValue] = useState("");
  const [debouncedValue] = useDebounce(inputValue, 400);
  const [filteredFields, setFilteredFields] = useState<Entity[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [skipSearch, setSkipSearch] = useState(false);

  const { getFields, fields, total } = useFields();

  useEffect(() => {
    setFilteredFields(fields);
  }, [fields]);

  useEffect(() => {
    if (skipSearch) {
      setSkipSearch(false);
      return;
    }

    const search = async () => {
      if (debouncedValue.length >= 3) {
        if (total <= 100 && fields.length > 0) {
          const filtered = fields.filter((field) =>
            field.name.toLowerCase().includes(debouncedValue.toLowerCase())
          );
          setFilteredFields(filtered);
        } else {
          await getFields(`name=${debouncedValue}`);
        }
        setShowSuggestions(true);
      } else {
        setFilteredFields([]);
        setShowSuggestions(false);
      }
    };

    search();
  }, [debouncedValue]);

  const handleSelect = (field: Entity) => {
    setSkipSearch(true);
    setInputValue(field.name);
    setShowSuggestions(false);
    onSelect(field);
  };

  return (
    <div className="relative">
      <div className="relative w-full">
        <svg
          width="18"
          height="18"
          viewBox="0 0 18 18"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="absolute left-3 top-2 transform -translate-y-1/2 mt-4 text-gray-500"
        >
          <path
            d="M7.65007 13.05C6.58203 13.05 5.53799 12.7333 4.64995 12.1399C3.76192 11.5466 3.06978 10.7032 2.66106 9.71649C2.25234 8.72977 2.1454 7.64401 2.35376 6.59651C2.56213 5.54902 3.07643 4.58683 3.83164 3.83162C4.58686 3.07642 5.54906 2.56212 6.59657 2.35376C7.64408 2.1454 8.72985 2.25234 9.71658 2.66105C10.7033 3.06977 11.5467 3.7619 12.1401 4.64992C12.7334 5.53795 13.0501 6.58198 13.0501 7.65C13.0485 9.08167 12.4791 10.4543 11.4667 11.4666C10.4544 12.4789 9.08176 13.0484 7.65007 13.05ZM7.65007 3.6C6.84904 3.6 6.06601 3.83753 5.39998 4.28255C4.73395 4.72757 4.21485 5.36009 3.90831 6.10013C3.60177 6.84017 3.52157 7.65449 3.67784 8.44011C3.83411 9.22574 4.21984 9.94738 4.78625 10.5138C5.35266 11.0802 6.07431 11.4659 6.85994 11.6222C7.64557 11.7784 8.4599 11.6982 9.19995 11.3917C9.94 11.0852 10.5725 10.5661 11.0176 9.90006C11.4626 9.23404 11.7001 8.45101 11.7001 7.65C11.699 6.5762 11.272 5.54669 10.5127 4.7874C9.7534 4.02811 8.72388 3.60107 7.65007 3.6Z"
            fill="#6B7280"
          />
          <path
            d="M15.0751 15.75C14.8961 15.75 14.7245 15.6788 14.5979 15.5522L11.8979 12.8522C11.7749 12.7249 11.7069 12.5544 11.7084 12.3774C11.71 12.2004 11.7809 12.0311 11.9061 11.906C12.0312 11.7808 12.2005 11.7099 12.3775 11.7083C12.5545 11.7068 12.725 11.7748 12.8523 11.8978L15.5524 14.5978C15.6467 14.6922 15.711 14.8124 15.737 14.9434C15.7631 15.0743 15.7497 15.21 15.6986 15.3333C15.6475 15.4566 15.5611 15.562 15.4501 15.6362C15.3391 15.7104 15.2086 15.75 15.0751 15.75Z"
            fill="#6B7280"
          />
        </svg>
        <InputField
          label={label}
          placeholder={placeholder}
          type="text"
          name="client"
          value={inputValue}
          onBlur={() => {
            setTimeout(() => {
              setShowSuggestions(false);
            }, 100);
          }}
          onChange={(e) => {
            onSelect(undefined);
            setInputValue(e.target.value);
          }}
          fullWidth
          inputClassName="pl-9 pr-3"
        />
        {field ? (
          <Check size={20} className="absolute right-3 top-4 text-custom-btn" />
        ) : (
          <ChevronDown
            size={18}
            className="absolute right-3 top-4 text-gray-500"
          />
        )}
      </div>
      {showSuggestions && (
        <div className="flex justify-between items-center">
          <ul className="absolute top-full left-0 w-full bg-white border rounded-lg shadow-md z-10">
            {filteredFields.length > 0 &&
              filteredFields.map((field, index) => (
                <li
                  key={index}
                  onClick={() => handleSelect(field)}
                  className="px-4 py-2 cursor-pointer hover:bg-gray-300"
                >
                  {field.name}
                </li>
              ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default FieldSearch;
