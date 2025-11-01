export interface StockData {
  name: string;
  category: string;
  quantityEntered: string;
  quantityConsumed: string;
  totalStock: string;
  realStock: string;
  difference: string;
  pricePerUnit: string;
  totalNet: string;
}

export interface Summary {
  total_kg: number;
  total_lt: number;
  total_usd: number;
}

export interface GetStocksResponse {
  items: GetStockItems[];
  net_total_usd: number;
  total_liters: number;
  total_kilograms: number;
}

export interface GetStockItems {
  id: number;
  supply_name: string;
  investor_name: string;
  stock_units: number;
  real_stock_units: number;
  stock_difference: number;
  total_usd: number;
  class_type: string;
  close_date: string;
  supply_unit_id: number;
  supply_unit_price: number;
  entry_stock: number;
  out_stock: number;
  consumed: number;
}
