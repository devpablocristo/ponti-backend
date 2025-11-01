import React from "react";

import * as actions from "./actions";
import { FieldCropReportData, InvestorContributionReportData, SummaryResultsReportData } from "./types";

interface ReportingState {
  fieldCropReportingData: FieldCropReportData | null;
  investorContributionReportingData: InvestorContributionReportData | null;
  summaryResultsReportingData: SummaryResultsReportData | null;
}

const initialState: ReportingState = {
  fieldCropReportingData: null,
  investorContributionReportingData: null,
  summaryResultsReportingData: null,
};

type Action =
  | { type: typeof actions.SET_FIELD_CROP_REPORTING; payload: FieldCropReportData | null }
  | { type: typeof actions.SET_INVESTOR_CONTRIBUTION_REPORTING; payload: InvestorContributionReportData | null }
  | { type: typeof actions.SET_SUMMARY_RESULTS_REPORTING; payload: SummaryResultsReportData | null };

const reportingReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_FIELD_CROP_REPORTING:
      return {
        ...state,
        fieldCropReportingData: action.payload,
      };
    case actions.SET_INVESTOR_CONTRIBUTION_REPORTING:
      return {
        ...state,
        investorContributionReportingData: action.payload,
      };
    case actions.SET_SUMMARY_RESULTS_REPORTING:
      return {
        ...state,
        summaryResultsReportingData: action.payload,
      }
    default:
      return state;
  }
};

const useReportingReducer = () => React.useReducer(reportingReducer, initialState);

export default useReportingReducer;
