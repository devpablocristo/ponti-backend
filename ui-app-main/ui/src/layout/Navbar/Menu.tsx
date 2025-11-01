// import { Token, User } from "@/pages/auth/types";
import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import Button from "../../components/Button/Button";

interface NavbarProps {
  username: string;
  setIsLogoutModalOpen: () => void;
}

const Menu: React.FC<NavbarProps> = ({ setIsLogoutModalOpen, username }) => {
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);

  const toggleDropdown = () => {
    setIsDropdownOpen(!isDropdownOpen);
  };

  const handleClickOutside = (event: MouseEvent) => {
    if (
      dropdownRef.current &&
      !dropdownRef.current.contains(event.target as Node) &&
      buttonRef.current &&
      !buttonRef.current.contains(event.target as Node)
    ) {
      setIsDropdownOpen(false);
    }
  };

  useEffect(() => {
    if (isDropdownOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    } else {
      document.removeEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isDropdownOpen]);

  return (
    <div className="relative flex items-center ms-3">
      <div>
        <button
          ref={buttonRef}
          type="button"
          className="flex text-sm bg-gray-800 rounded-full focus:ring-4 focus:ring-gray-600"
          aria-expanded={isDropdownOpen ? "true" : "false"}
          onClick={toggleDropdown}
        >
          <div className="w-10 h-10 rounded-full bg-white flex items-center justify-center">
            <svg
              className="w-8 h-8 text-blue-900"
              fill="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                fillRule="evenodd"
                d="M12 12c2.21 0 4-1.79 4-4S14.21 4 12 4 8 5.79 8 8s1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"
                clipRule="evenodd"
              />
            </svg>
          </div>
        </button>
      </div>
      <div
        ref={dropdownRef}
        className={`absolute right-0.5 top-full mt-2 z-[9999] ${
          isDropdownOpen ? "" : "hidden"
        } text-base list-none bg-white divide-y divide-gray-100 rounded shadow`}
      >
        <div className="px-4 py-3" role="none">
          <p className="text-sm text-gray-900 dark:text-white" role="none">
            <b>{username}</b>
          </p>
        </div>
        <ul className="py-1" role="none">
          <li>
            <Link
              to="/admin/users"
              onClick={toggleDropdown}
              className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
            >
              Usuarios
            </Link>
          </li>
          <li>
            <Link
              to="/admin/profile"
              onClick={toggleDropdown}
              className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
            >
              Mi Perfil
            </Link>
          </li>
          <li>
            <Button
              className="mx-2 mt-2 py-2 text-sm w-44"
              variant="danger"
              size="sm"
              onClick={setIsLogoutModalOpen}
            >
              Salir
            </Button>
          </li>
        </ul>
      </div>
    </div>
  );
};

export default Menu;
