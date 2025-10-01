import React from "react";
import { CommerceInfoData } from "./types";
import * as actions from "./actions";

interface CommerceState {
  result: string;
  commerceInfoList: CommerceInfoData[];
}

const initialState: CommerceState = {
  result: "",
  commerceInfoList: [],
};

type Action =
  | { type: typeof actions.SET_RESULT; payload: string }
  | {
      type: typeof actions.SET_COMMERCE_INFO_LIST;
      payload: CommerceInfoData[];
    };

const commerceReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_RESULT:
      return {
        ...state,
        result: action.payload,
      };
    case actions.SET_COMMERCE_INFO_LIST:
      return {
        ...state,
        commerceInfoList: action.payload,
      };
    default:
      return state;
  }
};

const useCommerceReducer = () =>
  React.useReducer(commerceReducer, initialState);

export default useCommerceReducer;
