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

    const cachedOptions = cache.get("options");
    if (cachedOptions) {
      res.status(200).json(cachedOptions);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: customers } = await apiClient.get<any>("/customers", headers);
    const { data: managers } = await apiClient.get<any>("/managers", headers);
    const { data: investors } = await apiClient.get<any>("/investors", headers);
    const { data: campaigns } = await apiClient.get<any>("/campaigns", headers);
    const { data: crops } = await apiClient.get<any>("/crops", headers);
    const { data: leaseTypes } = await apiClient.get<any>(
      "/lease-types",
      headers
    );

    const data = {
      success: true,
      data: {
        clients: customers.data,
        managers: managers,
        investors: investors.data,
        campaigns: campaigns,
        crops: crops.data,
        rentTypes: leaseTypes,
      },
    };

    setImmediate(() => {
      cache.set("options", data, 60 * 5);
    });

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
