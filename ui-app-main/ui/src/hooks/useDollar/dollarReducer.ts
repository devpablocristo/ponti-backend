import React from "react";

import * as actions from "./actions";
import { DollarData } from "./types";

interface DollarState {
  dollars: DollarData[];
  result: string;
}

const initialState: DollarState = {
  dollars: [],
  result: "",
};

type Action =
  | { type: typeof actions.SET_DOLLARS; payload: DollarData[] }
  | { type: typeof actions.SET_RESULT; payload: string };

const dollarReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_DOLLARS:
      return {
        ...state,
        dollars: action.payload,
      };
    case actions.SET_RESULT:
      return {
        ...state,
        result: action.payload,
      };
    default:
      return state;
  }
};

const useDollarReducer = () => React.useReducer(dollarReducer, initialState);

export default useDollarReducer;
