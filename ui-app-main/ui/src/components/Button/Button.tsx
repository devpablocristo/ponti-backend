import { ReactNode } from "react";
import { Link } from "react-router-dom";

interface ButtonProps {
  children: ReactNode;
  variant?:
    | "primary"
    | "secondary"
    | "success"
    | "danger"
    | "warning"
    | "light"
    | "dark"
    | "outlineGreen"
    | "outlineGray"
    | "outlinePonti";
  size?: "sm" | "md" | "lg" | "xs";
  iconLeft?: ReactNode;
  iconRight?: ReactNode;
  onClick?: () => void;
  href?: string;
  className?: string;
  type?: "button" | "submit" | "reset";
  disabled?: boolean;
}

const variantClasses = {
  primary: "bg-blue-700 hover:bg-blue-800 focus:ring-blue-300 text-white",
  secondary: "bg-gray-200 hover:bg-gray-300 focus:ring-gray-400 text-gray-900",
  success:
    "bg-custom-btn hover:bg-custom-btn/80 focus:ring-green-300 text-white",
  danger: "bg-red-700 hover:bg-red-800 focus:ring-red-300 text-white",
  warning: "bg-yellow-400 hover:bg-yellow-500 focus:ring-yellow-300 text-white",
  light:
    "bg-white border border-gray-300 hover:bg-gray-100 focus:ring-gray-100 text-gray-900",
  dark: "bg-gray-800 hover:bg-gray-900 focus:ring-gray-700 text-white",
  outlineGreen: "border border-custom-btn text-custom-btn hover:bg-green-50",
  outlineGray:
    "bg-white border border-gray-300 text-gray-900 hover:bg-gray-100",
  outlinePonti:
    "bg-transparent border border-[#547792] text-[#547792] hover:bg-gray-100",
};

const sizeClasses = {
  xs: "px-2 py-1.5 text-xs",
  sm: "px-3 py-2 text-sm",
  md: "px-5 py-2.5 text-md",
  lg: "px-6 py-3 text-lg",
};

export default function Button({
  children,
  variant = "primary",
  size = "md",
  iconLeft,
  iconRight,
  onClick,
  href,
  className = "",
  disabled = false,
  ...props
}: ButtonProps) {
  const classes = `inline-flex items-center font-medium rounded-lg focus:ring-4 focus:outline-none ${
    disabled ? "opacity-60 cursor-not-allowed" : ""
  } ${variantClasses[variant]} ${sizeClasses[size]} ${className}`;

  if (href) {
    return (
      <Link to={href} className={classes}>
        {iconLeft && <span className="me-2">{iconLeft}</span>}
        {children}
        {iconRight && <span className="ms-2">{iconRight}</span>}
      </Link>
    );
  }

  return (
    <button
      type={props.type || "button"}
      className={classes}
      disabled={disabled}
      onClick={disabled ? undefined : onClick}
      {...props}
    >
      {iconLeft && <span className="me-2">{iconLeft}</span>}
      {children}
      {iconRight && <span className="ms-2">{iconRight}</span>}
    </button>
  );
}
