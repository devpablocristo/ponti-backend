export type UserData = {
  username: string;
  password: string;
};

export type LogoutData = {
  refresh_token: string;
};

export interface DecodedToken {
  ID: number;
  Rol: number;
  Username: string;
  exp: number;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
}
