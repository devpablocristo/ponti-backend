import React, { useState } from "react";
import { AxiosError } from "axios";

import useProductReducer from "./productsReducer";
import * as actions from "./actions";
import { Product, Supply, SupplyResponse } from "./types";
import { SuccessResponse, ErrorResponse } from "../../restclient/types";
import APIClient from "../../restclient/apiInstance";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useProducts = () => {
  const [{ products, result, supplies }, dispatch] = useProductReducer();
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errorUpdate, setErrorUpdate] = useState<string | null>(null);
  const [resultUpdate, setResultUpdate] = useState<string | null>(null);

  const getSupplies = React.useCallback(async (projectId: number) => {
    setProcessing(true);
    try {
      const response = await request.get<SuccessResponse<SupplyResponse>>(
        `/supplies/${projectId}`
      );

      if (response.success) {
        dispatch({
          type: actions.SET_SUPPLIES,
          payload: response.data.data,
        });
        return;
      }
      setError("Ocurrio un error en la busqueda de lotes");
    } catch (err) {
      const axiosError = err as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          if (errorResponse.error.status === 409) {
            setError("Ya existe un insumo con el mismo nombre.");
            return;
          }
          const message =
            errorResponse.error.details ||
            "Error desconocido en la creación de los insumos.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, []);

  const saveProducts = React.useCallback(
    async (products: Product[], projectId: number) => {
      setProcessing(true);
      setError(null);
      dispatch({
        type: actions.SET_RESULT,
        payload: "",
      });

      try {
        const response = await request.put<SuccessResponse<any>>(
          `/supplies/${projectId}`,
          products
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT,
            payload: "Se han creado los insumos con éxito!",
          });
          return;
        }

        setError("Ocurrio un error en la creación de los insumos");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            if (errorResponse.error.status === 409) {
              setError("Ya existe un insumo con el mismo nombre.");
              return;
            }
            const message =
              errorResponse.error.details ||
              "Error desconocido en la creación de los insumos.";

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

  const deleteSupply = React.useCallback(async (id: number) => {
    setProcessing(true);
    setError(null);
    dispatch({
      type: actions.SET_RESULT,
      payload: "",
    });

    try {
      const response = await request.delete<SuccessResponse<any>>(
        `/supplies/${id}`
      );

      if (response.success) {
        dispatch({
          type: actions.SET_RESULT,
          payload: "Se ha eliminado el insumo con éxito!",
        });
        return;
      }

      setError("Ocurrio un error en la eliminación del insumo");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          if (errorResponse.error.status === 409) {
            setError("El insumo esta siendo usado en una orden de trabajo.");
            return;
          }
          const message =
            errorResponse.error.details ||
            "Error desconocido en la eliminación del insumo.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, []);

  const updateSupply = React.useCallback(
    async (projectId: number, supply: Supply) => {
      setProcessing(true);
      setErrorUpdate(null);
      setResultUpdate(null);

      try {
        const response = await request.put<SuccessResponse<any>>(
          `/supplies/projects/${projectId}/${supply.id}`,
          supply
        );

        if (response.success) {
          setResultUpdate("Se editado el insumo con éxito!");
          return;
        }

        setErrorUpdate("Ocurrio un error en la edicion del insumo");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            if (errorResponse.error.status === 404) {
              setErrorUpdate("El insumo no existe.");
              return;
            }
            const message =
              errorResponse.error.details ||
              "Error desconocido en la edicion del insumo.";

            setErrorUpdate(message);
            return;
          }
        }

        setErrorUpdate("Error en el servicio, inténtalo más tarde.");
      } finally {
        setProcessing(false);
      }
    },
    []
  );

  return {
    products,
    supplies,
    getSupplies,
    saveProducts,
    updateSupply,
    deleteSupply,
    processing,
    error,
    result,
    errorUpdate,
    resultUpdate,
  };
};

export default useProducts;
