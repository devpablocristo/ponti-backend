import { Request, Response, Router } from "express";
import { ApiClient, ApiResponse } from "../clients/ApiClient";
import { configService } from "../configService";
import { cache } from ".";

const apiClient = new ApiClient(configService.baseManagerApi);

const router: Router = Router();

router.get("/export/:id", async (req, res) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const project_id = parseInt(req.params.id as string) || 0;
    if (project_id === 0) {
      res.status(400).json({ message: "Proyecto obligatorio" });
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const response = await apiClient.get<any>(
      `/projects/${project_id}/supply-movements/export`,
      { headers, responseType: "arraybuffer" }
    );

    res.setHeader("Content-Disposition", 'attachment; filename="insumos.xlsx"');
    res.setHeader(
      "Content-Type",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    );

    res.send(response.data);
  } catch (error) {
    console.error(error);
    res.status(500).send("Error al exportar insumos");
  }
});

router.get("/:project_id", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const project_id = parseInt(req.params.project_id as string) || 0;
    if (project_id === 0) {
      res.status(400).json({ message: "Proyecto obligatorio" });
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const cachedMovements = cache.get(`movements/${project_id}`);
    if (cachedMovements) {
      res.status(200).json(cachedMovements);
      return;
    }

    const { data: movements } = await apiClient.get<any>(
      `/projects/${project_id}/supply-movements`,
      headers
    );

    const entries = movements.entries ?? [];
    const data = {
      success: true,
      data: {
        summary: movements.summary,
        entries,
        page_info: {
          total: entries.length,
          page: 1,
          per_page: 100,
          max_page: 1,
        },
      },
    };

    if (entries.length > 0) {
      cache.set(`movements/${project_id}`, data);
    }

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
      error: { status: 500, details: "No se pudo procesar la solicitud" },
    });
  }
});

router.delete(
  "/:id/project/:project_id",
  async (req: Request, res: Response) => {
    try {
      const userId = req.user?.userID;
      if (!userId) {
        res.status(401).json({ message: "Usuario no autenticado" });
        return;
      }

      const id = parseInt(req.params.id as string) || 0;
      if (id === 0) {
        res.status(400).json({ message: "Id obligatorio" });
        return;
      }

      const project_id = parseInt(req.params.project_id as string) || 0;
      if (project_id === 0) {
        res.status(400).json({ message: "Proyecto obligatorio" });
        return;
      }

      const headers = {
        "X-API-KEY": configService.apiKey,
        "X-User-Id": userId,
      };

      await apiClient.delete<any>(
        `/projects/${project_id}/supply-movements/${id}`,
        headers
      );

      const data = {
        success: true,
      };

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
        error: { status: 500, details: "No se pudo procesar la solicitud" },
      });
    }
  }
);

router.post("/:project_id", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const project_id = parseInt(req.params.project_id as string) || 0;
    if (project_id === 0) {
      res.status(400).json({ message: "Proyecto obligatorio" });
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: result } = await apiClient.post<any>(
      `/projects/${project_id}/supply-movements`,
      req.body,
      headers
    );

    const data = {
      success: true,
      data: result,
    };

    setImmediate(() => cache.flushAll());

    res.status(201).json(data);
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
      error: { status: 500, details: "No se pudo procesar la solicitud" },
    });
  }
});

export default router;
