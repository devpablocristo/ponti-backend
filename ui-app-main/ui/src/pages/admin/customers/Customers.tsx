import React, { useEffect, useState } from "react";
import { LoaderCircle } from "lucide-react";

import DataTable from "../../../components/Table/DataTable";
import { ProjectData } from "../../../hooks/useDatabase/projects/types";
import FilterBar from "../../../layout/FilterBar/FilterBar";
import { useNavigate } from "react-router-dom";
import { useWorkspaceFilters } from "../../../hooks/useWorkspaceFilters";
import useProjects from "../../../hooks/useDatabase/projects";
import ExpandedRow from "./ExpandedRow";
import { BaseModal } from "../../../components/Modal/BaseModal";

const columns: Column<ProjectData>[] = [
  { key: "customer", header: "Cliente/Sociedad" },
  {
    key: "name",
    header: "Proyecto",
    render: (value, data) => (
      <strong className="text-blue-700">
        <a href={`/admin/database/customers/${data.id}`}>
          {value as string} ({data.campaign})
        </a>
      </strong>
    ),
  },
  { key: "managers", header: "Responsable" },
  {
    key: "investors",
    header: "Inversores y aportes",
  },
];

export function Customers() {
  const navigate = useNavigate();

  const [isProcessing, setIsProcessing] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalConfig, setModalConfig] = useState({
    title: "",
    message: "",
    primaryButtonText: "",
    secondaryButtonText: "Cancelar",
    onConfirm: () => {},
  });

  const {
    projects,
    totalHectares,
    getProjects,
    deleteProject,
    pageInfo: projectPageInfo,
    processing,
    error,
  } = useProjects();

  const {
    selectedCustomer,
    selectedCampaignId,
    selectedProject,
    filters,
    loading, // Contains loading.customers, loading.projects, loading.campaigns
    errors,
  } = useWorkspaceFilters(["customer", "project", "campaign"]);

  useEffect(() => {
    if (selectedCustomer) {
      getProjects(
        `customer_id=${selectedCustomer.id}${
          selectedProject && selectedProject.id !== 0
            ? `&name=${selectedProject.name}`
            : ""
        }${
          selectedCampaignId && selectedCampaignId !== 0
            ? `&campaign_id=${selectedCampaignId}`
            : ""
        }`
      );
    }
  }, [selectedCustomer, selectedCampaignId, selectedProject, getProjects]);

  const renderExpandedRow = (rowData: ProjectData) => {
    return <ExpandedRow projectId={rowData.id} />;
  };

  const handlePageChange = (newPage: number) => {
    let queryString = `page=${newPage}&limit=${
      projectPageInfo?.per_page || 10
    }&customer_id=${selectedCustomer?.id}${
      selectedProject && selectedProject.id !== 0
        ? `&name=${selectedProject.name}`
        : ""
    }${
      selectedCampaignId && selectedCampaignId !== 0
        ? `&campaign_id=${selectedCampaignId}`
        : ""
    }`;
    getProjects(queryString);
  };

  const handlePreFinish = (id: number) => {
    setModalConfig({
      title: "Confirmar eliminación",
      message: "¿Está seguro que desea eliminar el proyecto?",
      primaryButtonText: "Sí, eliminar",
      secondaryButtonText: "Cancelar",
      onConfirm: () => handleFinishConfirmed(id),
    });
    setIsModalOpen(true);
  };

  const handleFinishConfirmed = async (id: number) => {
    setIsProcessing(true);

    try {
      await deleteProject(Number(id));
      getProjects(
        `customer_id=${selectedCustomer?.id}${
          selectedProject && selectedProject.id !== 0
            ? `&name=${selectedProject.name}`
            : ""
        }${
          selectedCampaignId && selectedCampaignId !== 0
            ? `&campaign_id=${selectedCampaignId}`
            : ""
        }`
      );
    } catch (error) {
      console.error("Error al finalizar:", error);
    } finally {
      setModalConfig({
        title: "Confirmación",
        message: "El proyecto ha sido eliminado.",
        primaryButtonText: "Volver",
        secondaryButtonText: "Volver",
        onConfirm: () => {
          window.location.href = "/admin/customers";
        },
      });
      setIsModalOpen(true);
      setIsProcessing(false);
    }
  };

  return (
    <div>
      <FilterBar
        filters={filters}
        actions={[
          {
            label: "+ Nuevo cliente",
            variant: "success",
            isPrimary: true,
            href: "/admin/database/customers",
          },
        ]}
      >
        <div className="flex items-center gap-2 bg-white rounded-xl px-4 shadow-sm py-2 border border-gray-400">
          <div>
            <p className="text-sm text-gray-900">
              Superficie total de hectáreas
            </p>
            <p className="text-xl font-bold text-gray-600">
              {totalHectares} Has
            </p>
          </div>
        </div>
      </FilterBar>

      {errors.customers ||
        errors.projects ||
        (errors.campaigns && (
          <div
            className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
            role="alert"
          >
            <span className="font-medium">Error!</span> {errors.customers}{" "}
            {errors.projects} {errors.campaigns}
          </div>
        ))}

      <div className="mt-2 relative">
        {processing ||
          loading.campaigns ||
          loading.customers ||
          (loading.projects && (
            <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
              <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
            </div>
          ))}
        <DataTable
          data={projects}
          columns={columns}
          expandableRowRender={renderExpandedRow}
          className={`${processing ? "pointer-events-none opacity-60" : ""}`}
          onCopy={(item) => navigate(`/admin/database/customers?id=${item.id}`)}
          onDelete={(item) => handlePreFinish(item.id)}
          pagination={
            projectPageInfo
              ? {
                  page: projectPageInfo.page,
                  perPage: projectPageInfo.per_page,
                  total: projectPageInfo.total,
                  onPageChange: handlePageChange,
                }
              : undefined
          }
        />
        {error && (
          <div
            className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
            role="alert"
          >
            <span className="font-medium">Error!</span> {error}
          </div>
        )}
      </div>
      <BaseModal
        isOpen={isModalOpen}
        isSaving={isProcessing}
        onClose={() => setIsModalOpen(false)}
        title={modalConfig.title}
        message={modalConfig.message}
        primaryButtonText={modalConfig.primaryButtonText}
        secondaryButtonText={modalConfig.secondaryButtonText}
        onPrimaryAction={() => {
          modalConfig.onConfirm();
          setIsModalOpen(false);
        }}
      >
        <div className="flex flex-col items-center gap-2">
          <p>{modalConfig.message}</p>
        </div>
      </BaseModal>
    </div>
  );
}

type Column<T> = {
  key: keyof T;
  header: string;
  render?: (value: any, item: T) => React.ReactNode;
};

export default Customers;
