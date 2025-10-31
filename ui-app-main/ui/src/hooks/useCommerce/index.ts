import React, { useState } from "react";
import { AxiosError } from "axios";

import useCommerceReducer from "./commerceReducer";
import * as actions from "./actions";
import { CommerceData, CommerceInfoData } from "./types";
import { SuccessResponse, ErrorResponse } from "../../restclient/types";
import APIClient from "../../restclient/apiInstance";

const request = new APIClient({
  timeout: 8000,
  baseURL: "/api",
});

const useCommerce = () => {
  const [{ result, commerceInfoList }, dispatch] = useCommerceReducer();
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getCommerceInfo = React.useCallback(async (id: number) => {
    setProcessing(true);
    setError(null);

    try {
      const response = await request.get<SuccessResponse<CommerceInfoData[]>>(
        `/projects/${id}/commerce`
      );

      if (response.success) {
        dispatch({
          type: actions.SET_COMMERCE_INFO_LIST,
          payload: response.data,
        });
        return;
      }

      setError("Ocurrio un error en la busqueda de los valores");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          if (errorResponse.error.status === 404) {
            dispatch({
              type: actions.SET_COMMERCE_INFO_LIST,
              payload: [],
            });
            return;
          }
          const message =
            errorResponse.error.details ||
            "Error desconocido en la busqueda de los valores.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, []);

  const saveCommerceInfo = React.useCallback(
    async (commerce: CommerceData[], id: number) => {
      setProcessing(true);
      setError(null);
      dispatch({
        type: actions.SET_RESULT,
        payload: "",
      });

      try {
        const response = await request.post<SuccessResponse<any>>(
          `/projects/${id}/commerce`,
          commerce
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
    processing,
    error,
    getCommerceInfo,
    saveCommerceInfo,
    result,
    commerceInfoList,
  };
};

export default useCommerce;
