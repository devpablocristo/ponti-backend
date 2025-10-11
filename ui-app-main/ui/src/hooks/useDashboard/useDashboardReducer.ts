import React from "react";

import * as actions from "./actions";
import { DashboardData } from "./types";

interface DashboardState {
  dashboard: DashboardData | null;
}

const initialState: DashboardState = {
  dashboard: null,
};

type Action =
  | { type: typeof actions.SET_DASHBOARD; payload: DashboardData | null }

const dashboardReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_DASHBOARD:
      return {
        ...state,
        dashboard: action.payload,
      };
    default:
      return state;
  }
};

const useDashboardReducer = () => React.useReducer(dashboardReducer, initialState);

export default useDashboardReducer;
