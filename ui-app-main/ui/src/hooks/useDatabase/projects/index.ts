import React from "react";
import { AxiosError } from "axios";

import * as actions from "./actions";
import APIClient from "../../../restclient/apiInstance";
import { Project, ProjectPayload, ProjectDropdownPayload } from "./types";
import { ErrorResponse, SuccessResponse } from "../../../restclient/types";

import useProjectReducer from "./projectReducer";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useProjects = () => {
  const [
    {
      projects,
      totalHectares,
      projectsDropdown,
      projectsDropdownPagination,
      pageInfo,
      selectedProject,
      error,
      processing,
      processingDropdown,
      result,
    },
    dispatch,
  ] = useProjectReducer();

  const saveProject = React.useCallback(
    async (userData: Project): Promise<void> => {
      dispatch({ type: actions.SET_RESULT, payload: "" });
      dispatch({ type: actions.SET_ERROR, payload: "" });
      dispatch({ type: actions.START_PROCESSING });

      try {
        const response = await request.post<SuccessResponse<Project>>(
          "/projects",
          userData
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT,
            payload: "Se ha creado un nuevo proyecto con éxito!",
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error al intentar guardar el proyecto",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            if (errorResponse.error.status === 409) {
              dispatch({
                type: actions.SET_ERROR,
                payload: "Ya existe un proyecto con el mismo nombre y campaña.",
              });
              return;
            }
            const message =
              errorResponse.error.details ||
              "Error desconocido en la creación del proyecto.";

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

  const getProjects = React.useCallback(
    async (queryString: string): Promise<void> => {
      dispatch({ type: actions.SET_ERROR, payload: "" });
      dispatch({ type: actions.START_PROCESSING });

      let queryParams = "";
      if (queryString !== "") {
        queryParams = `?${queryString}`;
      }

      try {
        const response = await request.get<SuccessResponse<ProjectPayload>>(
          "/projects" + queryParams
        );

        if (response.success) {
          dispatch({
            type: actions.SET_PROJECTS,
            payload: response.data.data,
          });

          dispatch({
            type: actions.SET_PAGINATION,
            payload: response.data.page_info,
          });

          dispatch({
            type: actions.SET_TOTAL_HECTARES,
            payload: response.data.total_hectares,
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error en la busqueda de proyectos",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda de proyectos.";

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

  const getProjectsDropdown = React.useCallback(
    async (id: number, queryString: string = ""): Promise<void> => {
      dispatch({ type: actions.SET_ERROR, payload: "" });
      dispatch({ type: actions.START_PROCESSING_DROPDOWN });

      try {
        const response = await request.get<
          SuccessResponse<ProjectDropdownPayload>
        >(`/projects/customer/${id}` + (queryString ? `?${queryString}` : ""));

        if (response.success) {
          dispatch({
            type: actions.SET_PROJECTS_DROPDOWN,
            payload: response.data.data,
          });

          dispatch({
            type: actions.SET_PROJECTS_DROPDOWN_PAGINATION,
            payload: response.data.page_info,
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR_DROPDOWN,
          payload: "Ocurrio un error en la busqueda de proyectos",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda de proyectos.";

            dispatch({
              type: actions.SET_ERROR_DROPDOWN,
              payload: message,
            });
            return;
          }
        }

        dispatch({
          type: actions.SET_ERROR_DROPDOWN,
          payload: "Error en el servicio, inténtalo más tarde.",
        });
      } finally {
        dispatch({ type: actions.STOP_PROCESSING_DROPDOWN });
      }
    },
    [dispatch]
  );

  const getProject = React.useCallback(
    async (id: number): Promise<void> => {
      dispatch({ type: actions.SET_RESULT, payload: "" });
      dispatch({ type: actions.CLEAR_SELECTED_PROJECT });
      dispatch({ type: actions.SET_ERROR, payload: "" });
      dispatch({ type: actions.START_PROCESSING });

      try {
        const response = await request.get<SuccessResponse<Project>>(
          "/projects/" + id
        );

        if (response.success) {
          dispatch({
            type: actions.SET_SELECTED_PROJECT,
            payload: response.data,
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error en la busqueda del proyecto",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido en la busqueda del proyecto";

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

  const updateProject = React.useCallback(
    async (id: number, project: Project): Promise<void> => {
      dispatch({ type: actions.SET_RESULT, payload: "" });
      dispatch({ type: actions.SET_ERROR, payload: "" });
      dispatch({ type: actions.START_PROCESSING });

      try {
        const response = await request.put<SuccessResponse<Project>>(
          "/projects/" + id,
          project
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT,
            payload: "Proyecto editado con exito",
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error al intentar editar un proyecto.",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            if (errorResponse.error.status === 404) {
              dispatch({
                type: actions.SET_ERROR,
                payload:
                  "No se encontró el proyecto o no tiene la última versión disponible.",
              });
              return;
            }

            const message =
              errorResponse.error.details ||
              "Error desconocido al intentar editar un proyecto.";

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

  const deleteProject = React.useCallback(
    async (id: number): Promise<void> => {
      dispatch({ type: actions.SET_ERROR, payload: "" });
      dispatch({ type: actions.START_PROCESSING });

      try {
        const response = await request.delete<SuccessResponse<string>>(
          "/projects/" + id
        );

        if (response.success) {
          dispatch({
            type: actions.SET_RESULT,
            payload: "Proyecto eliminado con exito",
          });
          return;
        }

        dispatch({
          type: actions.SET_ERROR,
          payload: "Ocurrio un error al intentar eliminar un proyecto.",
        });
      } catch (error) {
        const axiosError = error as AxiosError;

        if (axiosError.response) {
          const errorResponse = axiosError.response.data as ErrorResponse;

          if (errorResponse.error) {
            const message =
              errorResponse.error.details ||
              "Error desconocido al intentar eliminar un proyecto.";

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
    projects,
    totalHectares,
    projectsDropdown,
    pageInfo,
    projectsDropdownPagination,
    selectedProject,
    error,
    processing,
    processingDropdown,
    result,
    getProjects,
    getProjectsDropdown,
    getProject,
    saveProject,
    updateProject,
    deleteProject,
  };
};

export default useProjects;
