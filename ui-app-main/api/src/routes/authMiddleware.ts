import { Request, Response, NextFunction } from "express";
//import redis from "../clients/redisClient";

import { ApiClient } from "../clients/ApiClient";
import { configService } from "../configService";

const apiClient = new ApiClient(configService.baseLoginApi);

const AUTH_API_URL = "/auth/validate-token";
const REFRESH_API_URL = "/auth/access-token";

declare global {
  namespace Express {
    interface Request {
      user?: UserData;
    }
  }
}

export interface UserData {
  status: string;
  userID: string;
  rolID: string;
  hash: string;
  exp: number;
}

export const verifyToken = async (
  req: Request,
  res: Response,
  next: NextFunction
) => {
  try {
    const authHeader = req.headers.authorization;
    const token = authHeader?.split(" ")[1];
    if (!token) {
      res.status(401).json({ message: "No autorizado" });
      return;
    }

    // const redisKey = `session:${token}`;
    // const cachedUser = await redis.get(redisKey);

    // if (cachedUser) {
    //   //req.user = JSON.parse(cachedUser);
    //   return next();
    // }

    try {
      const headers = { Authorization: `Bearer ${token}` };
      const { data, success } = await apiClient.get<UserData>(
        AUTH_API_URL,
        headers
      );

      if (success) {
        req.user = data;
        next();
        return;
      }
    } catch (error) {
      console.error("Error validando el token:", error);
    }

    res.status(401).json({ message: "Sesión inválida" });
  } catch (error) {
    console.error("Error en autenticación:", error);
    res.status(500).json({ message: "Error en autenticación" });
  }
};
