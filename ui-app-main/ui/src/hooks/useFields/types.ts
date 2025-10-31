export type Data = {
  id: number;
  name: string;
  project_id: number;
};

export type Payload = {
  data: Data[];
  total: number;
};
