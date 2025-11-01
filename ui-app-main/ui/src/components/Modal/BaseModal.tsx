import React from "react";

interface BaseModalProps {
  isOpen: boolean;
  isSaving?: boolean;
  onClose: () => void;
  title?: string;
  message?: string;
  icon?: React.ReactNode;
  primaryButtonText?: string | null;
  secondaryButtonText?: string;
  onPrimaryAction?: () => void;
  onSecondaryAction?: () => void;
  children?: React.ReactNode;
  /**
   * Tailwind class string para el color del botón principal.
   * Ejemplo: "bg-blue-600 hover:bg-blue-800 focus:ring-blue-300 dark:focus:ring-blue-800"
   */
  primaryButtonColor?: string;
}

export function BaseModal({
  isOpen,
  isSaving = false,
  onClose,
  title = "Confirmación",
  message = "¿Está seguro de continuar?",
  icon,
  primaryButtonText,
  secondaryButtonText = "Cancelar",
  onPrimaryAction,
  onSecondaryAction,
  children,
  primaryButtonColor = "bg-red-600 hover:bg-red-800 focus:ring-red-300 dark:focus:ring-red-800",
}: BaseModalProps) {
  const defaultIcon = (
    <svg
      className="mx-auto mb-4 text-gray-800 w-12 h-12"
      aria-hidden="true"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 20 20"
    >
      <path
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
        d="M10 11V6m0 8h.01M19 10a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
      />
    </svg>
  );

  return (
    <div
      id="popup-modal"
      tabIndex={-1}
      className={`fixed top-0 right-0 left-0 z-50 flex justify-center items-center w-full h-screen bg-black bg-opacity-50 ${
        isOpen ? "flex" : "hidden"
      }`}
    >
      <div className="relative p-4 w-full max-w-md max-h-full">
        <div className="relative bg-white rounded-lg shadow-sm dark:bg-gray-700">
          <button
            type="button"
            onClick={onClose}
            disabled={isSaving}
            className="absolute top-3 end-2.5 text-gray-800 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm w-8 h-8 ms-auto inline-flex justify-center items-center"
          >
            <svg
              className="w-3 h-3"
              aria-hidden="true"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 14 14"
            >
              <path
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
              />
            </svg>
            <span className="sr-only">Cerrar modal</span>
          </button>

          {/* Contenido del modal */}
          <div className="p-4 md:p-5 text-center">
            {icon ? icon : title !== "" ? defaultIcon : null}
            <h3 className="mb-2 text-lg font-semibold text-gray-900">
              {title}
            </h3>

            {children ? (
              <div className="mb-5">{children}</div>
            ) : (
              <p className="mb-5 text-sm text-gray-700">{message}</p>
            )}

            {primaryButtonText && (
              <button
                disabled={isSaving}
                onClick={onPrimaryAction}
                type="button"
                className={`text-white font-medium rounded-lg text-sm inline-flex items-center px-5 py-2.5 text-center focus:ring-4 focus:outline-none ${primaryButtonColor}`}
              >
                {primaryButtonText}
              </button>
            )}
            <button
              disabled={isSaving}
              onClick={onSecondaryAction || onClose}
              type="button"
              className="py-2.5 px-5 ms-3 text-sm font-medium text-gray-900 focus:outline-none bg-white rounded-lg border border-gray-200 hover:bg-gray-100 hover:text-blue-700 focus:z-10 focus:ring-4 focus:ring-gray-100 dark:focus:ring-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-600 dark:hover:text-white dark:hover:bg-gray-700"
            >
              {secondaryButtonText}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
