import React from "react";

import Menu from "./Menu";

interface NavbarProps {
  username: string;
  title: string;
  isSidebarOpen: boolean;
  toggleSidebar: (e?: React.MouseEvent) => void;
  setIsLogoutModalOpen: () => void;
}

const Navbar: React.FC<NavbarProps> = ({
  username,
  title,
  isSidebarOpen,
  toggleSidebar,
  setIsLogoutModalOpen,
}) => {
  return (
    <nav className="w-full bg-custom-bg shadow-sm">
      <div className="mx-auto px-4 py-2 flex flex-col">
        <div className="flex items-center justify-between">
          <div className="flex items-center justify-start rtl:justify-end mt-2">
            {!isSidebarOpen && (
              <button
                onClick={(e) => toggleSidebar(e)}
                type="button"
                className="inline-flex items-center p-2 text-sm text-gray-500 rounded-lg hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-gray-200"
              >
                <span className="sr-only">Open sidebar</span>
                <svg
                  className="w-6 h-6"
                  aria-hidden="true"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    clipRule="evenodd"
                    fillRule="evenodd"
                    d="M2 4.75A.75.75 0 012.75 4h14.5a.75.75 0 010 1.5H2.75A.75.75 0 012 4.75zm0 10.5a.75.75 0 01.75-.75h7.5a.75.75 0 010 1.5h-7.5a.75.75 0 01-.75-.75zM2 10a.75.75 0 01.75-.75h14.5a.75.75 0 010 1.5H2.75A.75.75 0 012 10z"
                  ></path>
                </svg>
              </button>
            )}
            <h1 className="text-[30px] font-semibold leading-[45px] text-black">
              {title}
            </h1>
          </div>
          <div className="flex items-center">
            <Menu
              setIsLogoutModalOpen={setIsLogoutModalOpen}
              username={username}
            />
          </div>
        </div>
        <div className="border-b mt-2" style={{ borderColor: "#D1D5DB" }} />
      </div>
    </nav>
  );
};

export default Navbar;
