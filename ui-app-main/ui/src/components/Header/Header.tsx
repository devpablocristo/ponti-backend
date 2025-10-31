import React from "react";
import Button from "../Button/Button";

interface ActionButton {
  label: string;
  path: string;
  variant?: "primary" | "secondary" | "danger";
}

interface HeaderProps {
  title: string;
  actionButtons?: ActionButton[];
  topRightContent?: React.ReactNode;
}

const Header: React.FC<HeaderProps> = ({
  title,
  actionButtons = [],
  topRightContent,
}) => {
  return (
    <header className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      {title !== "" && (
        <h2 className="text-3xl font-extrabold leading-none tracking-tight text-gray-900 md:text-4xl dark:text-white">
          {title}
        </h2>
      )}
      {topRightContent && <div className="flex gap-2">{topRightContent}</div>}
      {actionButtons.length > 0 && (
        <div className="hidden ml-auto md:flex gap-2">
          {actionButtons.map((button, index) => (
            <Button
              key={index}
              variant={button.variant || "primary"}
              href={button.path}
            >
              {button.label}
            </Button>
          ))}
        </div>
      )}
      {actionButtons.length > 0 && (
        <div className="md:hidden fixed bottom-4 right-4 z-50 flex gap-2">
          {actionButtons.map((button, index) => (
            <Button
              key={index}
              variant={button.variant || "primary"}
              href={button.path}
              className="shadow-lg rounded-full p-4 flex items-center justify-center"
            >
              <b>{button.label}</b>
            </Button>
          ))}
        </div>
      )}
    </header>
  );
};

export default Header;
