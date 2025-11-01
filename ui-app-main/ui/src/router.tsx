import { ProtectedLayout } from "./layout/ProtectedLayout";
import ErrorPage from "./pages/ErrorPage";
import { Navigate } from "react-router-dom";
import { BaseLayout } from "./layout/BaseLayout";
import SignInPage from "./pages/login/Login";
import { Dashboard } from "./pages/admin/dashboard/Dashboard";
import { Products } from "./pages/admin/products/Products";
import { Users } from "./pages/admin/users/Users";
import { Profile } from "./pages/admin/profile/Profile";
import { Tasks } from "./pages/admin/tasks/Tasks";
import { WorkOrders } from "./pages/admin/workorders/WorkOrders";
import { Stock } from "./pages/admin/stock/Stock";
import { FormUser } from "./pages/admin/users/FormUser";
import Customers from "./pages/admin/customers/Customers";
import DatabaseCustomers from "./pages/admin/database/customers/Customers";
import DatabaseTasksForm from "./pages/admin/database/tasks/TasksForm";
import Lots from "./pages/admin/lots/Lots";
import Items from "./pages/admin/database/products/Items";
import DollarForm from "./pages/admin/database/dollar/DollarForm";
import CommerceForm from "./pages/admin/database/commerce/CommerceForm";
import WorkspaceSelectorPage from "./pages/login/WorkspaceSelector";
import ListItems from "./pages/admin/database/products/List";
import ListTasks from "./pages/admin/database/tasks/List";
import ByFieldOrCropReport from "./pages/admin/reports/ByFieldOrCropReport.tsx";
import SummaryResultsReport from "./pages/admin/reports/SummaryResultsReport.tsx";
import InvestorContribution from "./pages/admin/reports/InvestorContributionReport.tsx";

export default [
  {
    path: "",
    element: <Navigate to="/admin" />,
    errorElement: <ErrorPage />,
  },
  {
    path: "login",
    element: (
      <BaseLayout>
        <SignInPage />
      </BaseLayout>
    ),
  },
  {
    path: "workspace",
    element: (
      <BaseLayout>
        <WorkspaceSelectorPage />
      </BaseLayout>
    ),
  },
  {
    path: "/admin",
    element: <ProtectedLayout />,
    errorElement: <ErrorPage />,
    children: [
      {
        path: "dashboard",
        element: <Dashboard />,
      },
      {
        path: "customers",
        element: <Customers />,
      },
      {
        path: "lots",
        element: <Lots />,
      },
      {
        path: "products",
        element: <Products />,
      },
      {
        path: "tasks",
        element: <Tasks />,
      },
      {
        path: "stock",
        element: <Stock />,
      },
      {
        path: "work-orders",
        element: <WorkOrders />,
      },
      {
        path: "database/customers",
        element: <DatabaseCustomers />,
      },
      {
        path: "database/customers/:id",
        element: <DatabaseCustomers />,
      },
      {
        path: "database/tasks",
        element: <DatabaseTasksForm />,
      },
      {
        path: "database/items",
        element: <Items />,
      },
      {
        path: "database/items/list",
        element: <ListItems />,
      },
      {
        path: "database/tasks/list",
        element: <ListTasks />,
      },
      {
        path: "database/dollar",
        element: <DollarForm />,
      },
      {
        path: "database/commerce",
        element: <CommerceForm />,
      },
      {
        path: "users",
        element: <Users />,
      },
      {
        path: "users/new",
        element: <FormUser />,
      },
      {
        path: "users/edit/:id",
        element: <FormUser />,
      },
      {
        path: "profile",
        element: <Profile />,
      },
      {
        path: "informes/aportes",
        element: <InvestorContribution />,
      },
      {
        path: "informes/resumen",
        element: <SummaryResultsReport />,
      },
      {
        path: "informes/campo",
        element: <ByFieldOrCropReport />,
      },
      {
        path: "",
        element: <Navigate to="/admin/dashboard" />,
        errorElement: <ErrorPage />,
      },
      {
        path: "*",
        element: <Navigate to="/admin/dashboard" />,
        errorElement: <ErrorPage />,
      },
    ],
  },
  {
    path: "*",
    element: <Navigate to="/admin/dashboard" />,
    errorElement: <ErrorPage />,
  },
];
