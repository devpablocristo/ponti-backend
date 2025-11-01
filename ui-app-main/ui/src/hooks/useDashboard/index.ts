import React, { useState } from "react";
import { AxiosError } from "axios";

import useDashboardReducer from "./useDashboardReducer";
import * as actions from "./actions";
import { DashboardData } from "./types";
import { SuccessResponse, ErrorResponse } from "../../restclient/types";
import APIClient from "../../restclient/apiInstance";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useDashboard = () => {
  const [{ dashboard }, dispatch] = useDashboardReducer();
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getDashboardInfo = React.useCallback(async (queryString: string) => {
    setProcessing(true);
    setError(null);
    let queryParams = "";
    if (queryString !== "") {
      queryParams = `?${queryString}`;
    }

    try {
      const response = await request.get<SuccessResponse<DashboardData>>(
        `dashboard` + queryParams
      );

      if (response.success) {
        dispatch({
          type: actions.SET_DASHBOARD,
          payload: response.data,
        });
        return;
      }

      setError("Ocurrió un error en la búsqueda del dashboard");
    } catch (error) {
      dispatch({
        type: actions.SET_DASHBOARD,
        payload: null,
      });
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const message =
            errorResponse.error.details ||
            "Error desconocido en la búsqueda del dashboard.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, []);

  return {
    dashboard,
    processing,
    error,
    getDashboardInfo,
  };
};

export default useDashboard;
