import React, { useState } from "react";
import { AxiosError } from "axios";
import APIClient from "../../restclient/apiInstance";

import useStockReducer from "./useStockReducer";
import * as actions from "./actions";
import { SuccessResponse, ErrorResponse } from "../../restclient/types";
import { GetStocksResponse } from "./types";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useStock = () => {
  const [{ currentPage, stock, summary }, dispatch] = useStockReducer();
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [processingStock, setProcessingStock] = useState(false);
  const [errorStock, setErrorStock] = useState<string | null>(null);
  const [resultStock, setResultStock] = useState<string | null>(null);

  const [processingCloseStock, setProcessingCloseStock] = useState(false);
  const [errorCloseStock, setErrorCloseStock] = useState<string | null>(null);
  const [resultCloseStock, setResultCloseStock] = useState<string | null>(null);

  const [processingPeriods, setProcessingPeriods] = useState(false);
  const [errorPeriods, setErrorPeriods] = useState<string | null>(null);
  const [periods, setPeriods] = useState<string[] | null>(null);

  const getStock = React.useCallback(
    async (projectId: number, cutOffDate: string): Promise<void> => {
      setProcessing(true);
      setError(null);

      try {
        const response = await request.get<SuccessResponse<GetStocksResponse>>(
          `/stock/${projectId}?cutoff_date=${cutOffDate}`
        );

        if (response.success) {
          dispatch({
            type: actions.SET_SUMMARY,
            payload: {
              total_kg: response.data.total_kilograms,
              total_lt: response.data.total_liters,
              total_usd: response.data.net_total_usd,
            },
          });

          dispatch({
            type: actions.SET_STOCK,
            payload: response.data.items,
          });
          return;
        }
        setError("Ocurrio un error en la busqueda de STOCK");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda de stock.";

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

  const getPeriods = React.useCallback(
    async (projectId: number): Promise<void> => {
      setProcessingPeriods(true);
      setErrorPeriods(null);

      try {
        const response = await request.get<SuccessResponse<string[]>>(
          `/stock/periods/${projectId}`
        );

        if (response.success) {
          setPeriods(response.data);
          return;
        }
        setErrorPeriods("Ocurrio un error en la busqueda de PERIODOS");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda de stock.";

            setErrorPeriods(message);
            return;
          }
        }

        setErrorPeriods("Error en el servicio, inténtalo más tarde.");
      } finally {
        setProcessingPeriods(false);
      }
    },
    []
  );

  const updateStock = React.useCallback(
    async (projectId: number, id: number, realStock: number) => {
      setProcessingStock(true);
      setErrorStock(null);
      setResultStock(null);

      try {
        const response = await request.put<SuccessResponse<any>>(
          `/stock/${projectId}/${id}`,
          { real_stock_units: realStock }
        );

        if (response.success) {
          setResultStock("Se han actualizado el stock con éxito");
          return;
        }

        setErrorStock("Ocurrio un error en la modificacion del stock");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la modificacion del stock.";

            setErrorStock(message);
            return;
          }
        }

        setErrorStock("Error en el servicio, inténtalo más tarde.");
      } finally {
        setProcessingStock(false);
      }
    },
    [dispatch]
  );

  const closeStock = React.useCallback(
    async (projectId: number, closeDate: string) => {
      setProcessingCloseStock(true);
      setErrorCloseStock(null);
      setResultCloseStock(null);

      try {
        const response = await request.put<SuccessResponse<any>>(
          `/stock/close/${projectId}`,
          { close_date: closeDate }
        );

        if (response.success) {
          setResultCloseStock("Se ha cerrado el stock con éxito");
          return;
        }

        setErrorCloseStock("Ocurrio un error en el cierre del stock");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en el cierre del stock.";

            setErrorCloseStock(message);
            return;
          }
        }

        setErrorCloseStock("Error en el servicio, inténtalo más tarde.");
      } finally {
        setProcessingCloseStock(false);
      }
    },
    [dispatch]
  );

  return {
    stock,
    currentPage: currentPage,
    getStock,
    processing,
    error,
    summary,
    updateStock,
    processingStock,
    errorStock,
    resultStock,
    closeStock,
    processingCloseStock,
    errorCloseStock,
    resultCloseStock,
    getPeriods,
    processingPeriods,
    errorPeriods,
    periods,
  };
};

export default useStock;
