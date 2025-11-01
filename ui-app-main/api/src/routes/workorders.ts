import { Request, Response, Router } from "express";
import { ApiClient, ApiResponse } from "../clients/ApiClient";
import { configService } from "../configService";
import { cache } from ".";

const apiClient = new ApiClient(configService.baseManagerApi);
const router: Router = Router();

router.post("", async (req: Request, res: Response) => {
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
      ...req.body,
      date: new Date(req.body.date).toISOString(),
    };

    const { data: workorder } = await apiClient.post<any>(
      "/workorders",
      requestData,
      headers
    );

    const data = {
      success: true,
      data: workorder.data,
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
});

router.put("/:id", async (req: Request, res: Response) => {
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
      ...req.body,
      date: new Date(req.body.date).toISOString(),
    };

    const { data: workorder } = await apiClient.put<any>(
      `/workorders/${req.params.id}`,
      requestData,
      headers
    );

    const data = {
      success: true,
      data: workorder.data,
    };

    setImmediate(() => cache.flushAll());

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

router.get("", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const field_id = parseInt(req.query.field_id as string) || 0;
    const project_id = parseInt(req.query.project_id as string) || 0;
    const page = parseInt(req.query.page as string) || 1;
    const perPage = parseInt(req.query.per_page as string) || 1000;

    if (field_id === 0 && project_id === 0) {
      res.status(400).json({ message: "Campo o proyecto obligatorio" });
      return;
    }

    let query = `?page=${page}&per_page=${perPage}`;
    if (field_id !== 0) {
      query += `&field_id=${field_id}`;
    } else {
      query += `&project_id=${project_id}`;
    }

    const cachedWorkorders = cache.get(`workorders:query:${query}`);
    if (cachedWorkorders) {
      res.status(200).json(cachedWorkorders);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: workorders } = await apiClient.get<any>(
      `/workorders${query}`,
      headers
    );

    const data = {
      success: true,
      data: {
        data: workorders.items,
        page_info: workorders.page_info,
      },
    };

    setImmediate(() => cache.set(`workorders:query:${query}`, data));

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

router.get("/metrics", async (req: Request, res: Response) => {
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

    const field_id = parseInt(req.query.field_id as string) || 0;
    const project_id = parseInt(req.query.project_id as string) || 0;

    if (field_id === 0 && project_id === 0) {
      res.status(400).json({ message: "Campo o proyecto obligatorio" });
      return;
    }

    let query = "";
    if (field_id > 0) {
      query = `?field_id=${field_id}`;
    } else if (project_id > 0) {
      query = `?project_id=${project_id}`;
    }

    const { data: metrics } = await apiClient.get<any>(
      `/workorders/metrics${query}`,
      headers
    );

    const data = {
      success: true,
      data: metrics,
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
      `/workorders/export?project_id=${project_id}`,
      { headers, responseType: "arraybuffer" }
    );

    res.setHeader("Content-Disposition", 'attachment; filename="ordenes.xlsx"');
    res.setHeader(
      "Content-Type",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    );

    res.send(response.data);
  } catch (error) {
    console.error(error);
    res.status(500).send("Error al exportar ordenes");
  }
});

router.get("/:id", async (req: Request, res: Response) => {
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

    const { data: workorder } = await apiClient.get<any>(
      `/workorders/${req.params.id}`,
      headers
    );

    const data = {
      success: true,
      data: workorder,
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

router.delete("/:id", async (req: Request, res: Response) => {
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

    const { data: workorder } = await apiClient.delete<any>(
      `/workorders/${req.params.id}`,
      headers
    );

    const data = {
      success: true,
      data: workorder,
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
});

export default router;
