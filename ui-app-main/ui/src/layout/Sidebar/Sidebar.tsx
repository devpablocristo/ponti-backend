import { ChevronDown, ChevronUp } from "lucide-react";
import React, { ReactNode, useEffect, useState } from "react";

import { Link, useLocation } from "react-router-dom";

type MenuItem = {
  name: string;
  icon: (color: string) => ReactNode;
  route: string;
};

type SubItem = {
  name: string;
  route: string;
};

const menuReports = {
  name: "Informes",
  icon: (
    <svg
      width="20"
      height="20"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M10 2.66667H12C12.1769 2.66667 12.3464 2.7369 12.4714 2.86193C12.5965 2.98695 12.6667 3.15652 12.6667 3.33333V13.3333C12.6667 13.5101 12.5965 13.6797 12.4714 13.8047C12.3464 13.9298 12.1769 14 12 14H4.00004C3.82323 14 3.65366 13.9298 3.52864 13.8047C3.40361 13.6797 3.33337 13.5101 3.33337 13.3333V3.33333C3.33337 3.15652 3.40361 2.98695 3.52864 2.86193C3.65366 2.7369 3.82323 2.66667 4.00004 2.66667H6.00004M6.00004 4.66667H10M8.00004 8H10M6.00004 8H6.00671M8.00004 10.6667H10M6.00004 10.6667H6.00671M6.66671 2V4.66667H9.33337V2H6.66671Z"
        stroke="#64748B"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  ),
  children: [
    { name: "Aportes por inversor", route: "/admin/informes/aportes" },
    { name: "Por campo o cultivo", route: "/admin/informes/campo" },
    { name: "Resumen de resultados", route: "/admin/informes/resumen" },
  ],
};

const menuDatabase = {
  name: "Base de Datos",
  icon: (
    <svg
      width="20"
      height="20"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M14 8.66667V7.33333C14 7.15652 13.9298 6.98695 13.8047 6.86193C13.6797 6.7369 13.5101 6.66667 13.3333 6.66667H12.8287L12.3573 5.52867L12.714 5.17133C12.839 5.04631 12.9092 4.87678 12.9092 4.7C12.9092 4.52322 12.839 4.35369 12.714 4.22867L11.7713 3.286C11.6463 3.16102 11.4768 3.09081 11.3 3.09081C11.1232 3.09081 10.9537 3.16102 10.8287 3.286L10.4713 3.64267L9.33333 3.17133V2.66667C9.33333 2.48986 9.2631 2.32029 9.13807 2.19526C9.01305 2.07024 8.84348 2 8.66667 2H7.33333C7.15652 2 6.98695 2.07024 6.86193 2.19526C6.7369 2.32029 6.66667 2.48986 6.66667 2.66667V3.17133L5.52867 3.64267L5.17133 3.286C5.04631 3.16102 4.87678 3.09081 4.7 3.09081C4.52322 3.09081 4.35369 3.16102 4.22867 3.286L3.286 4.22867C3.16102 4.35369 3.09081 4.52322 3.09081 4.7C3.09081 4.87678 3.16102 5.04631 3.286 5.17133L3.64333 5.52867L3.17133 6.66667H2.66667C2.48986 6.66667 2.32029 6.7369 2.19526 6.86193C2.07024 6.98695 2 7.15652 2 7.33333V8.66667C2 8.84348 2.07024 9.01305 2.19526 9.13807C2.32029 9.2631 2.48986 9.33333 2.66667 9.33333H3.17133L3.64267 10.4713L3.286 10.8287C3.16102 10.9537 3.09081 11.1232 3.09081 11.3C3.09081 11.4768 3.16102 11.6463 3.286 11.7713L4.22867 12.714C4.35369 12.839 4.52322 12.9092 4.7 12.9092C4.87678 12.9092 5.04631 12.839 5.17133 12.714L5.52867 12.3573L6.66667 12.8287V13.3333C6.66667 13.5101 6.7369 13.6797 6.86193 13.8047C6.98695 13.9298 7.15652 14 7.33333 14H8.66667C8.84348 14 9.01305 13.9298 9.13807 13.8047C9.2631 13.6797 9.33333 13.5101 9.33333 13.3333V12.8287L10.4713 12.3567L10.8287 12.714C10.9537 12.839 11.1232 12.9092 11.3 12.9092C11.4768 12.9092 11.6463 12.839 11.7713 12.714L12.714 11.7713C12.839 11.6463 12.9092 11.4768 12.9092 11.3C12.9092 11.1232 12.839 10.9537 12.714 10.8287L12.3573 10.4713L12.8287 9.33333H13.3333C13.5101 9.33333 13.6797 9.2631 13.8047 9.13807C13.9298 9.01305 14 8.84348 14 8.66667Z"
        stroke="#64748B"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M8 10C8.53043 10 9.03914 9.78929 9.41421 9.41421C9.78929 9.03914 10 8.53043 10 8C10 7.46957 9.78929 6.96086 9.41421 6.58579C9.03914 6.21071 8.53043 6 8 6C7.46957 6 6.96086 6.21071 6.58579 6.58579C6.21071 6.96086 6 7.46957 6 8C6 8.53043 6.21071 9.03914 6.58579 9.41421C6.96086 9.78929 7.46957 10 8 10Z"
        stroke="#64748B"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  ),
  children: [
    {
      name: "Clientes y sociedades",
      route: "/admin/database/customers",
    },
    {
      name: "Labores",
      route: "/admin/database/tasks",
    },
    {
      name: "Insumos",
      route: "/admin/database/items",
    },
    {
      name: "Dólar promedio",
      route: "/admin/database/dollar",
    },
    {
      name: "Comercialización",
      route: "/admin/database/commerce",
    },
  ],
};

