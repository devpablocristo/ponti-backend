import React, { useState } from "react";
import { AxiosError } from "axios";

import useDollarReducer from "./dollarReducer";
import * as actions from "./actions";
import { DollarData } from "./types";
import { SuccessResponse, ErrorResponse } from "../../restclient/types";
import APIClient from "../../restclient/apiInstance";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useDollar = () => {
  const [{ dollars, result }, dispatch] = useDollarReducer();
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getDollarInfo = React.useCallback(async (id: number) => {
    setProcessing(true);
    setError(null);
    dispatch({
      type: actions.SET_RESULT,
      payload: "",
    });

    try {
      const response = await request.get<SuccessResponse<DollarData[]>>(
        `/projects/${id}/dollar-values`
      );

      if (response.success) {
        dispatch({
          type: actions.SET_DOLLARS,
          payload: response.data,
        });
        return;
      }

      setError("Ocurrio un error en la busqueda de lotes");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const message =
            errorResponse.error.details ||
            "Error desconocido en la busqueda de campañas.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, []);

  const saveDollarInfo = React.useCallback(
    async (dollar: DollarData[], id: number) => {
      setProcessing(true);
      setError(null);
      dispatch({
        type: actions.SET_RESULT,
        payload: "",
      });

      try {
        const response = await request.put<SuccessResponse<any>>(
          `/projects/${id}/dollar-values`,
          dollar
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT,
            payload: "Se han creado los valores con éxito!",
          });
          return;
        }

        setError("Ocurrio un error en la creación de los valores");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            if (errorResponse.error.status === 409) {
              setError("Ya existe un valor con el mismo nombre.");
              return;
            }
            const message =
              errorResponse.error.details ||
              "Error desconocido en la creación de los valores.";

            setError(message);
            return;
          }
        }

        setError("Error en el servicio, inténtalo más tarde.");
      } finally {
        setProcessing(false);
      }
    },
    []
  );

  return {
    dollars,
    processing,
    error,
    getDollarInfo,
    saveDollarInfo,
    result,
  };
};

export default useDollar;
