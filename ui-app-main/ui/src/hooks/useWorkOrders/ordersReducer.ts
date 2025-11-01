import React from "react";

import * as actions from "./actions";
import { OrdersData, WorkorderData, Metrics } from "./types";
import { PageInfo } from "../useDatabase/projects/types";

interface OrdersState {
  orders: OrdersData[];
  metrics: Metrics;
  selectedOrder: WorkorderData | null;
  processing: boolean;
  error: string;
  resultCreation: string;
  pageInfo: PageInfo | null;
}

const initialState: OrdersState = {
  orders: [],
  metrics: {
    surface_ha: 0,
    liters: 0,
    kilograms: 0,
    direct_cost: 0,
  },
  selectedOrder: null,
  processing: false,
  error: "",
  resultCreation: "",
  pageInfo: null,
};

type Action =
  | { type: typeof actions.SET_ORDERS; payload: OrdersData[] }
  | { type: typeof actions.SET_RESULT_CREATION; payload: string }
  | { type: typeof actions.SET_PAGE_INFO; payload: PageInfo }
  | { type: typeof actions.SET_SELECTED_ORDER; payload: WorkorderData | null }
  | { type: typeof actions.SET_METRICS; payload: Metrics };

const ordersReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_ORDERS:
      return {
        ...state,
        orders: action.payload,
      };
    case actions.SET_RESULT_CREATION:
      return {
        ...state,
        resultCreation: action.payload,
      };
    case actions.SET_PAGE_INFO:
      return {
        ...state,
        pageInfo: action.payload,
      };
    case actions.SET_SELECTED_ORDER:
      return {
        ...state,
        selectedOrder: action.payload,
      };
    case actions.SET_METRICS:
      return {
        ...state,
        metrics: action.payload,
      };
    default:
      return state;
  }
};

const useOrdersReducer = () => React.useReducer(ordersReducer, initialState);

export default useOrdersReducer;
