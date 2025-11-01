import React from "react";

import * as actions from "./actions";
import APIClient from "../../restclient/apiInstance";
import { UserData, UserNew } from "./types";
import { ErrorResponse, SuccessResponse } from "../../restclient/types";
import { AxiosError } from "axios";
import useUserReducer from "./userReducer";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useUsers = () => {
  const [
    { users, selectedUser, error, processing, result, deleteError, isDeleting },
    dispatch,
  ] = useUserReducer();

  const saveUser = React.useCallback(
    async (userData: UserNew): Promise<void> => {
      dispatch({ type: actions.SET_RESULT, payload: "" });
      dispatch({ type: actions.CLEAR_ERROR });
      dispatch({ type: actions.START_PROCESSING });

      try {
        const response = await request.post<SuccessResponse<UserData>>(
          "/users",
          userData
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT,
            payload: "Usuario agregado con exito",
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error al intentar guardar un usuario",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            // const status = errorResponse.error.status;
            const message =
              errorResponse.error.details || "Error desconocido en el login.";

            dispatch({
              type: actions.SET_ERROR,
              payload: message,
            });
            return;
          }
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Error en el servicio, inténtalo más tarde",
        });
      } finally {
        dispatch({ type: actions.STOP_PROCESSING });
      }
    },
    [dispatch]
  );

  const getUsers = React.useCallback(
    async (queryString: string): Promise<void> => {
      dispatch({ type: actions.CLEAR_ERROR });
      dispatch({ type: actions.START_PROCESSING });

      let queryParams = "";
      if (queryString !== "") {
        queryParams = `?${queryString}`;
      }

      try {
        const response = await request.get<SuccessResponse<UserData[]>>(
          "/users" + queryParams
        );

        if (response.success) {
          dispatch({
            type: actions.SET_USERS,
            payload: response.data,
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error en la busueda de usuarios",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busueda de usuarios.";

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

  const getUser = React.useCallback(
    async (id: number): Promise<void> => {
      dispatch({ type: actions.SET_RESULT, payload: "" });
      dispatch({ type: actions.CLEAR_SELECTED_USER });
      dispatch({ type: actions.CLEAR_ERROR });
      dispatch({ type: actions.START_PROCESSING });

      try {
        const response = await request.get<SuccessResponse<UserData>>(
          "/users/" + id
        );

        if (response.success) {
          dispatch({
            type: actions.SET_SELECTED_USER,
            payload: response.data,
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error en la busueda del usuario",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busueda de usuarios.";

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

  const updateUser = React.useCallback(
    async (id: number, user: UserNew): Promise<void> => {
      dispatch({ type: actions.SET_RESULT, payload: "" });
      dispatch({ type: actions.CLEAR_ERROR });
      dispatch({ type: actions.START_PROCESSING });

      try {
        const response = await request.put<SuccessResponse<UserData>>(
          "/users/" + id,
          user
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT,
            payload: "Usuario editado con exito",
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error en la busueda de usuarios",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busueda de usuarios.";

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

  const deleteUser = React.useCallback(
    async (id: number): Promise<void> => {
      dispatch({ type: actions.SET_RESULT, payload: "" });
      dispatch({ type: actions.SET_DELETE_ERROR, payload: "" });
      dispatch({ type: actions.START_DELETING });

      try {
        const response = await request.delete<SuccessResponse<UserData>>(
          "/users/" + id
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT,
            payload: "Usuario eliminado con exito",
          });
          return;
        }

        dispatch({
          type: actions.SET_DELETE_ERROR,
          payload: "Error al eliminar usuario",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busueda de usuarios.";

            dispatch({
              type: actions.SET_DELETE_ERROR,
              payload: message,
            });
            return;
          }
        }

        dispatch({
          type: actions.SET_DELETE_ERROR,
          payload: "Error en el servicio, inténtalo más tarde.",
        });
      } finally {
        dispatch({ type: actions.STOP_DELETING });
      }
    },
    [dispatch]
  );

  return {
    users,
    selectedUser,
    error,
    processing,
    deleteError,
    isDeleting,
    result,
    getUsers,
    getUser,
    saveUser,
    updateUser,
    deleteUser,
  };
};

export default useUsers;
