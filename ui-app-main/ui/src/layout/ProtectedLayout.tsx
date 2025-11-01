import { useEffect, useState } from "react";

import Navbar from "./Navbar/Navbar";
import Sidebar from "./Sidebar/Sidebar";
import { BaseModal } from "../components/Modal/BaseModal";
import { AuthProvider, useAuth } from "../pages/login/context/AuthProvider";
import { Outlet, useNavigate } from "react-router-dom";
// import Footer from "./Footer/Footer";
import LoadingScreen from "../components/LoadingScreen/LoadingScreen";
import { SelectionProvider } from "../pages/login/context/SelectionContext";

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const auth = useAuth();

  const isSmallScreen = window.innerWidth < 768;
  const [title, setTitle] = useState("Dashboard");
  const [isSidebarOpen, setIsSidebarOpen] = useState(!isSmallScreen);
  const [isLogoutModalOpen, setIsLogoutModalOpen] = useState(false);

  useEffect(() => {
    if (!auth?.loading && !auth?.isAuthenticated) {
      navigate("/login");
    }
  }, [auth?.isAuthenticated, auth?.loading, navigate]);

  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth < 768) {
        setIsSidebarOpen(false);
      } else {
        setIsSidebarOpen(true);
      }
    };

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  if (auth?.loading || auth.user === null)
    return <LoadingScreen title={["Cargando..."]} description={[""]} />;

  const toggleSidebar = (e?: React.MouseEvent) => {
    if (e?.currentTarget instanceof HTMLElement) {
      e.currentTarget.blur();
    }
    setIsSidebarOpen((prev) => !prev);
  };

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar
        setTitle={setTitle}
        setIsSidebarOpen={() => setIsSidebarOpen(false)}
        isSidebarOpen={isSidebarOpen}
      />
      <div className="flex flex-col flex-1 overflow-hidden">
        <Navbar
          title={title}
          username={auth.user.Username}
          isSidebarOpen={isSidebarOpen}
          toggleSidebar={toggleSidebar}
          setIsLogoutModalOpen={() => setIsLogoutModalOpen(true)}
        />
        <main
          id="main-scroll"
          className="flex-1 overflow-y-auto py-1 p-4 bg-custom-bg"
        >
          <Outlet />
          <BaseModal
            isOpen={isLogoutModalOpen}
            onClose={() => setIsLogoutModalOpen(false)}
            title="Cerrar sesión"
            message="¿Está seguro que desea salir?"
            primaryButtonText="Sí, cerrar sesión"
            secondaryButtonText="No, cancelar"
            onPrimaryAction={() => {
              auth.logout();
              setIsLogoutModalOpen(true);
            }}
          />
        </main>
      </div>
      {/* <Footer /> */}
    </div>
  );
};

export const ProtectedLayout = () => {
  return (
    <AuthProvider>
      <SelectionProvider>
        <MainLayout />
      </SelectionProvider>
    </AuthProvider>
  );
};
