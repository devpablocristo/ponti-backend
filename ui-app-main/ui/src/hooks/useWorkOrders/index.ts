import React, { useState } from "react";
import { AxiosError } from "axios";
import APIClient from "../../restclient/apiInstance";

import * as actions from "./actions";
import useOrdersReducer from "./ordersReducer";
import { SuccessResponse, ErrorResponse } from "../../restclient/types";
import { Metrics, Workorder, WorkorderData } from "./types";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useOrders = () => {
  const [
    { orders, pageInfo, resultCreation, selectedOrder, metrics },
    dispatch,
  ] = useOrdersReducer();
  const [processing, setProcessing] = useState(false);
  const [processingCreation, setProcessingCreation] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errorCreation, setErrorCreation] = useState<string | null>(null);

  const [processingMetrics, setProcessingMetrics] = useState(false);
  const [errorMetrics, setErrorMetrics] = useState<string | null>(null);

  const getOrders = React.useCallback(
    async (queryString: string): Promise<void> => {
      setProcessing(true);
      setError(null);

      let queryParams = "";
      if (queryString !== "") {
        queryParams = `?${queryString}`;
      }

      try {
        const response = await request.get<SuccessResponse<any>>(
          `/workorders${queryParams}`
        );

        if (response.success) {
          dispatch({
            type: actions.SET_ORDERS,
            payload: response.data.data,
          });

          dispatch({
            type: actions.SET_PAGE_INFO,
            payload: {
              page: response.data.page_info.page,
              per_page: response.data.page_info.per_page,
              total: response.data.page_info.total,
              max_page: response.data.page_info.max_page,
            },
          });
          return;
        }
        setError("Ocurrio un error en la busqueda de ordenes");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda de ordenes.";

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

  const getMetrics = React.useCallback(
    async (queryString: string) => {
      setProcessingMetrics(true);
      setErrorMetrics(null);
      let queryParams = "";
      if (queryString !== "") {
        queryParams = `?${queryString}`;
      }

      try {
        const response = await request.get<SuccessResponse<Metrics>>(
          "/workorders/metrics" + queryParams
        );

        if (response.success) {
          dispatch({
            type: actions.SET_METRICS,
            payload: response.data,
          });
          return;
        }

        setErrorMetrics("Ocurrio un error en la busqueda de kpis");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda de metricas.";

            setErrorMetrics(message);
            return;
          }
        }

        setErrorMetrics("Error en el servicio, inténtalo más tarde.");
      } finally {
        setProcessingMetrics(false);
      }
    },
    [dispatch]
  );

  const saveOrder = React.useCallback(async (order: Workorder) => {
    setProcessingCreation(true);
    setErrorCreation(null);
    dispatch({
      type: actions.SET_RESULT_CREATION,
      payload: "",
    });

    try {
      const response = await request.post<SuccessResponse<any>>(
        `/workorders`,
        order
      );

      if (response.success) {
        dispatch({
          type: actions.SET_RESULT_CREATION,
          payload: "Se ha creado la orden con éxito!",
        });
        return;
      }

      setErrorCreation("Ocurrio un error en la creación de la orden");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          if (errorResponse.error.status === 409) {
            setErrorCreation("Ya existe una orden con el mismo número.");
            return;
          }
          const message =
            errorResponse.error.details ||
            "Error desconocido en la creación de la orden.";

          setErrorCreation(message);
          return;
        }
      }

      setErrorCreation("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessingCreation(false);
    }
  }, []);

  const updateOrder = React.useCallback(
    async (id: number, order: Workorder) => {
      setProcessingCreation(true);
      setErrorCreation(null);
      dispatch({
        type: actions.SET_RESULT_CREATION,
        payload: "",
      });

      try {
        const response = await request.put<SuccessResponse<any>>(
          `/workorders/${id}`,
          order
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT_CREATION,
            payload: "Se ha actualizado la orden con éxito!",
          });
          return;
        }

        setErrorCreation("Ocurrio un error en la actualización de la orden");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la actualización de la orden.";

            setErrorCreation(message);
            return;
          }
        }

        setErrorCreation("Error en el servicio, inténtalo más tarde.");
      } finally {
        setProcessingCreation(false);
      }
    },
    []
  );

  const getWorkorder = React.useCallback(async (id: number) => {
    try {
      const response = await request.get<SuccessResponse<WorkorderData>>(
        `/workorders/${id}`
      );

      if (response.success) {
        dispatch({
          type: actions.SET_SELECTED_ORDER,
          payload: response.data,
        });
        return;
      }

      setErrorCreation("Ocurrio un error en la creación de la orden");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          if (errorResponse.error.status === 404) {
            setErrorCreation("No se encontro la orden.");
            return;
          }
          const message =
            errorResponse.error.details ||
            "Error desconocido en la busqueda de la orden.";

          setErrorCreation(message);
          return;
        }
      }

      setErrorCreation("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessingCreation(false);
    }
  }, []);

  const deleteOrder = React.useCallback(async (id: number): Promise<void> => {
    setProcessing(true);
    setError(null);

    try {
      const response = await request.delete<SuccessResponse<string>>(
        "/workorders/" + id
      );

      if (response.success) {
        return;
      }

      setError("Ocurrio un error al intentar eliminar una orden.");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const message =
            errorResponse.error.details ||
            "Error desconocido al intentar eliminar un proyecto.";

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
    orders,
    metrics,
    processingMetrics,
    errorMetrics,
    getOrders,
    getMetrics,
    saveOrder,
    updateOrder,
    getWorkorder,
    deleteOrder,
    selectedOrder,
    resultCreation,
    processing,
    error,
    processingCreation,
    errorCreation,
    pageInfo,
  };
};

export default useOrders;
