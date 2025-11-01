export type UserData = {
  id: number;
  username: string;
  createdAt: string;
  updatedAt: string;
  createdBy: string;
};

export type UserNew = {
  username: string;
  password: string;
  passwordConfirm: string;
};
