import { Request, Response, Router } from "express";
import { ApiClient, ApiResponse } from "../clients/ApiClient";
import { configService } from "../configService";
import { cache } from ".";

const apiClient = new ApiClient(configService.baseManagerApi);
const router: Router = Router();

router.get("", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const field_id = parseInt(req.query.field_id as string) || 0;
    const project_id = parseInt(req.query.project_id as string) || 0;
    if (field_id === 0 && project_id === 0) {
      res.status(400).json({ message: "Campo o proyecto obligatorio" });
      return;
    }

    const key = `lots:field:${field_id}:project:${project_id}`;

    const cachedLots = cache.get(key);
    if (cachedLots) {
      res.status(200).json(cachedLots);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const page = parseInt(req.query.page as string) || 1;
    const perPage = parseInt(req.query.per_page as string) || 1000;

    const { data: lots } = await apiClient.get<any>(
      `/lots?field_id=${field_id}&project_id=${project_id}&page=${page}&per_page=${perPage}`,
      headers
    );

    const adaptedLots = lots.items.map((lot: any) => {
      return {
        ...lot,
        harvest_date: lot.dates,
        cost_per_hectare: lot.cost_usd_per_ha,
        yield: lot.yield_tn_per_ha,
        sowed_area: lot.hectares,
        net_income: lot.income_net_per_ha,
        cost_us_ha: lot.cost_usd_per_ha,
        rent: lot.rent_per_ha,
        total_assets: lot.active_total_per_ha,
        operating_result: lot.operating_result_per_ha,
      };
    });

    const data = {
      success: true,
      data: {
        data: adaptedLots,
        page_info: lots.page_info,
      },
    };

    if (lots.items.length > 0) {
      setImmediate(() => cache.set(key, data));
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

router.get("/kpis", async (req: Request, res: Response) => {
  try {
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const field_id = parseInt(req.query.field_id as string) || 0;
    const project_id = parseInt(req.query.project_id as string) || 0;
    if (field_id === 0 && project_id === 0) {
      res.status(400).json({ message: "Campo o proyecto obligatorio" });
      return;
    }

    let key = `kpis:lots:field:${field_id}`;
    if (field_id === 0 && project_id !== 0) {
      key = `kpis:lots:project:${project_id}`;
    }

    const cachedLots = cache.get(key);
    if (cachedLots) {
      res.status(200).json(cachedLots);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: metrics } = await apiClient.get<any>(
      `/lots/metrics?field_id=${field_id}&project_id=${project_id}`,
      headers
    );

    const data = {
      success: true,
      data: metrics,
    };

    if (metrics.length > 0) {
      setImmediate(() => cache.set(key, data));
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
      `/lots/export?project_id=${project_id}`,
      { headers, responseType: "arraybuffer" }
    );

    res.setHeader("Content-Disposition", 'attachment; filename="lots.xlsx"');
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
      name: req.body.lot_name,
      hectares: parseFloat(req.body.sowed_area),
      season: req.body.season,
      current_crop_id: req.body.current_crop_id,
      previous_crop_id: req.body.previous_crop_id,
      variety: req.body.variety,
      dates: req.body.dates,
      updated_at: req.body.updated_at,
    };

    await apiClient.put<any>(`/lots/${req.params.id}`, requestData, headers);

    setImmediate(() => cache.flushAll());

    const data = {
      success: true,
      message: "Lote actualizado exitosamente",
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

router.put("/:id/tons", async (req: Request, res: Response) => {
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

    const tons = parseInt(req.body.tons) || 0;
    if (tons === 0) {
      res.status(400).json({ message: "Toneladas obligatorio" });
      return;
    }

    await apiClient.put<any>(
      `/lots/${req.params.id}/tons`,
      { tons: tons.toString() },
      headers
    );

    setImmediate(() => cache.flushAll());

    const data = {
      success: true,
      message: "Lote actualizado exitosamente",
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
