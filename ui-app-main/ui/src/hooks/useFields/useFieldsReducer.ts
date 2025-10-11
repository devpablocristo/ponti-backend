import React from "react";

import * as actions from "./actions";
import { Data } from "./types";

interface ProductState {
  fields: Data[];
  total: number;
  processing: boolean;
  error: string;
}

const initialState: ProductState = {
  fields: [],
  total: 0,
  processing: false,
  error: "",
};

type Action =
  | { type: typeof actions.SET_FIELDS; payload: Data[] }
  | { type: typeof actions.SET_TOTAL; payload: number }
  | { type: typeof actions.SET_ERROR; payload: string }
  | { type: typeof actions.START_PROCESSING }
  | { type: typeof actions.STOP_PROCESSING };

const fieldsReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_FIELDS:
      return {
        ...state,
        fields: action.payload,
      };

    case actions.SET_TOTAL:
      return {
        ...state,
        total: action.payload,
      };

    case actions.SET_ERROR:
      return {
        ...state,
        error: action.payload,
        projects: [],
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
    default:
      return state;
  }
};

const useFieldsReducer = () => React.useReducer(fieldsReducer, initialState);

export default useFieldsReducer;
