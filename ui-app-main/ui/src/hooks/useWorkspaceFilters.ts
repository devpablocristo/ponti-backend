import { useState, useEffect, useCallback } from "react";
import { useSelection } from "../pages/login/context/SelectionContext";
import useCustomers from "./useCustomers";
import useProjects from "./useDatabase/projects";
import useCampaigns from "./useCampaigns";
import useFields from "./useFields";

export interface Customer {
  id: number;
  name: string;
}

export interface Project {
  id: number;
  name: string;
}

export interface Campaign {
  id: number;
  name: string;
  project_id: number;
}

export interface Field {
  id: number;
  name: string;
  project_id: number;
}

export interface FilterBarFilter {
  type: "search" | "select";
  name: string;
  label: string;
  placeholder: string;
  ref?: string;
  total?: number;
  options: Customer[] | Project[] | Campaign[];
  value: any;
  onChange: (value: any) => void;
  setData: (data: any | undefined) => void;
  disabled?: boolean;
}

export interface UseWorkspaceFiltersReturn {
  customers: Customer[];
  projectsDropdown: Project[];
  campaigns: Campaign[];
  fields: Field[];
  projectPageInfo: any; // Replace 'any' with your actual PageInfo type from useProjects
  selectedCustomer: Customer | undefined;
  selectedProject: Project | undefined;
  projectId: number | null;
  selectedCampaignId: number | undefined; // Or Campaign object if you prefer
  selectedField: Field | undefined;
  seasons: { name: string; id: number }[];
  filters: FilterBarFilter[];
  errors: {
    customers: string | null;
    projects: string | null;
    campaigns: string | null;
    fields: string | null;
  };
  loading: {
    customers: boolean;
    projects: boolean;
    campaigns: boolean;
    fields: boolean;
  };
}

type FilterKey = "customer" | "project" | "campaign" | "field";

