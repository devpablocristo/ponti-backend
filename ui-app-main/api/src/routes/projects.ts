import { Request, Response, Router } from "express";
import { ApiClient, ApiResponse } from "../clients/ApiClient";
import { configService } from "../configService";
import { cache } from ".";

const apiClient = new ApiClient(configService.baseManagerApi);

const router: Router = Router();

// Small utility to simulate latency where needed
//const sleep = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms));

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

    const customer_id = parseInt(req.query.customer_id as string) || 0;
    const campaign_id = parseInt(req.query.campaign_id as string) || 0;
    const name = (req.query.name as string) || "";
    const page = parseInt(req.query.page as string) || 1;
    const perPage = parseInt(req.query.per_page as string) || 10;

    const cacheKey = `projects:${customer_id}:${campaign_id}:${name}:${page}:${perPage}`;

    const cachedProjects = cache.get(cacheKey);
    if (cachedProjects) {
      res.status(200).json(cachedProjects);
      return;
    }

    const { data: projects } = await apiClient.get<any>(
      `/projects?page=${page}&per_page=${perPage}&customer_id=${customer_id}&campaign_id=${campaign_id}&name=${name}`,
      headers
    );

    const adaptedProjects = projects.data.map((project: any) => {
      const client = project.customer?.name || "No client";
      const projectName = project.name;

      const managers =
        project.managers?.map((m: any) => m.name).join(", ") || "No managers";

      const investors =
        project.investors
          ?.map((inv: any) => {
            return `${inv.name} - ${inv.percentage}%`;
          })
          .join(", ") || "No investors";

      return {
        id: project.id,
        name: projectName,
        customer: client,
        campaign: project.campaign?.name || "No campaign",
        managers,
        investors,
      };
    });

    const data = {
      success: true,
      data: {
        data: adaptedProjects,
        total_hectares: projects.total_hectares,
        page_info: {
          per_page: projects.page_info.per_page,
          page: projects.page_info.page,
          max_page: projects.page_info.max_page,
          total: projects.page_info.total,
        },
      },
    };

    setImmediate(() => {
      if (projects.data.length > 0) {
        cache.set(cacheKey, data, 60 * 5);
      }
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

router.get("/customer/:id", async (req: Request, res: Response) => {
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

    const page = parseInt(req.query.page as string) || 1;
    const perPage = parseInt(req.query.per_page as string) || 1000;

    const url = `projects/customer/${id}?page=${page}&per_page=${perPage}`;

    const cachedProjects = cache.get(url);
    if (cachedProjects) {
      res.status(200).json(cachedProjects);
      return;
    }

    const { data: projects } = await apiClient.get<any>(url, headers);

    const data = {
      success: true,
      data: {
        data: projects.data,
        page_info: {
          per_page: projects.page_info.per_page,
          page: projects.page_info.page,
          max_page: projects.page_info.max_page,
          total: projects.page_info.total,
        },
      },
    };

    if (projects.data.length > 0) {
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

router.get("/:id", async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const userId = req.user?.userID;
    if (!userId) {
      res.status(401).json({ message: "Usuario no autenticado" });
      return;
    }

    const cachedProject = cache.get(`project:${id}`);
    if (cachedProject) {
      res.status(200).json(cachedProject);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const data = await apiClient.get<any>(`/projects/${id}`, headers);

    setImmediate(() => cache.set(`project:${id}`, data));

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
    };

    await apiClient.post<any>(`/projects`, requestData, headers);

    setImmediate(() => cache.flushAll());

    res.status(200).json({
      success: true,
      msg: "ok",
    });
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

    const requestData = {
      ...req.body,
    };

    await apiClient.put<any>(`/projects/${id}`, requestData, headers);

    setImmediate(() => cache.flushAll());

    res.status(200).json({
      success: true,
      msg: "ok",
    });
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
  "/:id_project/labors/:id",
  async (req: Request, res: Response) => {
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

      const data = await apiClient.delete<any>(`/projects/${id}`, headers);
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
  }
);

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

    const data = await apiClient.delete<any>(`/projects/${id}`, headers);
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

router.get("/:id/dollar-values", async (req: Request, res: Response) => {
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

    const cachedDollar = cache.get(`dollar:${projectId}`);
    if (cachedDollar) {
      res.status(200).json(cachedDollar);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: dollar } = await apiClient.get<any>(
      `/projects/${projectId}/dollar-values`,
      headers
    );

    const data = {
      success: true,
      data: dollar,
    };

    if (dollar.length > 0) {
      setImmediate(() => cache.set(`dollar:${projectId}`, data));
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

router.put("/:id/dollar-values", async (req: Request, res: Response) => {
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

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const requestData = {
      year: new Date().getFullYear(),
      values: req.body,
    };

    await apiClient.put<any>(
      `/projects/${projectId}/dollar-values`,
      requestData,
      headers
    );

    setImmediate(() => cache.flushAll());

    const data = {
      success: true,
      message: "Dolar actualizado exitosamente",
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

router.post("/:id/labors", async (req: Request, res: Response) => {
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

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const requestData = {
      labors: req.body,
    };

    await apiClient.post<any>(
      `/projects/${projectId}/labors`,
      requestData,
      headers
    );

    setImmediate(() => cache.flushAll());

    const data = {
      success: true,
      message: "Labores actualizados exitosamente",
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

router.get("/:id/labors", async (req: Request, res: Response) => {
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

    const cachedLabors = cache.get(`labors:${projectId}`);
    if (cachedLabors) {
      res.status(200).json(cachedLabors);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: labors } = await apiClient.get<any>(
      `/projects/${projectId}/labors`,
      headers
    );

    const data = {
      success: true,
      data: labors,
    };

    if (labors.length > 0) {
      setImmediate(() => cache.set(`labors:${projectId}`, data));
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

router.get("/:id/commerce", async (req: Request, res: Response) => {
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

    const cachedCommerce = cache.get(`commerce:${projectId}`);
    if (cachedCommerce) {
      res.status(200).json(cachedCommerce);
      return;
    }

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: commerce } = await apiClient.get<any>(
      `/projects/${projectId}/commercializations`,
      headers
    );

    const data = {
      success: true,
      data: commerce,
    };

    if (commerce.length > 0) {
      setImmediate(() => cache.set(`commerce:${projectId}`, data));
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

router.post("/:id/commerce", async (req: Request, res: Response) => {
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

    const headers = {
      "X-API-KEY": configService.apiKey,
      "X-User-Id": userId,
    };

    const { data: commerce } = await apiClient.post<any>(
      `/projects/${projectId}/commercializations`,
      { values: req.body },
      headers
    );

    const data = {
      success: true,
      data: commerce,
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
