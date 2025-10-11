export type UserData = {
  id: number;
  username: string;
  rol: string;
  created_at: string;
  updated_at: string;
  created_by: string;
};

export type UserNew = {
  username: string;
  password: string;
  passwordConfirm: string;
};
