import React, { useCallback, useRef, useState } from "react";
import Button from "../../components/Button/Button";
import Search from "../../components/Input/Search";
import SelectField from "../../components/Input/SelectField";
import { useClickOutside } from "../../pages/login/useClickOutside";
import { Entity } from "../../hooks/useDatabase/options/types";

interface FilterOption {
  id: number;
  name: string;
}

interface FilterItem {
  type: "select" | "search";
  ref?: string;
  name: string;
  label: string;
  total?: number;
  placeholder?: string;
  options?: FilterOption[];
  value?: string | null;
  disabled?: boolean;
  onChange: (value: string) => void;
  setData: (c: Entity | undefined) => void;
}

export interface ActionButton {
  label: string;
  variant?:
    | "primary"
    | "success"
    | "secondary"
    | "outlineGreen"
    | "danger"
    | "warning"
    | "light"
    | "dark"
    | "outlineGray"
    | "outlinePonti"
    | undefined;
  icon?: React.ReactNode;
  disabled?: boolean;
  onClick?: () => void;
  href?: string;
  isPrimary?: boolean;
}

interface ResponsiveButtonProps {
  actions: ActionButton[];
}

const ResponsiveButtonContainer: React.FC<ResponsiveButtonProps> = ({
  actions,
}) => {
  const [open, setOpen] = useState(false);

  return (
    <div className="fixed bottom-6 right-6 z-50">
      {open && (
        <div className="mb-2 flex flex-col items-end gap-2">
          {actions.map((action, index) => (
            <Button
              key={`floating-action-${index}`}
              variant={
                action.variant ||
                (action.isPrimary ? "success" : "outlineGreen")
              }
              size="sm"
              className="whitespace-nowrap disabled:opacity-50 disabled:cursor-not-allowed"
              iconLeft={action.icon}
              href={action.href}
              onClick={action.onClick}
              disabled={action.disabled}
            >
              {action.label}
            </Button>
          ))}
        </div>
      )}
      <button
        onClick={() => setOpen(!open)}
        className="bg-custom-btn hover:bg-custom-btn/80 text-white w-14 h-14 rounded-full shadow-lg flex items-center justify-center"
        aria-label="Acciones"
      >
        <span className="text-2xl">+</span>
      </button>
    </div>
  );
};

interface FilterBarProps {
  filters: FilterItem[];
  actions?: ActionButton[];
  children?: React.ReactNode;
  className?: string;
  inputSize?: "sm" | "md";
}

