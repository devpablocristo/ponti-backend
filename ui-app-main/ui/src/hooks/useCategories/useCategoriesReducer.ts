import React from "react";

import * as actions from "./actions";
import { CategoryData, TypeData } from "./types";

interface ProductState {
  categories: CategoryData[];
  types: TypeData[];
  processing: boolean;
  error: string;
}

const initialState: ProductState = {
  categories: [],
  types: [],
  processing: false,
  error: "",
};

type Action =
  | { type: typeof actions.SET_CATEGORIES; payload: CategoryData[] }
  | { type: typeof actions.SET_TYPES; payload: TypeData[] };

const categoriesReducer = (state: typeof initialState, action: Action) => {
  switch (action.type) {
    case actions.SET_CATEGORIES:
      return {
        ...state,
        categories: action.payload,
      };

    case actions.SET_TYPES:
      return {
        ...state,
        types: action.payload,
      };

    default:
      return state;
  }
};

const useCategoriesReducer = () =>
  React.useReducer(categoriesReducer, initialState);

export default useCategoriesReducer;
