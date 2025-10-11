import React, { useEffect, useRef, useState } from "react";

import Search from "../../../../components/Input/Search";

type BaseItem = {
  id: number;
  name: string;
};

type Props<T extends BaseItem> = {
  name: string;
  label: string;
  placeholder: string;
  selectedItems: T[];
  query: string;
  setQuery: (value: string) => void;
  options: T[] | undefined;
  setItems: (value: React.SetStateAction<T[]>) => void;
  handleSuggestionClick: (item: T) => void;
  customAddLabel: string;
  customAddItem: T;
  renderTag?: (item: T) => React.ReactNode;
};

function AutocompleteSelect<T extends BaseItem>({
  name,
  label,
  placeholder,
  options,
  selectedItems,
  query,
  setQuery,
  handleSuggestionClick,
  setItems,
  customAddLabel,
  customAddItem,
  renderTag,
}: Props<T>) {
  const [highlightedIndex, setHighlightedIndex] = useState(0);
  const [suggestions, setSuggestions] = useState<T[]>([]);

  const [showSuggestions, setShowSuggestions] = useState<boolean>(false);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (!showSuggestions) return;

    if (e.key === "ArrowDown" && suggestions.length > 0) {
      e.preventDefault();
      setHighlightedIndex((prev) =>
        prev < suggestions.length - 1 ? prev + 1 : 0
      );
    } else if (e.key === "ArrowUp" && suggestions.length > 0) {
      e.preventDefault();
      setHighlightedIndex((prev) =>
        prev > 0 ? prev - 1 : suggestions.length - 1
      );
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (suggestions.length === 0 && query.trim() !== "") {
        suggestionClick(customAddItem);
      } else if (suggestions.length > 0) {
        const selected = suggestions[highlightedIndex];
        if (
          selected &&
          !selectedItems.some((item) => item.name === selected.name)
        ) {
          suggestionClick(selected);
        }
      }
    }
  };

  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (options) {
      setSuggestions(options);
    }
  }, [options]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        wrapperRef.current &&
        !wrapperRef.current.contains(event.target as Node)
      ) {
        setShowSuggestions(false);
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setShowSuggestions(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const onFocus = () => {
    setShowSuggestions(true);
    if (query === "" && options) {
      setSuggestions(options);
    }
  };

  const handleChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQuery(value);

    if (options) {
      const filtered =
        value.trim() === ""
          ? options
          : options.filter((option) =>
              option.name.toLowerCase().includes(value.toLowerCase())
            );

      setSuggestions(filtered);
      setShowSuggestions(true);
    }
  };

  const suggestionClick = (item: T) => {
    handleSuggestionClick(item);
    setQuery("");
    setShowSuggestions(false);
  };

  const removeItem = (name: string) => {
    setItems(selectedItems.filter((m) => m.name !== name));
  };

  return (
    <div ref={wrapperRef} className="relative">
      <div className="w-full">
        <Search
          label={label}
          placeholder={placeholder}
          name={name}
          value={query}
          onClick={() => {
            if (!showSuggestions) {
              setShowSuggestions(true);
            }
          }}
          onFocus={onFocus}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          className={"w-full"}
          fullWidth
        />
        <div className="flex flex-wrap gap-2 mt-2">
          {selectedItems.map((item, index) => (
            <div
              key={index}
              className="w-fit flex mt-1 items-center bg-custom-btn text-white text-sm font-sm rounded-full px-2 py-1"
            >
              <span>{renderTag ? renderTag(item) : item.name}</span>
              <button
                onClick={() => removeItem(item.name)}
                className="ml-1 hover:text-red-300"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
          ))}
        </div>
      </div>
      {showSuggestions && (
        <ul className="absolute top-16 mt-3 mb-1 w-full bg-white border rounded-lg shadow-md z-50 max-h-[200px] overflow-y-auto">
          {suggestions.map((item, index) => {
            const alreadySelected = selectedItems.some(
              (i) => i.name === item.name
            );
            return (
              <li
                key={index}
                onClick={
                  !alreadySelected ? () => suggestionClick(item) : undefined
                }
                className={`px-4 py-2 cursor-pointer ${
                  alreadySelected
                    ? "text-gray-400 cursor-not-allowed bg-gray-100"
                    : index === highlightedIndex
                    ? "bg-gray-300 font-medium"
                    : "hover:bg-gray-300 hover:font-medium"
                }`}
              >
                {item.name}
              </li>
            );
          })}
          {query !== "" && (
            <li className="px-4 py-2 text-gray-500">
              <button
                onClick={() => suggestionClick(customAddItem)}
                className="text-custom-btn hover:underline"
              >
                {customAddLabel}
              </button>
            </li>
          )}
        </ul>
      )}
    </div>
  );
}

export default AutocompleteSelect;