const FilterBar: React.FC<FilterBarProps> = ({
  filters,
  actions = [],
  children,
  className = "",
  inputSize = "sm",
}) => {
  const [suggestionsVisible, setSuggestionsVisible] = useState<{
    [key: string]: boolean;
  }>({});
  const [highlightedIndex, setHighlightedIndex] = useState<{
    [key: string]: number;
  }>({});

  const hideClientSuggestions = useCallback(() => {
    hideSuggestions("cliente");
  }, []);

  const hideProjectSuggestions = useCallback(() => {
    hideSuggestions("proyecto");
  }, []);

  const showSuggestions = useCallback((name: string) => {
    setSuggestionsVisible((prev) => ({ ...prev, [name]: true }));
  }, []);

  const hideSuggestions = useCallback((name: string) => {
    setSuggestionsVisible((prev) => ({ ...prev, [name]: false }));
  }, []);

  const handleSuggestionClick = useCallback(
    (
      name: string,
      option: any,
      onChange: (val: any) => void,
      setData: (c: Entity | undefined) => void
    ) => {
      setData(option);
      onChange(option.name);
      hideSuggestions(name);
    },
    [hideSuggestions]
  );

  const customerRef = useRef<HTMLDivElement>(null);
  const projectRef = useRef<HTMLDivElement>(null);

  type FilterRefKey = "client" | "project";

  const refs: Record<FilterRefKey, React.RefObject<HTMLDivElement | null>> = {
    client: customerRef,
    project: projectRef,
  };

  useClickOutside(customerRef, hideClientSuggestions);
  useClickOutside(projectRef, hideProjectSuggestions);

  const createHandleKeyDown = useCallback(
    (
        name: string,
        options: any[],
        onChange: (val: any) => void,
        setData: (c: any | undefined) => void
      ) =>
      (e: React.KeyboardEvent<HTMLInputElement>) => {
        const currentIndex = highlightedIndex[name] || 0;
        let newIndex = currentIndex;

        if (e.key === "ArrowDown") {
          e.preventDefault();
          newIndex = Math.min(options.length - 1, currentIndex + 1);
        } else if (e.key === "ArrowUp") {
          e.preventDefault();
          newIndex = Math.max(0, currentIndex - 1);
        } else if (e.key === "Enter") {
          e.preventDefault();
          if (options.length > 0) {
            handleSuggestionClick(
              name,
              options[currentIndex],
              onChange,
              setData
            );
          }
        } else if (e.key === "Escape") {
          e.preventDefault();
          hideSuggestions(name);
        }

        setHighlightedIndex((prev) => ({ ...prev, [name]: newIndex }));
      },
    [highlightedIndex, handleSuggestionClick, hideSuggestions]
  );

  return (
    <div className={`w-full ${className}`}>
      <div className="flex flex-col sm:flex-row items-start sm:items-end justify-between gap-3 sm:gap-4 px-1 py-2">
        <div className="flex flex-col w-full sm:flex-row sm:flex-1 gap-4">
          {filters.map((filter, index) => {
            const ref = filter.ref
              ? refs[filter.ref as FilterRefKey]
              : undefined;

            return (
              <div
                key={`filter-${index}`}
                className="w-full sm:w-auto flex-1 sm:max-w-60"
              >
                {filter.type === "select" ? (
                  <SelectField
                    className={
                      filter.disabled ? "opacity-50 cursor-not-allowed" : ""
                    }
                    label={filter.label}
                    name={filter.name}
                    placeholder={
                      filter.placeholder ||
                      `Seleccione ${filter.label.toLowerCase()}`
                    }
                    value={
                      filter.value !== undefined && filter.value !== null
                        ? String(filter.value)
                        : ""
                    }
                    onChange={(e) => {
                      const selectedId = e.target.value;
                      const selectedItem = filter.options?.find(
                        (opt) => String(opt.id) === selectedId
                      );
                      filter.setData(selectedItem);
                    }}
                    options={filter.options || []}
                    disabled={filter.disabled}
                    fullWidth
                    size={inputSize}
                  />
                ) : (
                  <div key={filter.name} ref={ref} className="relative">
                    <Search
                      label={filter.label}
                      name={filter.name}
                      placeholder={
                        filter.placeholder ||
                        `Buscar ${filter.label.toLowerCase()}`
                      }
                      value={filter.value || ""}
                      size={inputSize}
                      onClick={() => showSuggestions(filter.name)}
                      onChange={(e) => filter.onChange(e.target.value)}
                      onFocus={() => showSuggestions(filter.name)}
                      onKeyDown={createHandleKeyDown(
                        filter.name,
                        filter.options || [],
                        filter.onChange,
                        filter.setData
                      )}
                      className={"w-full"}
                      disabled={filter.disabled}
                      fullWidth
                    />
                    {suggestionsVisible[filter.name] && filter.options && (
                      <div className="flex justify-between items-center">
                        <ul className="absolute top-full mb-1 w-full bg-white border rounded-lg shadow-md z-10 max-h-[200px] overflow-y-auto">
                          <li
                            className="p-2 cursor-pointer hover:bg-gray-100"
                            onClick={() => {
                              filter.setData({
                                id: 0,
                                name: "Todos los registros",
                              });
                              filter.onChange("Todos los registros");
                              hideSuggestions(filter.name);
                            }}
                          >
                            Todos los registros
                          </li>
                          {filter.options.length > 0 &&
                            filter.options.map((option, index) => (
                              <li
                                key={index}
                                onClick={() =>
                                  handleSuggestionClick(
                                    filter.name,
                                    option,
                                    filter.onChange,
                                    filter.setData
                                  )
                                }
                                className={`px-4 py-2 cursor-pointer ${
                                  highlightedIndex[filter.name] === index
                                    ? "bg-gray-300 font-medium"
                                    : "hover:bg-gray-300 hover:font-medium"
                                }`}
                              >
                                {option.name}
                              </li>
                            ))}
                        </ul>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
          {children && <div className="mx-2">{children}</div>}
        </div>
        {actions && actions.length > 0 && (
          <>
            <div className={`hidden sm:flex gap-2 items-center justify-end}`}>
              {actions.map((action, index) => (
                <Button
                  key={`action-${index}`}
                  variant={
                    action.variant ||
                    (action.isPrimary ? "success" : "outlineGreen")
                  }
                  disabled={action.disabled}
                  href={action.href}
                  onClick={action.onClick}
                  size={inputSize}
                  className="whitespace-nowrap"
                  iconLeft={action.icon}
                >
                  {action.label}
                </Button>
              ))}
            </div>
            <div className="sm:hidden">
              <ResponsiveButtonContainer actions={actions} />
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default FilterBar;
