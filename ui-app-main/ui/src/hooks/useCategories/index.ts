import React, { useState } from "react";

import * as actions from "./actions";

import APIClient from "../../restclient/apiInstance";
import { TypeData, CategoryData } from "./types";
import { AxiosError } from "axios";
import { ErrorResponse, SuccessResponse } from "../../restclient/types";
import useCategoriesReducer from "./useCategoriesReducer";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useCategories = () => {
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [{ categories, types }, dispatch] = useCategoriesReducer();

  const getCategories = React.useCallback(
    async (queryString: string): Promise<void> => {
      setProcessing(true);
      setError(null);

      let queryParams = "";
      if (queryString !== "") {
        queryParams = `?${queryString}`;
      }

      try {
        const response = await request.get<SuccessResponse<CategoryData[]>>(
          "/categories" + queryParams
        );

        if (response.success) {
          dispatch({
            type: actions.SET_CATEGORIES,
            payload: response.data,
          });
          return;
        }

        setError("Ocurrio un error en la busqueda de categorías");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda de categorías.";

            setError(message);
            return;
          }
        }

        setError("Error en el servicio, inténtalo más tarde.");
      } finally {
        setProcessing(false);
      }
    },
    [dispatch]
  );

  const getTypes = React.useCallback(async (): Promise<void> => {
    setProcessing(true);
    setError(null);

    try {
      const response = await request.get<SuccessResponse<TypeData[]>>("/types");

      if (response.success) {
        dispatch({
          type: actions.SET_TYPES,
          payload: response.data,
        });
        return;
      }

      setError("Ocurrio un error en la busqueda de tipos");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const message =
            errorResponse.error.details ||
            "Error desconocido en la busqueda de categorías.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, [dispatch]);

  return {
    getCategories,
    getTypes,
    categories,
    types,
    processing,
    error,
  };
};

export default useCategories;
