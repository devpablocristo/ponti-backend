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
      `/stocks/export/all`,
      { headers, responseType: "arraybuffer" }
    );

    res.setHeader("Content-Disposition", 'attachment; filename="stock.xlsx"');
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

router.get("/periods/:id", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const projectId = parseInt(req.params.id) || 0;
    if (projectId === 0) {
      res.status(400).json({ message: "Proyecto obligatorio" });
      return;
    }

    const cachedStock = cache.get(`stock:periods:${projectId}`);
    if (cachedStock) {
      res.status(200).json(cachedStock);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: periods } = await apiClient.get<any>(
      `/projects/${projectId}/stocks/periods`,
      headers
    );

    const data = {
      success: true,
      data: periods,
    };

    if (periods.length > 0) {
      setImmediate(() => cache.set(`stock:periods:${projectId}`, data));
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

router.get("/:id", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const projectId = parseInt(req.params.id) || 0;
    if (projectId === 0) {
      res.status(400).json({ message: "Proyecto obligatorio" });
      return;
    }

    let queryString = "";
    const cutOffDate = req.query.cutoff_date as string;
    if (cutOffDate && cutOffDate !== "") {
      queryString = `?cutoff_date=${cutOffDate}`;
    }

    const cachedStock = cache.get(`stock:${projectId}:${queryString}`);
    if (cachedStock) {
      res.status(200).json(cachedStock);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: stock } = await apiClient.get<any>(
      `/projects/${projectId}/stocks/summary${queryString}`,
      headers
    );

    const data = {
      success: true,
      data: stock,
    };

    if (stock.length > 0) {
      setImmediate(() => cache.set(`stock:${projectId}:${queryString}`, data));
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

router.put("/close/:id", async (req: Request, res: Response) => {
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

    const date = new Date(req.body.close_date);
    const month = date.getMonth() + 1;
    const year = date.getFullYear();

    const requestData = {
      close_date: date.toISOString(),
    };

    await apiClient.put<any>(
      `/projects/${req.params.id}/stocks/close-date?month_period=${month}&year_period=${year}`,
      requestData,
      headers
    );

    setImmediate(() => cache.flushAll());

    const data = {
      success: true,
      message: "Stock actualizado exitosamente",
    };

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

router.put("/:id/:idStock", async (req: Request, res: Response) => {
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

    const requestData = {
      real_stock_units: req.body.real_stock_units,
    };

    await apiClient.put<any>(
      `/projects/${req.params.id}/stocks/real-stock/${req.params.idStock}`,
      requestData,
      headers
    );

    setImmediate(() => cache.flushAll());

    const data = {
      success: true,
      message: "Stock actualizado exitosamente",
    };

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

export default router;
