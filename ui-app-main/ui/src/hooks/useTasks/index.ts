import React, { useState } from "react";

import useTaskReducer from "./tasksReducer";
import * as actions from "./actions";
import { InvoiceData, Metrics, TaskInfo, TaskToSave } from "./types";
import { SuccessResponse, ErrorResponse } from "../../restclient/types";
import APIClient from "../../restclient/apiInstance";
import { AxiosError } from "axios";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useTask = () => {
  const [
    { tasks, labors, result, pageInfo, resultInvoice, metrics },
    dispatch,
  ] = useTaskReducer();
  const [processing, setProcessing] = useState(false);
  const [processingInvoice, setProcessingInvoice] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errorUpdate, setErrorUpdate] = useState<string | null>(null);
  const [resultUpdate, setResultUpdate] = useState<string | null>(null);
  const [errorInvoice, setErrorInvoice] = useState<string | null>(null);

  const [processingMetrics, setProcessingMetrics] = useState(false);
  const [errorMetrics, setErrorMetrics] = useState<string | null>(null);

  const getTasks = React.useCallback(
    async (projectId: number, query: string) => {
      setProcessing(true);
      setError(null);

      dispatch({
        type: actions.SET_TASKS,
        payload: [],
      });

      try {
        const response = await request.get<SuccessResponse<any>>(
          `/labors/${projectId}${query}`
        );

        if (response.success) {
          dispatch({
            type: actions.SET_TASKS,
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

        setError("Ocurrio un error en la busqueda de labores");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda de labores.";

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
    async (projectId: number, queryString: string) => {
      setProcessingMetrics(true);
      setErrorMetrics(null);

      try {
        const response = await request.get<SuccessResponse<Metrics>>(
          `/labors/metrics/${projectId}` + queryString
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

  const createInvoice = React.useCallback(async (invoice: InvoiceData) => {
    setProcessingInvoice(true);
    setErrorInvoice(null);
    dispatch({
      type: actions.SET_RESULT_INVOICE,
      payload: "",
    });

    try {
      const response = await request.post<SuccessResponse<any>>(
        `/labors/invoice`,
        invoice
      );

      if (response.success) {
        dispatch({
          type: actions.SET_RESULT_INVOICE,
          payload: "Se ha creado la factura con éxito!",
        });
        return;
      }

      setErrorInvoice("Ocurrio un error en la creación de la factura");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const message =
            errorResponse.error.details ||
            "Error desconocido en la creación de la factura.";

          setErrorInvoice(message);
          return;
        }
      }

      setErrorInvoice("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessingInvoice(false);
    }
  }, []);

  const updateInvoice = React.useCallback(
    async (id: number, invoice: InvoiceData) => {
      setProcessingInvoice(true);
      setErrorInvoice(null);
      dispatch({
        type: actions.SET_RESULT_INVOICE,
        payload: "",
      });

      try {
        const response = await request.put<SuccessResponse<any>>(
          `/labors/invoice/${id}`,
          invoice
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT_INVOICE,
            payload: "Se ha actualizado la factura con éxito!",
          });
          return;
        }

        setErrorInvoice("Ocurrio un error en la actualización de la factura");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la actualización de la factura.";

            setErrorInvoice(message);
            return;
          }
        }

        setErrorInvoice("Error en el servicio, inténtalo más tarde.");
      } finally {
        setProcessingInvoice(false);
      }
    },
    []
  );

  const saveTasks = React.useCallback(
    async (tasks: TaskToSave[], projectId: number) => {
      setProcessing(true);
      setError(null);
      dispatch({
        type: actions.SET_RESULT,
        payload: "",
      });

      try {
        const response = await request.post<SuccessResponse<any>>(
          `/projects/${projectId}/labors`,
          tasks
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT,
            payload: "Se han creado los labores con éxito!",
          });
          return;
        }

        setError("Ocurrio un error en la creación de los labores");
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

  const getLabors = React.useCallback(async (projectId: number) => {
    setProcessing(true);
    setError(null);
    dispatch({
      type: actions.SET_RESULT,
      payload: "",
    });

    try {
      const response = await request.get<SuccessResponse<any>>(
        `/projects/${projectId}/labors`
      );

      if (response.success) {
        dispatch({
          type: actions.SET_LABORS,
          payload: response.data.data,
        });
        return;
      }

      setError("Ocurrio un error en la busqueda de labores");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const message =
            errorResponse.error.details ||
            "Error desconocido en la busqueda de labores.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, []);

  const deleteLabor = React.useCallback(async (id: number) => {
    setProcessing(true);
    setError(null);
    dispatch({
      type: actions.SET_RESULT,
      payload: "",
    });

    try {
      const response = await request.delete<SuccessResponse<any>>(
        `/labors/${id}`
      );

      if (response.success) {
        dispatch({
          type: actions.SET_RESULT,
          payload: "Se ha eliminado el labor con éxito!",
        });
        return;
      }

      setError("Ocurrio un error en la eliminación del labor");
    } catch (error) {
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          if (errorResponse.error.status === 409) {
            setError("El labor esta siendo usado en una orden de trabajo.");
            return;
          }
          const message =
            errorResponse.error.details ||
            "Error desconocido en la eliminación del labor.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, []);

  const updateLabor = React.useCallback(
    async (projectId: number, labor: TaskInfo) => {
      setProcessing(true);
      setErrorUpdate(null);
      setResultUpdate(null);

      try {
        const response = await request.put<SuccessResponse<any>>(
          `/labors/projects/${projectId}/${labor.id}`,
          labor
        );

        if (response.success) {
          setResultUpdate("Se editado el labor con éxito!");
          return;
        }

        setErrorUpdate("Ocurrio un error en la edicion del labor");
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            if (errorResponse.error.status === 404) {
              setErrorUpdate("El labor no existe.");
              return;
            }
            const message =
              errorResponse.error.details ||
              "Error desconocido en la edicion del labor.";

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
    tasks,
    metrics,
    getTasks,
    getMetrics,
    getLabors,
    deleteLabor,
    updateLabor,
    saveTasks,
    updateInvoice,
    createInvoice,
    result,
    resultUpdate,
    resultInvoice,
    labors,
    processing,
    error,
    errorUpdate,
    errorInvoice,
    processingInvoice,
    pageInfo,
    processingMetrics,
    errorMetrics,
  };
};

export default useTask;
