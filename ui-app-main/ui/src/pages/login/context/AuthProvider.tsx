import {
  createContext,
  useContext,
  useState,
  ReactNode,
  useCallback,
  useEffect,
  useMemo,
} from "react";
import { useNavigate } from "react-router-dom";
import { jwtDecode } from "jwt-decode";

import { DecodedToken, UserData } from "../types";
import { AuthService } from "../authService";
import {
  clearLocalStorage,
  getRefreshToken,
  setAccessToken,
  setLocalStorage,
} from "./useLocalStorage";
import { RequestError } from "../../../restclient/types";

interface AuthContextType {
  loading: boolean;
  isAuthenticated: boolean;
  user: DecodedToken | null;
  login: (data: UserData) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<DecodedToken | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  const verifyToken = useCallback(async () => {
    const accessToken = localStorage.getItem("access_token");
    if (!accessToken) {
      setIsAuthenticated(false);
      setLoading(false);
      return;
    }

    // try {
    //   await AuthService.validateToken();
    // } catch (error) {
    //   if (error instanceof RequestError && error.getStatus() === 401) {
    //     try {
    //       const accessToken = await AuthService.refreshToken();
    //       setAccessToken(accessToken);
    //     } catch (refreshError) {
    //       alert("Tu sesión ha expirado. Por favor, inicia sesión nuevamente");
    //       await logout();
    //       return;
    //     }
    //   } else {
    //     alert("Tu sesión ha expirado. Por favor, inicia sesión nuevamente");
    //     await logout();
    //     return;
    //   }
    // }

    // if (accessToken) {
    const decoded = jwtDecode<DecodedToken>(accessToken);
    setUser(decoded);
    setIsAuthenticated(true);
    // }

    setLoading(false);
  }, []);

  useEffect(() => {
    verifyToken();
  }, [location.pathname, verifyToken]);

  const login = useCallback(
    async (loginData: UserData) => {
      try {
        const { token, user } = await AuthService.login(loginData);

        setLocalStorage(token);
        setUser(user);
        setIsAuthenticated(true);

        navigate("/workspace");
      } catch (error) {
        throw error;
      }
    },
    [navigate]
  );

  const logout = useCallback(async () => {
    const refresh = getRefreshToken();
    if (!refresh) return;

    try {
      setLoading(true);
      await AuthService.logout(refresh);
      clearLocalStorage();
      setUser(null);
      setIsAuthenticated(false);
      navigate("/login");
    } catch (error) {
      if (error instanceof RequestError && error.getStatus() === 401) {
        try {
          const accessToken = await AuthService.refreshToken();
          setAccessToken(accessToken);

          await logout();
        } catch (refreshError) {
          throw refreshError;
        }
      } else {
        throw error;
      }
    } finally {
      setLoading(false);
    }
  }, [navigate]);

  const value = useMemo(
    () => ({
      isAuthenticated,
      loading,
      user,
      login,
      logout,
    }),
    [isAuthenticated, loading, user, login, logout]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth debe usarse dentro de un AuthProvider");
  }
  return context;
}
