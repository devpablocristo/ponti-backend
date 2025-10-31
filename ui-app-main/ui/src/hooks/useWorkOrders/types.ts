export interface OrderInputDetail {
  input: string;
  consumption: string;
  category: string;
  dose: number;
  cost: number;
  unitPrice: number;
  total: number;
}

export interface OrdersData {
  id: number;
  number: string;
  project_name: string;
  field_name: string;
  lot_name: string;
  date: string;
  crop_name: string;
  labor_name: string;
  labor_category_name: string;
  type_name: string;
  contractor: string;
  surface_ha: string;
  supply_name: string;
  consumption: string;
  category_name: string;
  dose: number;
  cost_per_ha: number;
  unit_price: number;
  total_cost: number;
}

export interface WorkorderData {
  id: number;
  number: string;
  project_id: number;
  field_id: number;
  lot_id: number;
  crop_id: number;
  labor_id: number;
  contractor: string;
  observations: string;
  date: string;
  investor_id: number;
  effective_area: number;
  items: WorkorderItem[];
}

export type WorkorderItem = {
  supply_id: number;
  total_used: number;
  final_dose: number;
};

export type Workorder = {
  number: string;
  project_id: number;
  field_id: number;
  lot_id: number;
  crop_id: number;
  labor_id: number;
  contractor: string;
  observations: string;
  date: string;
  investor_id: number;
  effective_area: number;
  items: WorkorderItem[];
};

export type Metrics = {
  surface_ha: number;
  liters: number;
  kilograms: number;
  direct_cost: number;
};
