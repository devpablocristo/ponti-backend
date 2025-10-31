import { Request, Response, Router } from "express";
import { cache } from "./index";

import { configService } from "../configService";
import { ApiClient, ApiResponse } from "../clients/ApiClient";

const apiClient = new ApiClient(configService.baseLoginApi);

const router: Router = Router();

router.post("", async (req: Request, res: Response) => {
  try {
    const authHeader = req.headers.authorization;
    const headers = authHeader ? { Authorization: authHeader } : {};

    const requestData = {
      ...req.body,
      password_confirm: req.body.passwordConfirm,
    };

    delete requestData.passwordConfirm;

    const data = await apiClient.post<any>(`auth/users`, requestData, headers);
    cache.del("users");
    res.status(200).json(data);
  } catch (error) {
    const err = error as ApiResponse<null>;

    if ("error" in err) {
      res.status(err.error?.status || 500).json(err);
      return;
    }

    res.status(500).json({
      success: false,
      message: "Error inesperado",
      error: { status: 500, details: "No se pudo procesar la solicitud" },
    });
  }
});

router.get("", async (req: Request, res: Response) => {
  const cachedUsers = cache.get("users");
  if (cachedUsers) {
    res.status(200).json(cachedUsers);
    return;
  }

  try {
    const authHeader = req.headers.authorization;
    const headers = authHeader ? { Authorization: authHeader } : {};

    const data = await apiClient.get<any>("/auth/users", headers);
    if (data && data.data === null) {
      data.data = [];
    }

    cache.set("users", data);
    res.status(200).json(data);
  } catch (error: any) {
    const err = error as ApiResponse<null>;

    if ("error" in err) {
      res.status(err.error?.status || 500).json(err);
      return;
    }

    res.status(500).json({
      success: false,
      message: "Error inesperado",
      error: { status: 500, details: "No se pudo procesar la solicitud" },
    });
  }
});

router.get("/:id", async (req: Request, res: Response) => {
  try {
    const { id } = req.params;

    const cachedUser = cache.get(`users_${id}`);
    if (cachedUser) {
      res.status(200).json(cachedUser);
      return;
    }

    const authHeader = req.headers.authorization;
    const headers = authHeader ? { Authorization: authHeader } : {};

    const data = await apiClient.get<any>(`/auth/users/${id}`, headers);
    cache.set(`users_${id}`, data);
    res.status(200).json(data);
  } catch (error: any) {
    const err = error as ApiResponse<null>;

    if ("error" in err) {
      res.status(err.error?.status || 500).json(err);
      return;
    }

    res.status(500).json({
      success: false,
      message: "Error inesperado",
      error: { status: 500, details: "No se pudo obtener el usuario" },
    });
  }
});

router.delete("/:id", async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const authHeader = req.headers.authorization;
    const headers = authHeader ? { Authorization: authHeader } : {};

    const data = await apiClient.delete<any>(`/auth/users/${id}`, headers);
    cache.del("users");
    cache.del(`users_${id}`);
    res.status(200).json(data);
  } catch (error: any) {
    const err = error as ApiResponse<null>;

    if ("error" in err) {
      res.status(err.error?.status || 500).json(err);
      return;
    }

    res.status(500).json({
      success: false,
      message: "Error inesperado",
      error: { status: 500, details: "No se pudo obtener el usuario" },
    });
  }
});

router.put("/:id", async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const authHeader = req.headers.authorization;
    const headers = authHeader ? { Authorization: authHeader } : {};

    const requestData = {
      ...req.body,
      password_confirm: req.body.passwordConfirm,
    };

    delete requestData.passwordConfirm;

    const data = await apiClient.put<any>(
      `/auth/users/${id}`,
      requestData,
      headers
    );

    cache.del("users");
    cache.del(`users_${id}`);

    res.status(200).json(data);
  } catch (error: any) {
    const err = error as ApiResponse<null>;

    if ("error" in err) {
      res.status(err.error?.status || 500).json(err);
      return;
    }

    res.status(500).json({
      success: false,
      message: "Error inesperado",
      error: { status: 500, details: "No se pudo actualizar el usuario" },
    });
  }
});

export default router;
