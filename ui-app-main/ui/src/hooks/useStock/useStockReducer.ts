import React from "react";

import * as actions from "./actions";
import { GetStockItems, Summary } from "./types";

interface StockState {
  stock: GetStockItems[];
  processing: boolean;
  isLastPage: boolean;
  error: string;
  currentPage: number;
  summary: Summary;
}

const initialState: StockState = {
  stock: [],
  currentPage: 1,
  processing: false,
  error: "",
  isLastPage: false,
  summary: {
    total_kg: 0,
    total_lt: 0,
    total_usd: 0,
  },
};

type Action =
  | { type: typeof actions.SET_STOCK; payload: GetStockItems[] }
  | { type: typeof actions.SET_SUMMARY; payload: Summary }
  | { type: typeof actions.SET_PAGE; payload: number };

const stockReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_PAGE:
      return {
        ...state,
        currentPage: action.payload || 1,
      };
    case actions.SET_STOCK:
      return {
        ...state,
        stock: action.payload,
      };
    case actions.SET_SUMMARY:
      return {
        ...state,
        summary: action.payload,
      };
    default:
      return state;
  }
};

const useStockReducer = () => React.useReducer(stockReducer, initialState);

export default useStockReducer;
