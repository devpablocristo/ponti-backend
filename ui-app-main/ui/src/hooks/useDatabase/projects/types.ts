export type Project = {
  name: string;
  customer: Data;
  campaign: Data;
  managers: Data[];
  investors: {
    id: number | null;
    name: string;
    percentage: number;
  }[];
  admin_cost_investors: {
    id: number | null;
    name: string;
    percentage: number;
  }[];
  admin_cost: number;
  fields: Field[];
  updated_at: string | undefined;
};

type Data = {
  id: number | null;
  name: string;
};

export type Field = {
  id: number;
  name: string;
  lease_type_name?: string;
  lease_type_id: number;
  lease_type_percent: number;
  lease_type_value: number;
  lots: Plot[];
};

export type Plot = {
  id: number;
  name: string;
  hectares: number;
  previous_crop_id: number;
  current_crop_id: number;
  current_crop_name?: string;
  previous_crop_name?: string;
  season: string;
};

export type PageInfo = {
  per_page: number;
  page: number;
  max_page: number;
  total: number;
};

export type ProjectPayload = {
  data: ProjectData[];
  page_info: PageInfo;
  total_hectares: number;
};

export type ProjectDropdownPayload = {
  data: ProjectDropdown[];
  page_info: PageInfo;
};

export type ProjectData = {
  id: number;
  name: string;
  customer: string;
  campaign: string;
  managers: string;
  investors: string;
  fields: FieldData[];
};

export type ProjectDropdown = {
  id: number;
  name: string;
  customerID: number;
  customerName: string;
};

type FieldData = {
  name: string;
  lease_type: string;
  hectares: string;
  crops: string;
};
