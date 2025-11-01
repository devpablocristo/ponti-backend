import React from "react";

import * as actions from "./actions";
import { SupplyMovement } from "./types";
import { PageInfo } from "../useDatabase/projects/types";
import { Summary } from "./types";

interface SupplyMovementsState {
  supplyMovements: SupplyMovement[];
  selectedSupplyMovement: SupplyMovement | null;
  summary: Summary;
  processing: boolean;
  error: string;
  resultCreation: SupplyMovementResponse;
  pageInfo: PageInfo | null;
}

interface SupplyMovementResult {
  supply_movement_id: number;
  is_saved: boolean;
  error_detail: string;
}

interface SupplyMovementResponse {
  supply_movements: SupplyMovementResult[];
}

const initialState: SupplyMovementsState = {
  supplyMovements: [],
  selectedSupplyMovement: null,
  summary: {
    total_kg: 0,
    total_lt: 0,
    total_usd: 0,
  },
  processing: false,
  error: "",
  resultCreation: {
    supply_movements: [],
  },
  pageInfo: null,
};

type Action =
  | { type: typeof actions.SET_SUPPLY_MOVEMENTS; payload: SupplyMovement[] }
  | {
      type: typeof actions.SET_RESULT_CREATION;
      payload: SupplyMovementResponse;
    }
  | { type: typeof actions.SET_PAGE_INFO; payload: PageInfo }
  | {
      type: typeof actions.SET_SELECTED_SUPPLY_MOVEMENT;
      payload: SupplyMovement | null;
    }
  | { type: typeof actions.SET_SUMMARY; payload: Summary };

const ordersReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_SUPPLY_MOVEMENTS:
      return {
        ...state,
        supplyMovements: action.payload,
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
    case actions.SET_SELECTED_SUPPLY_MOVEMENT:
      return {
        ...state,
        selectedSupplyMovement: action.payload,
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

const useOrdersReducer = () => React.useReducer(ordersReducer, initialState);

export default useOrdersReducer;
