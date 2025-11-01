export interface ProductData {
  ingreso: string;
  remito: string;
  fecha: string;
  inversor: string;
  insumo: string;
  cantidad: string;
  rubro: string;
  tipoClase: string;
  proveedor: string;
  precioUnidad: number;
  totalNeto: number;
}

export type Product = {
  name: string;
  unit: number;
  price: number;
  type: number;
  category: number;
};

export type Supply = {
  id: number;
  name: string;
  price: string;
  unit_id?: number;
  type_name: string;
  category_name: string;
  category_id?: number;
  type_id?: number;
};

export type SupplyResponse = {
  data: Supply[];
};
