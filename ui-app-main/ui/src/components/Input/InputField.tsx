import React from "react";

type InputFieldProps = {
  label: string;
  type?: "text" | "number" | "password" | "date";
  name: string;
  value: string | number;
  onBlur?: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onFocus?: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  onPaste?: (e: React.ClipboardEvent<HTMLInputElement>) => void;
  disabled?: boolean;
  placeholder?: string;
  step?: string;
  required?: boolean;
  fullWidth?: boolean;
  inputClassName?: string;
  className?: string;
  size?: "sm" | "md" | "lg" | "xs";
};

const InputField = React.forwardRef<HTMLInputElement, InputFieldProps>(
  (
    {
      label = "",
      type = "text",
      name,
      value,
      onBlur,
      onChange,
      onKeyDown,
      onFocus,
      onPaste,
      disabled = false,
      placeholder,
      step,
      required = false,
      fullWidth = false,
      inputClassName = "",
      className = "",
      size = "md",
    },
    ref
  ) => {
    const sizeClasses =
      size === "sm"
        ? "text-sm py-2 px-4"
        : size === "lg"
        ? "text-lg py-4 px-6"
        : "text-base py-3 px-4";

    return (
      <div className={`${fullWidth ? "w-full" : ""} ${className}`}>
        {label !== "" && (
          <label
            className={`"block mb-2 text-sm font-medium ${
              disabled ? "text-gray-600" : "text-gray-900"
            } leading-[1.5]"`}
          >
            {label}
          </label>
        )}
        <input
          ref={ref}
          autoComplete="off"
          type={type}
          name={name}
          value={value}
          onBlur={onBlur}
          onChange={onChange}
          onKeyDown={onKeyDown}
          onFocus={onFocus}
          onPaste={onPaste}
          placeholder={placeholder}
          step={step}
          className={`border rounded-lg block w-full ${
            disabled
              ? "bg-gray-50 border-gray-300 text-gray-800 cursor-not-allowed"
              : "bg-white border-gray-500 text-gray-900"
          } ${sizeClasses} ${inputClassName}`}
          required={required}
          disabled={disabled}
        />
      </div>
    );
  }
);

export default InputField;
