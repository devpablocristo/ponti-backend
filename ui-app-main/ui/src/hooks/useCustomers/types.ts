export type CustomerData = {
  id: number;
  name: string;
};

export type CustomerPayload = {
  data: CustomerData[];
  total: number;
};
