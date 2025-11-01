import axios, { AxiosError, AxiosInstance, AxiosRequestConfig } from "axios";

export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
  error?: ErrorResponse;
}

interface ErrorResponse {
  status: number;
  details: string;
}

export class ApiClient {
  private axiosInstance: AxiosInstance;

  constructor(baseURL: string) {
    this.axiosInstance = axios.create({
      baseURL,
      timeout: 8000,
    });

    this.axiosInstance.interceptors.response.use(
      (response) => response,
      this.handleErrorResponse
    );
  }

  private handleErrorResponse(error: AxiosError): Promise<ApiResponse<null>> {
    return Promise.reject<ApiResponse<null>>({
      success: false,
      message: "Error en la solicitud",
      error: {
        status: error.response?.status ?? 500,
        details:
          (error.response?.data as { error?: string })?.error ||
          "Detalles no disponibles",
      },
    });
  }

  public async get<T>(
    url: string,
    //headers?: Record<string, string | undefined>
    options?: Record<string, string | undefined> | AxiosRequestConfig
  ): Promise<ApiResponse<T>> {
    try {
      //const config: AxiosRequestConfig = headers ? { headers } : {};
      let config: AxiosRequestConfig = {};

      if (options) {
        if ("headers" in options || "responseType" in options) {
          config = options as AxiosRequestConfig;
        } else {
          config = { headers: options as Record<string, string | undefined> };
        }
      }

      const response = await this.axiosInstance.get<T>(url, config);
      return {
        success: true,
        message: "Operación exitosa",
        data: response.data,
      };
    } catch (error) {
      throw error as ApiResponse<null>;
    }
  }

  public async delete<T>(
    url: string,
    headers?: Record<string, string | undefined>
  ): Promise<ApiResponse<T>> {
    try {
      const config: AxiosRequestConfig = headers ? { headers } : {};
      const response = await this.axiosInstance.delete<T>(url, config);
      return {
        success: true,
        message: "Operación exitosa",
        data: response.data,
      };
    } catch (error) {
      throw error as ApiResponse<null>;
    }
  }

  public async post<T>(
    url: string,
    data: any,
    headers?: Record<string, string | undefined>
  ): Promise<ApiResponse<T>> {
    try {
      const config: AxiosRequestConfig = headers ? { headers } : {};
      const response = await this.axiosInstance.post<T>(url, data, config);
      return {
        success: true,
        message: "Operación exitosa",
        data: response.data,
      };
    } catch (error) {
      throw error as ApiResponse<null>;
    }
  }

  public async put<T>(
    url: string,
    data: any,
    headers?: Record<string, string | undefined>
  ): Promise<ApiResponse<T>> {
    try {
      const config: AxiosRequestConfig = headers ? { headers } : {};
      const response = await this.axiosInstance.put<T>(url, data, config);
      return {
        success: true,
        message: "Operación exitosa",
        data: response.data,
      };
    } catch (error) {
      throw error as ApiResponse<null>;
    }
  }
}
