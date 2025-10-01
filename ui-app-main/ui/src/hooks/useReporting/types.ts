export interface FieldCropReportData {
  project_id: number;
  project_name: string;
  customer_id: number;
  customer_name: string;
  campaign_id: number;
  campaign_name: string;
  columns: {
    id: string;
    field_id: number;
    field_name: string;
    crop_id: number;
    crop_name: string;
  }[];
  rows: {
    key: string;
    unit: string;
    value_type: string;
    values: {
      [columnId: string]: {
        number: string;
      };
    };
  }[];
}

export interface InvestorContributionReportData {
  project_id: number;
  project_name: string;
  customer_id: number;
  customer_name: string;
  campaign_id: number;
  campaign_name: string;
  investor_headers: {
    investor_id: number;
    investor_name: string;
    share_pct: number;
  }[];
  general: {
    surface_total_ha: number;
    lease_fixed_usd: number;
    lease_is_fixed: boolean;
    admin_per_ha_usd: number;
    admin_total_usd: number;
  };
  contributions: {
    key: string;
    sort_index: number;
    type: string;
    label: string;
    total_usd: number;
    total_usd_ha: number;
    requires_manual_attribution: boolean;
    attribution_note?: string;
    investors: {
      investor_id: number;
      investor_name: string;
      amount_usd: number;
      share_pct: number;
    }[];
  }[];
  pre_harvest: {
    total_usd: number;
    total_us_ha: number;
    investors: {
      investor_id: number;
      investor_name: string;
      amount_usd: number;
      share_pct: number;
    }[];
  };
  comparison: {
    investor_id: number;
    investor_name: string;
    agreed_share_pct: number;
    agreed_usd: number;
    actual_usd: number;
    adjustment_usd: number;
  }[];
  harvest: {
    rows: {
      key: string;
      type: string;
      total_usd: number;
      total_us_ha: number;
      investors: {
        investor_id: number;
        investor_name: string;
        amount_usd: number;
        share_pct: number;
      }[];
    }[];
    footer_payment_agreed: {
      investor_id: number;
      investor_name: string;
      amount_usd: number;
      share_pct: number;
    }[];
    footer_payment_adjustment: {
      investor_id: number;
      investor_name: string;
      amount_usd: number;
    }[];
  };
}

export interface SummaryResultsReportData {
  project_id: number;
  project_name: string;
  customer_id: number;
  customer_name: string;
  campaign_id: number;
  campaign_name: string;
  crops: {
    crop_id: number;
    crop_name: string;
    surface_ha: string;
    net_income_usd: string;
    direct_costs_usd: string;
    rent_usd: string;
    structure_usd: string;
    total_invested_usd: string;
    operating_result_usd: string;
    crop_return_pct: string;
  }[];
  totals: {
    total_surface_ha: string;
    total_net_income_usd: string;
    total_direct_costs_usd: string;
    total_rent_usd: string;
    total_structure_usd: string;
    total_invested_project_usd: string;
    total_operating_result_usd: string;
    project_return_pct: string;
  };
  general_crops: {
    total_surface_ha: string;
    total_net_income_usd: string;
    total_direct_costs_usd: string;
    total_rent_usd: string;
    total_structure_usd: string;
    total_invested_project_usd: string;
    total_operating_result_usd: string;
    project_return_pct: string;
  };
}

export interface RowToRender {
  label: string;
  key: string;
  valueFormat: {
    [key: string]: (...args: number[]) => string;
  };
  classNameHeader?: string;
  classNameRows?: string;
  showIndicator?: boolean;
}