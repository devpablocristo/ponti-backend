import axios, { AxiosInstance, InternalAxiosRequestConfig } from "axios";
import {
  getAccessToken,
  getRefreshToken,
  clearLocalStorage,
  setAccessToken,
} from "../pages/login/context/useLocalStorage";

declare module "axios" {
  export interface InternalAxiosRequestConfig {
    _retry?: boolean;
  }
}

const refreshTokenUrl = "/auth/access-token";

type HttpClientOptions = {
  baseURL?: string;
  timeout?: number;
};

class APIClient {
  private accessToken: string | null;
  private client: AxiosInstance;

  constructor(options: HttpClientOptions) {
    this.accessToken = null;
    this.client = axios.create({
      baseURL: options.baseURL || "/api",
      timeout: options.timeout || 30000,
    });

    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        if (config.url === refreshTokenUrl) {
          const refreshToken = getRefreshToken();
          if (refreshToken) {
            config.headers.Authorization = `Bearer ${refreshToken}`;
          }
        } else {
          this.accessToken = getAccessToken();
          if (this.accessToken) {
            config.headers.Authorization = `Bearer ${this.accessToken}`;
          }
        }

        if (!config.headers["Content-Type"]) {
          config.headers["Content-Type"] = "application/json";
        }

        return config;
      }
    );

    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;

          try {
            const newAccessToken = await this.refreshAccessToken();
            this.accessToken = newAccessToken;

            originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
            return this.client(originalRequest);
          } catch (refreshError) {
            clearLocalStorage();
            window.location.href = "/login";
            return Promise.reject(refreshError);
          }
        }

        return Promise.reject(error);
      }
    );
  }

  private async refreshAccessToken(): Promise<string> {
    try {
      const {
        data: { access_token },
      } = (await this.client.get(refreshTokenUrl)).data;
      setAccessToken(access_token);

      return access_token;
    } catch (error) {
      throw new Error("No se pudo refrescar el token");
    }
  }

  async get<T>(
    endpoint: string,
    params?: object,
    options?: { responseType?: "json" | "blob" | "arraybuffer" }
  ): Promise<T> {
    const response = await this.client.get(endpoint, {
      params,
      responseType: options?.responseType ?? "json",
    });
    return response.data;
  }

  async post<T>(endpoint: string, data: unknown): Promise<T> {
    let headers = {};

    if (data instanceof FormData) {
      headers = { "Content-Type": "multipart/form-data" };
    }

    const response = await this.client.post(endpoint, data, { headers });
    return response.data;
  }

  async put<T>(endpoint: string, data: unknown): Promise<T> {
    let headers = {};

    if (data instanceof FormData) {
      headers = { "Content-Type": "multipart/form-data" };
    }

    const response = await this.client.put(endpoint, data, { headers });
    return response.data;
  }

  async delete<T>(endpoint: string, params?: object): Promise<T> {
    const response = await this.client.delete(endpoint, { params });
    return response.data;
  }
}

export default APIClient;