const menuItems: MenuItem[] = [
  {
    name: "Dashboard",
    icon: (color: string) => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M4.08863 13.9999C3.30329 13.9999 2.66663 13.3466 2.66663 12.5399V6.67194C2.66663 6.22861 2.86329 5.80861 3.19996 5.53194L7.11129 2.31994C7.3612 2.113 7.67549 1.99976 7.99996 1.99976C8.32443 1.99976 8.63872 2.113 8.88863 2.31994L12.7993 5.53194C13.1366 5.80861 13.3333 6.22861 13.3333 6.67194V12.5399C13.3333 13.3466 12.6966 13.9999 11.9113 13.9999H4.08863Z"
          stroke={color}
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M6 10.6666C6.56667 11.0866 7.25667 11.3333 8 11.3333C8.74333 11.3333 9.43333 11.0866 10 10.6666M6.33333 7.66659V7.33325M9.66667 7.66659V7.33325"
          stroke={color}
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    route: "/admin/dashboard",
  },
  {
    name: "Clientes y sociedades",
    icon: (color: string) => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M9.66667 10.6666H12.3333C12.5101 10.6666 12.6797 10.5963 12.8047 10.4713C12.9298 10.3463 13 10.1767 13 9.99992V9.33325C13 8.80282 12.7893 8.29411 12.4142 7.91904C12.0391 7.54397 11.5304 7.33325 11 7.33325H9.66667M8.176 4.66659C8.44603 4.96849 8.80137 5.18129 9.19499 5.27683C9.58862 5.37236 10.002 5.34613 10.3804 5.2016C10.7587 5.05707 11.0843 4.80105 11.314 4.46743C11.5437 4.13381 11.6667 3.7383 11.6667 3.33325C11.6667 2.9282 11.5437 2.5327 11.314 2.19908C11.0843 1.86545 10.7587 1.60944 10.3804 1.46491C10.002 1.32038 9.58862 1.29414 9.19499 1.38968C8.80137 1.48521 8.44603 1.69801 8.176 1.99992M1 9.99992V9.33325C1 8.80282 1.21071 8.29411 1.58579 7.91904C1.96086 7.54397 2.46957 7.33325 3 7.33325H5.66667C6.1971 7.33325 6.70581 7.54397 7.08088 7.91904C7.45595 8.29411 7.66667 8.80282 7.66667 9.33325V9.99992C7.66667 10.1767 7.59643 10.3463 7.4714 10.4713C7.34638 10.5963 7.17681 10.6666 7 10.6666H1.66667C1.48986 10.6666 1.32029 10.5963 1.19526 10.4713C1.07024 10.3463 1 10.1767 1 9.99992ZM6.33333 3.33325C6.33333 3.86369 6.12262 4.37239 5.74755 4.74747C5.37247 5.12254 4.86377 5.33325 4.33333 5.33325C3.8029 5.33325 3.29419 5.12254 2.91912 4.74747C2.54405 4.37239 2.33333 3.86369 2.33333 3.33325C2.33333 2.80282 2.54405 2.29411 2.91912 1.91904C3.29419 1.54397 3.8029 1.33325 4.33333 1.33325C4.86377 1.33325 5.37247 1.54397 5.74755 1.91904C6.12262 2.29411 6.33333 2.80282 6.33333 3.33325Z"
          stroke={color}
          strokeLinecap="round"
        />
      </svg>
    ),
    route: "/admin/customers",
  },
  {
    name: "Lotes",
    icon: (color: string) => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M3.33671 7.46V8L8.00204 10.6947L12.6667 8V7.46M3.33337 10.7667V11.3067L7.99804 14L12.6634 11.3053V10.7653M8.00204 2L3.33671 4.69467L8.00204 7.38933L12.6667 4.69467L8.00204 2Z"
          stroke={color}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    route: "/admin/lots",
  },
  {
    name: "Órdenes de trabajo",
    icon: (color: string) => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M6.66671 2V4.66667C6.66671 4.84348 6.59647 5.01305 6.47144 5.13807C6.34642 5.2631 6.17685 5.33333 6.00004 5.33333H3.33337M8.66671 4H10.6667M8.66671 6H10.6667M8.00004 8V12M10.6667 10H5.33337M12.6667 2.66667V13.3333C12.6667 13.5101 12.5965 13.6797 12.4714 13.8047C12.3464 13.9298 12.1769 14 12 14H4.00004C3.82323 14 3.65366 13.9298 3.52864 13.8047C3.40361 13.6797 3.33337 13.5101 3.33337 13.3333V5.276C3.33341 5.0992 3.40367 4.92966 3.52871 4.80467L6.13804 2.19533C6.26304 2.0703 6.43258 2.00004 6.60937 2H12C12.1769 2 12.3464 2.07024 12.4714 2.19526C12.5965 2.32029 12.6667 2.48986 12.6667 2.66667ZM5.33337 8V12H10.6667V8H5.33337Z"
          stroke={color}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    route: "/admin/work-orders",
  },
  {
    name: "Labores",
    icon: (color: string) => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M8.66667 4.66659H12.6667L14 7.33325M8.66667 4.66659V9.99992M8.66667 4.66659V3.99992C8.66667 3.82311 8.59643 3.65354 8.4714 3.52851C8.34638 3.40349 8.17681 3.33325 8 3.33325H2.66667C2.48986 3.33325 2.32029 3.40349 2.19526 3.52851C2.07024 3.65354 2 3.82311 2 3.99992V9.99992H3.33333M14 7.33325V9.99992H12.6667M14 7.33325H10.6667M8.66667 9.99992H6M8.66667 9.99992H10M13 10.9999C13 11.4419 12.8244 11.8659 12.5118 12.1784C12.1993 12.491 11.7754 12.6666 11.3333 12.6666C10.8913 12.6666 10.4674 12.491 10.1548 12.1784C9.84226 11.8659 9.66667 11.4419 9.66667 10.9999C9.66667 10.5579 9.84226 10.134 10.1548 9.82141C10.4674 9.50885 10.8913 9.33325 11.3333 9.33325C11.7754 9.33325 12.1993 9.50885 12.5118 9.82141C12.8244 10.134 13 10.5579 13 10.9999ZM6.33333 10.9999C6.33333 11.4419 6.15774 11.8659 5.84518 12.1784C5.53262 12.491 5.10869 12.6666 4.66667 12.6666C4.22464 12.6666 3.80072 12.491 3.48816 12.1784C3.17559 11.8659 3 11.4419 3 10.9999C3 10.5579 3.17559 10.134 3.48816 9.82141C3.80072 9.50885 4.22464 9.33325 4.66667 9.33325C5.10869 9.33325 5.53262 9.50885 5.84518 9.82141C6.15774 10.134 6.33333 10.5579 6.33333 10.9999Z"
          stroke={color}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    route: "/admin/tasks",
  },
  {
    name: "Insumos",
    icon: (color: string) => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M6.55932 8.30298C6.55932 8.30298 7.39257 6.24417 6.41515 4.77629C5.43775 3.30841 3.05785 2.73935 2.932 2.82918C2.80615 2.91902 2.18662 5.04615 3.16402 6.51403C4.14142 7.98191 6.55932 8.30298 6.55932 8.30298ZM6.55932 8.30298C6.89264 8.96965 8.00004 10.0001 8.00004 12.0001V13.3334C8.00004 12.0001 7.71311 11.0542 9.37977 9.38758M9.37977 9.38758C9.37977 9.38758 9.05351 7.52838 10.1268 6.48233C11.2002 5.43628 13.1204 5.45111 13.2402 5.62614C13.3601 5.80117 13.528 7.52198 12.5196 8.50471C11.4463 9.55078 9.37977 9.38758 9.37977 9.38758Z"
          stroke={color}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    route: "/admin/products",
  },
  {
    name: "Stock",
    icon: (color: string) => (
      <svg
        width="20"
        height="20"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          d="M2 10.0001V12.6667M6 8.66675V12.6667M10 10.0001V12.6667M14 8.66675V12.6667M2 7.33341L6 4.00008L10 7.33341L13.6667 3.66675"
          stroke={color}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    ),
    route: "/admin/stock",
  },
];

