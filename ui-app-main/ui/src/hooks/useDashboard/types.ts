export interface DashboardData {
  metrics: Metrics;
  management_balance: ManagementBalance;
  crop_incidence: CropIncidence;
  operational_indicators: OperationalIndicators;
}

export interface Metrics {
  sowing: SowingMetric;
  harvest: HarvestMetric;
  costs: CostsMetric;
  investor_contributions: InvestorContributions;
  operating_result: OperatingResultMetric;
}

export interface SowingMetric {
  progress_pct: string;
  hectares: string;
  total_hectares: string;
}

export interface HarvestMetric {
  progress_pct: string;
  hectares: string;
  total_hectares: string;
}

export interface CostsMetric {
  progress_pct: string;
  executed_usd: string;
  budget_usd: string;
}

export interface InvestorContributions {
  items: InvestorItem[];
}

export interface InvestorItem {
  investor_id: number;
  investor_name: string;
  share_pct: string;
}

export interface OperatingResultMetric {
  margin_pct: string;
  result_usd: string;
  total_costs_usd: string;
}

export interface ManagementBalance {
  totals: BalanceTotals;
  items: BalanceItem[];
}

export interface BalanceTotals {
  executed_usd: string;
  invested_usd: string;
  stock_usd: string;
}

export interface BalanceItem {
  category: string;
  label: string;
  executed_usd: string;
  invested_usd: string;
  stock_usd?: string;
  order: number;
}

export interface CropIncidence {
  items: CropItem[];
  total: CropTotal;
}

export interface CropItem {
  crop_id: number;
  name: string;
  hectares: string;
  cost_per_ha_usd: string;
  incidence_pct: string;
}

export interface CropTotal {
  hectares: string;
  avg_cost_per_ha_usd: string;
}

export interface OperationalIndicators {
  items: OperationalItem[];
}

export interface OperationalItem {
  type: string;
  title: string;
  date: string;
  workorder_id?: number;
}