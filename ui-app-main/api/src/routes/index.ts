import { Router } from "express";

import auth from "./auth";
import users from "./users";
import options from "./options";
import projects from "./projects";
import customers from "./customers";
import campaigns from "./campaigns";
import fields from "./fields";
import lots from "./lots";
import crops from "./crops";
import supplies from "./supplies";
import { verifyToken } from "./authMiddleware";
import NodeCache from "node-cache";
import categories from "./categories";
import types from "./types";
import workorders from "./workorders";
import labors from "./labors";
import providers from "./providers";
import movements from "./movements";
import stock from "./stock";
import dashboard from "./dashboard";
import reports from "./reports";

const router: Router = Router();
export const cache = new NodeCache({ stdTTL: 1800, checkperiod: 1800 });

router.get("/ping", (req, res) => {
  res.status(200).json({ message: "UI says Pong!" });
});

router.use("/auth", auth);

router.use(verifyToken);

router.use("/users", users);
router.use("/projects", projects);
router.use("/customers", customers);
router.use("/campaigns", campaigns);
router.use("/fields", fields);
router.use("/lots", lots);
router.use("/crops", crops);
router.use("/supplies", supplies);
router.use("/categories", categories);
router.use("/types", types);
router.use("/workorders", workorders);
router.use("/labors", labors);
router.use("/providers", providers);
router.use("/supply_movements", movements);
router.use("/stock", stock);
router.use("/dashboard", dashboard);
router.use("/reports", reports);

router.use("/form-options", options);

export default router;