interface SidebarProps {
  isSidebarOpen: boolean;
  setTitle: (title: string) => void;
  setIsSidebarOpen: () => void;
}

interface SidebarItemProps {
  item: MenuItem;
  setTitle: (title: string) => void;
  setIsSidebarOpen: () => void;
}

function SidebarItem({ item, setIsSidebarOpen, setTitle }: SidebarItemProps) {
  const location = useLocation();
  const isActive = (route: string) => location.pathname === route;
  const iconColor = isActive(item.route) ? "#0E9F6E" : "#64748B";

  useEffect(() => {
    if (isActive(item.route)) {
      setTitle(item.name);
    }
  }, [location.pathname, item, setTitle]);

  return (
    <Link
      to={item.route}
      className={`flex items-center w-full h-[34px] px-2.5 rounded-md gap-2 transition duration-200 ${
        isActive(item.route)
          ? "bg-custom-bg border border-[#CBD5E1] text-custom-green"
          : "text-gray-900 hover:bg-gray-100"
      } leading-5 font-medium`}
      onClick={() => {
        setTitle(item.name);
        if (window.innerWidth < 768) setIsSidebarOpen();
      }}
    >
      {item.icon(iconColor)}
      <span>{item.name}</span>
    </Link>
  );
}

interface SidebarSubmenuProps {
  setTitle: (title: string) => void;
  setIsSidebarOpen: () => void;
  item: {
    name: string;
    icon: React.ReactNode;
    children: SubItem[];
  };
}

