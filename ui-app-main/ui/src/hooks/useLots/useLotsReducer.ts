import React from "react";

import * as actions from "./actions";
import { Crop, LotsData, LotKPIs } from "./types";
import { PageInfo } from "../useDatabase/projects/types";

interface ProductState {
  lots: LotsData[];
  kpis: LotKPIs;
  pageInfo: PageInfo | null;
  crops: Crop[];
  processing: boolean;
  isLastPage: boolean;
  error: string;
  result: string;
}

const initialState: ProductState = {
  lots: [],
  kpis: {
    seeded_area: 0,
    harvested_area: 0,
    yield_tn_per_ha: 0,
    cost_per_hectare: 0,
    superficie_total: 0,
  },
  pageInfo: null,
  crops: [],
  processing: false,
  error: "",
  isLastPage: false,
  result: "",
};

type Action =
  | { type: typeof actions.SET_LOTS; payload: LotsData[] }
  | { type: typeof actions.SET_PAGE_INFO; payload: PageInfo }
  | { type: typeof actions.SET_CROPS; payload: Crop[] }
  | { type: typeof actions.SET_RESULT; payload: string }
  | { type: typeof actions.SET_KPIS; payload: LotKPIs };

const lotsReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_PAGE_INFO:
      return {
        ...state,
        pageInfo: action.payload,
      };
    case actions.SET_LOTS:
      return {
        ...state,
        lots: action.payload,
      };
    case actions.SET_CROPS:
      return {
        ...state,
        crops: action.payload,
      };
    case actions.SET_RESULT:
      return {
        ...state,
        result: action.payload,
      };
    case actions.SET_KPIS:
      return {
        ...state,
        kpis: action.payload,
      };
    default:
      return state;
  }
};

const useLotsReducer = () => React.useReducer(lotsReducer, initialState);

export default useLotsReducer;
