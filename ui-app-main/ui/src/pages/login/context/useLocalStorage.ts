import { TokenResponse } from "../types";

export const getAccessToken = (): string | null => {
  return localStorage.getItem("access_token");
};

export const setAccessToken = (token: string) => {
  localStorage.setItem("access_token", token);
};

export const getRefreshToken = (): string | null => {
  return localStorage.getItem("refresh_token");
};

export const clearLocalStorage = () => {
  localStorage.clear();
};

export const setLocalStorage = (token: TokenResponse) => {
  localStorage.setItem("access_token", token.access_token);
  localStorage.setItem("refresh_token", token.refresh_token);
};