interface SidebarSubItemProps {
  item: SubItem;
  setTitle: (title: string) => void;
  setIsSidebarOpen: () => void;
}

function SidebarSubmenuItem({
  item,
  setTitle,
  setIsSidebarOpen,
}: SidebarSubItemProps) {
  const location = useLocation();
  const isActive = (route: string) => location.pathname.startsWith(route);

  useEffect(() => {
    if (isActive(item.route)) {
      setTitle(item.name);
    }
  }, [location.pathname, item, setTitle]);

  return (
    <Link
      key={item.route}
      to={item.route}
      className={`flex items-center w-full h-[34px] pl-4 pr-4 rounded-md gap-2 text-sm leading-5 text-left transition duration-200 ${
        isActive(item.route)
          ? "bg-custom-bg border border-[#CBD5E1] text-custom-green font-bold"
          : "text-gray-900 hover:bg-gray-100"
      }`}
      onClick={() => {
        setTitle(item.name);
        if (window.innerWidth < 768) setIsSidebarOpen();
      }}
    >
      {item.name}
    </Link>
  );
}

function SidebarSubmenu({
  item,
  setTitle,
  setIsSidebarOpen,
}: SidebarSubmenuProps) {
  const location = useLocation();
  const [open, setOpen] = useState(
    item.children?.some((child) => location.pathname.startsWith(child.route)) ??
      false
  );

  if (item) {
    return (
      <div className="w-full">
        <button
          onClick={() => setOpen(!open)}
          className="flex items-center justify-between w-full h-[34px] px-1 rounded-md hover:bg-gray-100 transition duration-200"
        >
          <div className="flex items-center gap-2">
            {item.icon}
            <span className="text-gray-900 font-medium">{item.name}</span>
          </div>
          {open ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
        </button>
        {open && (
          <ul id="dropdown-menu" className="py-2 space-y-2">
            {item.children.map((subItem, index) => {
              return (
                <li key={index}>
                  <SidebarSubmenuItem
                    setTitle={setTitle}
                    item={subItem}
                    setIsSidebarOpen={setIsSidebarOpen}
                  />
                </li>
              );
            })}
          </ul>
        )}
      </div>
    );
  }
}

const Sidebar: React.FC<SidebarProps> = ({
  isSidebarOpen,
  setTitle,
  setIsSidebarOpen,
}) => {
  return (
    <aside
      id="logo-sidebar"
      className={`mt-2 bg-white border-r border-slate-300 transition-all duration-300 ease-in-out 
    ${
      isSidebarOpen
        ? "w-64 opacity-100 translate-x-0"
        : "w-0 opacity-0 -translate-x-full"
    }`}
      aria-label="Sidebar"
    >
      <div className="flex flex-col h-full pt-3 px-2 pb-3 gap-4 overflow-y-auto bg-white">
        <div className="flex items-center justify-between h-[30px] w-full px-2">
          <Link
            to="/admin/dashboard"
            onClick={() => {
              setTitle("Dashboard");
            }}
          >
            <h1 className="text-xl font-bold">Ponti Soft</h1>
          </Link>
          <button
            onClick={() => setIsSidebarOpen()}
            className="p-1 rounded-md hover:bg-gray-100"
          >
            <svg
              width="24"
              height="24"
              viewBox="0 0 18 18"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M6.75 2.625V15.375M12 11.25L9.75 9L12 6.75"
                stroke="#64748B"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
              <path
                d="M2.25 7.05C2.25 5.37 2.25 4.53 2.577 3.888C2.86462 3.32354 3.32354 2.86462 3.888 2.577C4.53 2.25 5.37 2.25 7.05 2.25H10.95C12.63 2.25 13.47 2.25 14.112 2.577C14.6765 2.86462 15.1354 3.32354 15.423 3.888C15.75 4.53 15.75 5.37 15.75 7.05V10.95C15.75 12.63 15.75 13.47 15.423 14.112C15.1354 14.6765 14.6765 15.1354 14.112 15.423C13.47 15.75 12.63 15.75 10.95 15.75H7.05C5.37 15.75 4.53 15.75 3.888 15.423C3.32354 15.1354 2.86462 14.6765 2.577 14.112C2.25 13.47 2.25 12.63 2.25 10.95V7.05Z"
                stroke="#64748B"
                strokeWidth="1.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </button>
        </div>
        <ul className="flex flex-col gap-2 font-medium">
          {menuItems.map((item) => {
            return (
              <li key={item.name}>
                <SidebarItem
                  setTitle={setTitle}
                  item={item}
                  setIsSidebarOpen={setIsSidebarOpen}
                />
              </li>
            );
          })}
        </ul>
        <SidebarSubmenu
          setTitle={() => setTitle("Informes")}
          item={menuReports}
          setIsSidebarOpen={setIsSidebarOpen}
        />
        <SidebarSubmenu
          setTitle={setTitle}
          item={menuDatabase}
          setIsSidebarOpen={setIsSidebarOpen}
        />
      </div>
    </aside>
  );
};

export default Sidebar;
