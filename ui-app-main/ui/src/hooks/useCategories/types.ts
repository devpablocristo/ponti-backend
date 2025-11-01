export type CategoryData = {
  id: number;
  name: string;
  type_id: number;
};

export type TypeData = {
  id: number;
  name: string;
};

export type PayloadCategories = {
  categories: CategoryData[];
};

export type PayloadTypes = {
  types: TypeData[];
};
