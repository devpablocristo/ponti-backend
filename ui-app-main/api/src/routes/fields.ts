import { Request, Response, Router } from "express";
import { ApiClient, ApiResponse } from "../clients/ApiClient";
import { configService } from "../configService";
import { cache } from "./index";

const apiClient = new ApiClient(configService.baseManagerApi);
const router: Router = Router();

router.get("", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const project_id = parseInt(req.query.project_id as string) || 0;
    if (project_id === 0) {
      res.status(400).json({ message: "Proyecto no encontrado" });
      return;
    }

    const url = `projects/fields/${project_id}`;

    const cachedFields = cache.get(url);
    if (cachedFields) {
      res.status(200).json(cachedFields);
      return;
    }

    const { data: fields } = await apiClient.get<any>(
      `/projects/${project_id}/fields`,
      headers
    );

    const data = {
      success: true,
      data: {
        data: fields,
        total: fields.length,
      },
    };

    if (fields.length > 0) {
      cache.set(url, data);
    }

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

router.delete("/:id", async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const data = await apiClient.delete<any>(`/${id}`, headers);
    setImmediate(() => cache.flushAll());
    res.status(200).json(data);
  } catch (error: any) {
    console.log(error);
    const err = error as ApiResponse<null>;

    if ("error" in err) {
      res.status(err.error?.status || 500).json(err);
      return;
    }

    res.status(500).json({
      success: false,
      message: "Error inesperado",
      error: { status: 500, details: "No se pudo obtener el proyecto" },
    });
  }
});

export default router;
