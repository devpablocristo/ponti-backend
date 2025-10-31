import { AxiosError } from "axios";
import { jwtDecode } from "jwt-decode";

import APIClient from "../../restclient/apiInstance";
import { DecodedToken, LogoutData, TokenResponse, UserData } from "./types";
import {
  ErrorResponse,
  RequestError,
  SuccessResponse,
} from "../../restclient/types";

const restClient = new APIClient({
  baseURL: "/api",
});

export class AuthService {
  static async login(
    userData: UserData
  ): Promise<{ token: TokenResponse; user: DecodedToken }> {
    try {
      const response = await restClient.post<SuccessResponse<TokenResponse>>(
        "/auth/login",
        userData
      );

      if (
        response.success &&
        response.data?.access_token &&
        response.data?.refresh_token
      ) {
        const decoded = jwtDecode<DecodedToken>(response.data.access_token);
        return { token: response.data, user: decoded };
      }

      throw new Error("Respuesta inválida del servidor");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const status = errorResponse.error.status;
          const message =
            errorResponse.error.details || "Error desconocido en el login.";
          throw new RequestError(status, message);
        }
      }

      throw new RequestError(500, "Error en el servicio, inténtalo más tarde.");
    }
  }

  static async logout(refreshToken: string): Promise<void> {
    const logoutData: LogoutData = {
      refresh_token: refreshToken,
    };

    try {
      await restClient.post<any>("/auth/logout", logoutData);
      return;
    } catch (error) {
      const axiosError = error as AxiosError;

      throw new RequestError(
        axiosError.status,
        "Ocurrió un error en la busqueda de datos, por favor intente mas tarde."
      );
    }
  }

  static async refreshToken(): Promise<string> {
    try {
      const response = await restClient.get<SuccessResponse<TokenResponse>>(
        "/auth/access-token"
      );

      if (!response.data.access_token) {
        throw new Error("No se recibió un nuevo access token.");
      }

      return response.data.access_token;
    } catch (error) {
      const axiosError = error as AxiosError;

      throw new RequestError(
        axiosError.status,
        "Error al refrescar el token, por favor intente más tarde."
      );
    }
  }

  static async validateToken(): Promise<void> {
    try {
      await restClient.get("/auth/session");
    } catch (error) {
      const axiosError = error as AxiosError;

      throw new RequestError(
        axiosError.status,
        "Ocurrió un error en la busqueda de datos, por favor intente mas tarde."
      );
    }
  }
}
