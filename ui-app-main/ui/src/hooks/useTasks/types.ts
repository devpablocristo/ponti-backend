export interface TaskData {
  workorder_id: number;
  workorder_number: string;
  date: string;
  field_name: string;
  crop_name: string;
  contractor: string;
  labor_name: string;
  category_name: string;
  surface_ha: string;
  cost_ha: number;
  investor_name: string;

  usd_avg_value: number;
  net_total: number;
  total_iva: number;
  usd_cost_ha: number;
  usd_net_total: number;

  invoice_id: number;
  invoice_number: string;
  invoice_company: string;
  invoice_date: string;
  invoice_status: string;
}

export interface InvoiceData {
  workorder_id: number;
  invoice_id: number;
  invoice_number: string;
  invoice_company: string;
  invoice_date: string;
  invoice_status: string;
}

export interface TaskToSave {
  name: string;
  category_id: number;
  price: string;
  contractor_name: string;
}

export interface TaskInfo {
  id: number;
  name: string;
  category_id: number;
  price: string;
  contractor_name: string;
  category_name: string;
}

export type Metrics = {
  surface_ha: number;
  net_total_cost: number;
  avg_cost_per_ha: number;
};
