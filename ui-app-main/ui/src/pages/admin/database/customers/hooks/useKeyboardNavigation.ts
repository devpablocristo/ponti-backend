import { useState } from "react";

type UseKeyboardNavigationProps<T> = {
  suggestions: T[];
  showSuggestions: boolean;
  onSelect: (item: T) => void;
  onEscape?: () => void;
};

export const useKeyboardNavigation = <T>({
  suggestions,
  showSuggestions,
  onSelect,
  onEscape,
}: UseKeyboardNavigationProps<T>) => {
  const [highlightedIndex, setHighlightedIndex] = useState(0);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (!showSuggestions || suggestions.length === 0) return;

    if (e.key === "ArrowDown") {
      e.preventDefault();
      setHighlightedIndex((prev) =>
        prev < suggestions.length - 1 ? prev + 1 : 0
      );
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setHighlightedIndex((prev) =>
        prev > 0 ? prev - 1 : suggestions.length - 1
      );
    } else if (e.key === "Enter") {
      e.preventDefault();
      const selected = suggestions[highlightedIndex];
      onSelect(selected);
    } else if (e.key === "Escape") {
      e.preventDefault();
      onEscape?.();
    }
  };

  return {
    highlightedIndex,
    setHighlightedIndex,
    handleKeyDown,
  };
};
