import React from "react";

import * as actions from "./actions";
import { UserData } from "./types";

interface UserState {
  users: UserData[];
  selectedUser: UserData | null;
  processing: boolean;
  error: string;
  result: string;
  deleteError: string;
  isDeleting: boolean;
}

type UserAction =
  | { type: typeof actions.SET_USERS; payload: UserData[] }
  | { type: typeof actions.SET_SELECTED_USER; payload: UserData }
  | { type: typeof actions.CLEAR_SELECTED_USER }
  | { type: typeof actions.SET_ERROR; payload: string }
  | { type: typeof actions.SET_DELETE_ERROR; payload: string }
  | { type: typeof actions.CLEAR_ERROR }
  | { type: typeof actions.START_PROCESSING }
  | { type: typeof actions.STOP_PROCESSING }
  | { type: typeof actions.SET_RESULT; payload: string }
  | { type: typeof actions.START_DELETING }
  | { type: typeof actions.STOP_DELETING };

const initialState: UserState = {
  users: [],
  selectedUser: null,
  processing: false,
  error: "",
  deleteError: "",
  result: "",
  isDeleting: false,
};

const userReducer = (state: UserState, action: UserAction) => {
  switch (action.type) {
    case actions.SET_USERS:
      return {
        ...state,
        users: action.payload,
      };

    case actions.SET_SELECTED_USER:
      return {
        ...state,
        selectedUser: action.payload,
      };

    case actions.CLEAR_SELECTED_USER:
      return {
        ...state,
        selectedUser: null,
      };

    case actions.SET_ERROR:
      return {
        ...state,
        error: action.payload,
        users: [],
      };

    case actions.CLEAR_ERROR:
      return {
        ...state,
        error: "",
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

    case actions.START_DELETING:
      return { ...state, isDeleting: true, deleteError: "" };

    case actions.STOP_DELETING:
      return { ...state, isDeleting: false };

    case actions.SET_DELETE_ERROR:
      return { ...state, deleteError: action.payload };

    default:
      return state;
  }
};

const useUserReducer = () => React.useReducer(userReducer, initialState);

export default useUserReducer;