export const useWorkspaceFilters = (
  enabledFilters: FilterKey[] = ["customer", "project", "campaign", "field"]
): UseWorkspaceFiltersReturn => {
  const {
    customer: contextCustomer,
    setCustomer: contextSetCustomer,
    project: contextProject,
    setProject: contextSetProject,
    projectId: contextProjectId,
    setProjectId: contextSetProjectId,
    campaign: contextCampaign,
    setCampaign: contextSetCampaign,
    field: contextField,
    setField: contextSetField,
    seasons,
  } = useSelection();

  const selectedCustomer = contextCustomer as Customer | undefined;
  const selectedProject = contextProject as Project | undefined;
  const projectId = contextProjectId;
  const selectedCampaign = contextCampaign as Campaign | undefined;
  const selectedCampaignId = selectedCampaign?.id;
  const selectedField = contextField as Field | undefined;

  const setSelectedCustomer: React.Dispatch<
    React.SetStateAction<Customer | undefined>
  > = (value) => {
    if (typeof value === "function") {
      // Not supported for contextSetCustomer
      return;
    }
    contextSetCustomer(value);
    contextSetProject(undefined);
    contextSetCampaign(undefined);
    contextSetField(undefined);
    contextSetProjectId(undefined);
  };

  const setSelectedProject: React.Dispatch<
    React.SetStateAction<Project | undefined>
  > = (value) => {
    if (typeof value === "function") {
      return;
    }
    contextSetProject(value);
    contextSetCampaign(undefined);
    contextSetField(undefined);
    contextSetProjectId(undefined);
  };

  const setSelectedCampaign: (campaign: Campaign | undefined) => void = (
    campaign
  ) => {
    if (campaign && selectedProject) {
      const updatedProject = { ...selectedProject, id: campaign.project_id };
      contextSetProject(updatedProject);
      contextSetProjectId(campaign.project_id);
      contextSetCampaign(campaign);
    }
  };

  const setSelectedField: React.Dispatch<
    React.SetStateAction<Field | undefined>
  > = (value) => {
    if (typeof value === "function") {
      return;
    }
    if (value?.id === 0) {
      contextSetField(undefined);
      return;
    }
    contextSetField(value);
  };

  const [queryCustomer, setQueryCustomer] = useState<string>("");
  const [queryProject, setQueryProject] = useState<string>("");

  const {
    customers,
    getCustomers,
    total: totalCustomers,
    processing: loadingCustomers,
    error: loadingCustomersError,
  } = useCustomers();

  const {
    projectsDropdown,
    getProjectsDropdown,
    projectsDropdownPagination: projectPageInfo,
    processingDropdown: loadingProjects,
    error: loadingProjectsError,
  } = useProjects();

  const {
    campaigns,
    getCampaigns,
    total: totalCampaigns,
    processing: loadingCampaigns,
    error: loadingCampaignsError,
  } = useCampaigns();

  const {
    fields,
    getFields,
    total: totalFields,
    processing: loadingFields,
    error: loadingFieldsError,
  } = useFields();

  useEffect(() => {
    if (selectedCustomer) {
      setQueryCustomer(selectedCustomer.name);
    }
  }, [selectedCustomer]);

  useEffect(() => {
    if (selectedProject) {
      setQueryProject(selectedProject.name);
    } else {
      setQueryProject("Todos los proyectos");
    }
  }, [selectedProject]);

  const filters: FilterBarFilter[] = [];

  useEffect(() => {
    if (enabledFilters.includes("customer")) {
      //TODO: limit=1000, implement pagination
      getCustomers("limit=1000");
    }
  }, [getCustomers]);

  const handleSetCustomer = useCallback(
    (customer: Customer | undefined) => {
      setSelectedCustomer(customer);
      setSelectedProject(undefined);
      setQueryProject("Todos los proyectos");
      setSelectedCampaign(undefined);
    },
    [setSelectedCustomer, setSelectedProject, setSelectedCampaign]
  );

  if (enabledFilters.includes("customer")) {
    filters.push({
      type: "search",
      name: "cliente",
      ref: "client",
      label: "Cliente",
      placeholder: "Buscar cliente",
      options: customers || [],
      total: totalCustomers,
      value: queryCustomer,
      onChange: setQueryCustomer,
      setData: handleSetCustomer,
      disabled: loadingCustomers,
    });
  }

  useEffect(() => {
    if (enabledFilters.includes("project")) {
      if (selectedCustomer && selectedCustomer.id !== 0) {
        getProjectsDropdown(selectedCustomer.id);
      }
    }
  }, [selectedCustomer, getProjectsDropdown]);

  const handleSetProject = useCallback(
    (project: Project | undefined) => {
      setSelectedProject(project);
      setSelectedCampaign(undefined);
      //contextSetProjectId(null);
    },
    [setSelectedProject, setSelectedCampaign]
  );

  if (enabledFilters.includes("project")) {
    filters.push({
      type: "search",
      name: "proyecto",
      ref: "project",
      label: "Proyecto",
      placeholder: "Buscar proyecto",
      options: selectedCustomer ? projectsDropdown || [] : [],
      total: projectPageInfo?.total || 0,
      value: queryProject,
      onChange: setQueryProject,
      setData: handleSetProject,
      disabled:
        loadingProjects ||
        !selectedCustomer ||
        (selectedCustomer.id === 0 && !selectedCampaign),
    });
  }

  useEffect(() => {
    if (
      enabledFilters.includes("campaign") &&
      selectedCustomer &&
      selectedCustomer.id !== 0 &&
      selectedProject &&
      selectedProject.id !== 0
    ) {
      getCampaigns(
        `customer_id=${selectedCustomer.id}&project_name=${selectedProject.name}&limit=100`
      );
    }
  }, [selectedCustomer, selectedProject, getCampaigns]);

  if (enabledFilters.includes("campaign")) {
    filters.push({
      type: "select",
      name: "campaña",
      label: "Campaña",
      placeholder: "Seleccione campaña",
      options: selectedCustomer && selectedProject ? campaigns || [] : [],
      total: selectedCustomer ? totalCampaigns : 0,
      value: selectedCampaignId,
      onChange: () => {},
      setData: setSelectedCampaign,
      disabled:
        !selectedCustomer ||
        selectedCustomer.id === 0 ||
        !selectedProject ||
        selectedProject.id === 0 ||
        loadingCampaigns,
    });
  }

  useEffect(() => {
    if (enabledFilters.includes("field") && projectId && projectId !== 0) {
      getFields(`project_id=${projectId}`);
    }
  }, [getFields, projectId]);

  if (enabledFilters.includes("field")) {
    filters.push({
      type: "select",
      name: "campo",
      label: "Campo",
      placeholder: "Seleccione campo",
      options:
        selectedCustomer && selectedProject
          ? [{ id: 0, name: "Todos los campos" }, ...(fields || [])]
          : [],
      total: selectedCustomer ? totalFields : 0,
      value: selectedField?.id,
      onChange: () => {},
      setData: setSelectedField,
      disabled:
        !selectedCustomer ||
        selectedCustomer.id === 0 ||
        !selectedProject ||
        selectedProject.id === 0 ||
        loadingFields,
    });
  }

  return {
    customers: customers || [],
    campaigns: campaigns || [],
    projectsDropdown: projectsDropdown || [],
    fields: (fields as Field[]) || [],
    projectPageInfo: projectPageInfo,
    selectedCustomer,
    selectedProject,
    projectId,
    selectedCampaignId,
    selectedField,
    filters,
    seasons,
    loading: {
      customers: loadingCustomers,
      projects: loadingProjects,
      campaigns: loadingCampaigns,
      fields: loadingFields,
    },
    errors: {
      customers: loadingCustomersError,
      campaigns: loadingCampaignsError,
      projects: loadingProjectsError,
      fields: loadingFieldsError,
    },
  };
};
