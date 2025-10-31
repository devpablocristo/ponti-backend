import { Request, Response, Router } from "express";
import { ApiClient, ApiResponse } from "../clients/ApiClient";
import { configService } from "../configService";
import { cache } from ".";

const apiClient = new ApiClient(configService.baseManagerApi);
const router: Router = Router();

const months = [
  "Enero",
  "Febrero",
  "Marzo",
  "Abril",
  "Mayo",
  "Junio",
  "Julio",
  "Agosto",
  "Septiembre",
  "Octubre",
  "Noviembre",
  "Diciembre",
];

router.post("/invoice", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    if (!req.body.workorder_id || req.body.workorder_id === 0) {
      res.status(400).json({ message: "El id de la orden es obligatorio" });
      return;
    }

    if (!req.body.invoice_number || req.body.invoice_number === 0) {
      res
        .status(400)
        .json({ message: "El número de la factura es obligatorio" });
      return;
    }

    if (!req.body.invoice_date) {
      res
        .status(400)
        .json({ message: "La fecha de la factura es obligatoria" });
      return;
    }

    if (!req.body.invoice_company) {
      res
        .status(400)
        .json({ message: "La empresa de la factura es obligatoria" });
      return;
    }

    if (!req.body.invoice_status) {
      res
        .status(400)
        .json({ message: "El estado de la factura es obligatorio" });
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const requestData = {
      number: req.body.invoice_number,
      date: req.body.invoice_date + "T00:00:00Z",
      company: req.body.invoice_company,
      status: req.body.invoice_status,
    };

    const { data: workorder } = await apiClient.post<any>(
      `/invoice/${req.body.workorder_id}`,
      requestData,
      headers
    );

    setImmediate(() => cache.flushAll());

    const data = {
      success: true,
      data: workorder.data,
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

router.put("/invoice/:id", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    if (!req.body.workorder_id || req.body.workorder_id === 0) {
      res.status(400).json({ message: "El id de la orden es obligatorio" });
      return;
    }

    if (!req.body.invoice_number || req.body.invoice_number === 0) {
      res
        .status(400)
        .json({ message: "El número de la factura es obligatorio" });
      return;
    }

    if (!req.body.invoice_date) {
      res
        .status(400)
        .json({ message: "La fecha de la factura es obligatoria" });
      return;
    }

    if (!req.body.invoice_company) {
      res
        .status(400)
        .json({ message: "La empresa de la factura es obligatoria" });
      return;
    }

    if (!req.body.invoice_status) {
      res
        .status(400)
        .json({ message: "El estado de la factura es obligatorio" });
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const requestData = {
      number: req.body.invoice_number,
      date: req.body.invoice_date + "T00:00:00Z",
      company: req.body.invoice_company,
      status: req.body.invoice_status,
    };

    const { data: workorder } = await apiClient.put<any>(
      `/invoice/${req.body.workorder_id}`,
      requestData,
      headers
    );

    setImmediate(() => cache.flushAll());

    const data = {
      success: true,
      data: workorder.data,
    };

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

router.get("/metrics/:id", async (req: Request, res: Response) => {
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
    const project_id = parseInt(req.params.id as string) || 0;

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
      `/labors/metrics${query}`,
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
      `/labors/export/${project_id}?usd_month=${new Date().getMonth() + 1}`,
      { headers, responseType: "arraybuffer" }
    );

    res.setHeader("Content-Disposition", 'attachment; filename="labores.xlsx"');
    res.setHeader(
      "Content-Type",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    );

    res.send(response.data);
  } catch (error) {
    console.error(error);
    res.status(500).send("Error al exportar labores");
  }
});

router.get("/database-export/:id", async (req, res) => {
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
      `/labors/export/all?project_id=${project_id}&usd_month=${new Date().getMonth() + 1}`,
      { headers, responseType: "arraybuffer" }
    );

    res.setHeader("Content-Disposition", 'attachment; filename="laboresdb.xlsx"');
    res.setHeader(
      "Content-Type",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    );

    res.send(response.data);
  } catch (error) {
    console.error(error);
    res.status(500).send("Error al exportar labores");
  }
});

router.get("/:id", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const project_id = parseInt(req.params.id as string) || 0;

    const field_id = parseInt(req.query.field_id as string) || 0;
    const page = parseInt(req.query.page as string) || 1;
    const perPage = parseInt(req.query.per_page as string) || 1000;

    if (field_id === 0 && project_id === 0) {
      res.status(400).json({ message: "Campo o proyecto obligatorio" });
      return;
    }

    const currentMonthIndex = new Date().getMonth();
    const currentMonthName = months[currentMonthIndex];

    let query = `?usd_month=${currentMonthName}&page=${page}&per_page=${perPage}`;
    if (field_id !== 0) {
      query += `&field_id=${field_id}`;
    }

    const cachedLabors = cache.get(
      `labors:project:${project_id}:query:${query}`
    );
    if (cachedLabors) {
      res.status(200).json(cachedLabors);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: labors } = await apiClient.get<any>(
      `/labors/group/${project_id}${query}`,
      headers
    );

    const data = {
      success: true,
      data: {
        data: labors.data,
        page_info: labors.page_info,
      },
    };

    setImmediate(() =>
      cache.set(`labors:project:${project_id}:query:${query}`, data)
    );

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

    const { data: labor } = await apiClient.delete<any>(
      `/labors/${req.params.id}`,
      headers
    );

    const data = {
      success: true,
      data: labor,
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

router.put("/projects/:project_id/:id", async (req: Request, res: Response) => {
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
      id: req.body.id,
      name: req.body.name,
      category_id: req.body.category_id,
      price: req.body.price,
      contractor_name: req.body.contractor_name,
    };

    const { data: workorder } = await apiClient.put<any>(
      `projects/${req.params.project_id}/labors/${req.params.id}`,
      requestData,
      headers
    );

    setImmediate(() => cache.flushAll());

    const data = {
      success: true,
      data: workorder.data,
    };

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

export default router;
