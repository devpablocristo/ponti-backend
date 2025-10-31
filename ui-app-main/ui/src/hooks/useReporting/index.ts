import React, { useState } from "react";
import { AxiosError } from "axios";

import useReportingReducer from "./useReportingReducer.ts";
import * as actions from "./actions";
import { FieldCropReportData, InvestorContributionReportData, SummaryResultsReportData } from "./types";
import { SuccessResponse, ErrorResponse } from "../../restclient/types";
import APIClient from "../../restclient/apiInstance";

const request = new APIClient({
  timeout: 15000,
  baseURL: "/api",
});

const useReporting = () => {
  const [{
    fieldCropReportingData,
    investorContributionReportingData,
    summaryResultsReportingData
  }, dispatch] = useReportingReducer();
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getFieldCropReportingData = React.useCallback(async (queryString: string) => {
    setProcessing(true);
    setError(null);
    let queryParams = "";
    if (queryString !== "") {
      queryParams = `?${ queryString }`;
    }

    try {
      const response = await request.get<SuccessResponse<FieldCropReportData>>(
        `reports/field-crop` + queryParams
      );

      if (response.success) {
        dispatch({
          type: actions.SET_FIELD_CROP_REPORTING,
          payload: response.data,
        });
        return;
      }

      setError("Ocurrió un error en la búsqueda del reporte");
    } catch (error) {
      dispatch({
        type: actions.SET_FIELD_CROP_REPORTING,
        payload: null,
      });
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const message =
            errorResponse.error.details ||
            "Error desconocido en la búsqueda del reporte.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, [dispatch]);

  const getInvestorContributionReportingData = React.useCallback(async (queryString: string) => {
    setProcessing(true);
    setError(null);
    let queryParams = "";
    if (queryString !== "") {
      queryParams = `?${ queryString }`;
    }

    try {
      const response = await request.get<SuccessResponse<InvestorContributionReportData>>(
        `reports/investor-contribution` + queryParams
      );

      if (response.success) {
        dispatch({
          type: actions.SET_INVESTOR_CONTRIBUTION_REPORTING,
          payload: response.data,
        });
        return;
      }

      setError("Ocurrió un error en la búsqueda del reporte");
    } catch (error) {
      dispatch({
        type: actions.SET_INVESTOR_CONTRIBUTION_REPORTING,
        payload: null,
      });
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const message =
            errorResponse.error.details ||
            "Error desconocido en la búsqueda del reporte.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, [dispatch]);

  const getSummaryResultsReportingData = React.useCallback(async (queryString: string) => {
    setProcessing(true);
    setError(null);
    let queryParams = "";
    if (queryString !== "") {
      queryParams = `?${ queryString }`;
    }

    try {
      const response = await request.get<SuccessResponse<SummaryResultsReportData>>(
        `reports/summary-results` + queryParams
      );

      if (response.success) {
        dispatch({
          type: actions.SET_SUMMARY_RESULTS_REPORTING,
          payload: response.data,
        });
        return;
      }

      setError("Ocurrió un error en la búsqueda del reporte");
    } catch (error) {
      dispatch({
        type: actions.SET_SUMMARY_RESULTS_REPORTING,
        payload: null,
      });
      const axiosError = error as AxiosError;

      if (axiosError.response) {
        const errorResponse = axiosError.response.data as ErrorResponse;

        if (errorResponse.error) {
          const message =
            errorResponse.error.details ||
            "Error desconocido en la búsqueda del reporte.";

          setError(message);
          return;
        }
      }

      setError("Error en el servicio, inténtalo más tarde.");
    } finally {
      setProcessing(false);
    }
  }, [dispatch]);

  return {
    fieldCropReportingData,
    investorContributionReportingData,
    summaryResultsReportingData,
    processing,
    error,
    getFieldCropReportingData,
    getInvestorContributionReportingData,
    getSummaryResultsReportingData,
  };
};

export default useReporting;
