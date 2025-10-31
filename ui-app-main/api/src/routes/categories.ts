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

    let key = "categories";
    const type_id = parseInt(req.query.type_id as string) || 0;
    if (type_id !== 0) {
      key = `categories:${type_id}`;
    }

    const cachedCategories = cache.get(key);
    if (cachedCategories) {
      res.status(200).json(cachedCategories);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    let { data: categories } = await apiClient.get<any>("/categories", headers);
    if (Array.isArray(categories) && type_id !== 0) {
      categories = categories.filter(
        (category: any) => category.type_id === type_id
      );
    }

    const data = {
      success: true,
      data: categories,
    };

    setImmediate(() => {
      cache.set(key, data, 60 * 5);
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
