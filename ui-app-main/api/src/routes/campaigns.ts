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

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const customerId = parseInt(req.query.customer_id as string) || 0;
    const projectName = (req.query.project_name as string) || "";

    const url = `campaigns?customer_id=${customerId}&project_name=${projectName}`;

    const cachedCampaigns = cache.get(url);
    if (cachedCampaigns) {
      res.status(200).json(cachedCampaigns);
      return;
    }

    const { data: campaigns } = await apiClient.get<any>(url, headers);

    const data = {
      success: true,
      data: {
        data: campaigns,
        total: campaigns.length,
      },
    };

    if (campaigns.length > 0) {
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

export default router;
