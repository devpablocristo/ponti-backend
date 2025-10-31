import React from "react";

import * as actions from "./actions";
import { PageInfo, Project, ProjectData, ProjectDropdown } from "./types";

type ProjectState = {
  projects: ProjectData[];
  totalHectares: number;
  projectsDropdown: ProjectDropdown[];
  projectsDropdownPagination: PageInfo | null;
  pageInfo: PageInfo | null;
  selectedProject: Project | null;
  processing: boolean;
  processingDropdown: boolean;
  error: string;
  errorDropdown: string;
  result: string;
};

type ProjectAction =
  | { type: typeof actions.SET_PROJECTS; payload: ProjectData[] }
  | { type: typeof actions.SET_PROJECTS_DROPDOWN; payload: ProjectDropdown[] }
  | { type: typeof actions.SET_PROJECTS_DROPDOWN_PAGINATION; payload: PageInfo }
  | { type: typeof actions.SET_PAGINATION; payload: PageInfo }
  | { type: typeof actions.SET_SELECTED_PROJECT; payload: Project }
  | { type: typeof actions.CLEAR_SELECTED_PROJECT }
  | { type: typeof actions.SET_RESULT; payload: string }
  | { type: typeof actions.SET_ERROR; payload: string }
  | { type: typeof actions.SET_ERROR_DROPDOWN; payload: string }
  | { type: typeof actions.START_PROCESSING }
  | { type: typeof actions.STOP_PROCESSING }
  | { type: typeof actions.START_PROCESSING_DROPDOWN }
  | { type: typeof actions.STOP_PROCESSING_DROPDOWN }
  | { type: typeof actions.SET_TOTAL_HECTARES; payload: number };

const initialState: ProjectState = {
  projects: [],
  totalHectares: 0,
  projectsDropdown: [],
  projectsDropdownPagination: null,
  pageInfo: null,
  selectedProject: null,
  processing: false,
  processingDropdown: false,
  error: "",
  errorDropdown: "",
  result: "",
};

const projectReducer = (state: ProjectState, action: ProjectAction) => {
  switch (action.type) {
    case actions.SET_PROJECTS:
      return {
        ...state,
        projects: action.payload,
      };

    case actions.SET_PROJECTS_DROPDOWN:
      return {
        ...state,
        projectsDropdown: action.payload,
      };

    case actions.SET_PAGINATION:
      return {
        ...state,
        pageInfo: action.payload,
      };

    case actions.SET_TOTAL_HECTARES:
      return {
        ...state,
        totalHectares: action.payload,
      };

    case actions.SET_PROJECTS_DROPDOWN_PAGINATION:
      return {
        ...state,
        projectsDropdownPagination: action.payload,
      };

    case actions.SET_SELECTED_PROJECT:
      return {
        ...state,
        selectedProject: action.payload,
      };

    case actions.CLEAR_SELECTED_PROJECT:
      return {
        ...state,
        selectedProject: null,
      };

    case actions.SET_ERROR:
      return {
        ...state,
        error: action.payload,
        projects: [],
      };

    case actions.SET_ERROR_DROPDOWN:
      return {
        ...state,
        errorDropdown: action.payload,
        projectsDropdown: [],
      };

    case actions.SET_RESULT:
      return {
        ...state,
        result: action.payload,
      };

    case actions.START_PROCESSING:
      return {
        ...state,
        processing: true,
      };

    case actions.STOP_PROCESSING:
      return {
        ...state,
        processing: false,
      };

    case actions.START_PROCESSING_DROPDOWN:
      return {
        ...state,
        processingDropdown: true,
      };

    case actions.STOP_PROCESSING_DROPDOWN:
      return {
        ...state,
        processingDropdown: false,
      };

    default:
      return state;
  }
};

const useProjectReducer = () => React.useReducer(projectReducer, initialState);

export default useProjectReducer;
