export type FormOptions = {
  clients: Entity[];
  managers: Entity[];
  campaigns: Entity[];
  investors: Investor[];
  crops: Entity[];
  rentTypes: Entity[];
};

export type Entity = {
  id: number;
  name: string;
};

export type Investor = {
  id: number;
  name: string;
  percentage: number;
};
