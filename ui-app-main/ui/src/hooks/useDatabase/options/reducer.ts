import React from "react";

import * as actions from "./actions";
import { FormOptions } from "./types";

type State = {
  options: FormOptions | null;
};

const initialState: State = {
  options: null,
};

type Action = { type: typeof actions.SET_OPTIONS; payload: FormOptions };

const optionsReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_OPTIONS:
      return {
        ...state,
        options: action.payload,
      };
    default:
      return state;
  }
};

const useOptionsReducer = () => React.useReducer(optionsReducer, initialState);

export default useOptionsReducer;
