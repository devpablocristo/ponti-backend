import React from "react";
import { AxiosError } from "axios";

import * as actions from "./actions";
import APIClient from "../../restclient/apiInstance";
import { Payload } from "./types";
import { ErrorResponse, SuccessResponse } from "../../restclient/types";
import useFieldsReducer from "./useFieldsReducer";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useFields = () => {
  const [{ total, fields, processing, error }, dispatch] = useFieldsReducer();

  const getFields = React.useCallback(
    async (queryString: string): Promise<void> => {
      dispatch({ type: actions.SET_ERROR, payload: "" });
      dispatch({ type: actions.START_PROCESSING });

      let queryParams = "";
      if (queryString !== "") {
        queryParams = `?${queryString}`;
      }

      try {
        const response = await request.get<SuccessResponse<Payload>>(
          "/fields" + queryParams
        );

        if (response.success) {
          dispatch({
            type: actions.SET_FIELDS,
            payload: response.data.data,
          });

          dispatch({
            type: actions.SET_TOTAL,
            payload: response.data.total,
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error en la busqueda de campos",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda de campos.";

            dispatch({
              type: actions.SET_ERROR,
              payload: message,
            });
            return;
          }
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Error en el servicio, inténtalo más tarde.",
        });
      } finally {
        dispatch({ type: actions.STOP_PROCESSING });
      }
    },
    [dispatch]
  );

  return {
    getFields,
    total,
    fields,
    processing,
    error,
  };
};

export default useFields;
