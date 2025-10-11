import { createRoot } from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import "flowbite";

import "./index.css";
import routes from "./router";

createRoot(document.getElementById("root")!).render(
  <RouterProvider router={createBrowserRouter(routes)} />
);
