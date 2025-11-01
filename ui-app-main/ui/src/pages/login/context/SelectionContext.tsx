// contexts/SelectionContext.tsx
import { createContext, useContext, useState, useEffect } from "react";
import { Entity } from "../../../hooks/useDatabase/options/types";
import { ProjectData } from "../../../hooks/useDatabase/projects/types";
import { Data } from "../../../hooks/useFields/types";

type SelectionContextType = {
  customer: Entity;
  setCustomer: (c: Entity | undefined) => void;
  project: ProjectData;
  setProject: (p: Entity | undefined) => void;
  projectId: number | null;
  setProjectId: (p: number | undefined) => void;
  campaign: Entity;
  setCampaign: (c: Entity | undefined) => void;
  field: Data;
  setField: (f: Data | undefined) => void;
  seasons: { name: string; id: number }[];
};

const SelectionContext = createContext<SelectionContextType | null>(null);

export const SelectionProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const seasons = [
    { name: "Otoño", id: 1 },
    { name: "Invierno", id: 2 },
    { name: "Primavera", id: 3 },
    { name: "Verano", id: 4 },
  ];

  const [customer, setCustomer] = useState(() => {
    const stored = localStorage.getItem("customer");
    return stored ? JSON.parse(stored) : null;
  });

  const [project, setProject] = useState(() => {
    const stored = localStorage.getItem("project");
    return stored ? JSON.parse(stored) : null;
  });

  const [projectId, setProjectId] = useState(() => {
    const stored = localStorage.getItem("project_id");
    return stored ? JSON.parse(stored) : null;
  });

  const [campaign, setCampaign] = useState(() => {
    const stored = localStorage.getItem("campaign");
    return stored ? JSON.parse(stored) : null;
  });

  const [field, setField] = useState(() => {
    const stored = localStorage.getItem("field");
    return stored ? JSON.parse(stored) : null;
  });

  useEffect(() => {
    if (customer === undefined || customer === null) {
      localStorage.removeItem("customer");
    } else {
      localStorage.setItem("customer", JSON.stringify(customer));
    }
  }, [customer]);

  useEffect(() => {
    if (project === undefined || project === null) {
      localStorage.removeItem("project");
    } else {
      localStorage.setItem("project", JSON.stringify(project));
    }
  }, [project]);

  useEffect(() => {
    if (projectId === undefined || projectId === null) {
      localStorage.removeItem("project_id");
    } else {
      localStorage.setItem("project_id", projectId);
    }
  }, [projectId]);

  useEffect(() => {
    if (campaign === undefined || campaign === null) {
      localStorage.removeItem("campaign");
    } else {
      localStorage.setItem("campaign", JSON.stringify(campaign));
      localStorage.setItem("project_id", campaign.project_id);
    }
  }, [campaign]);

  useEffect(() => {
    if (field === undefined || field === null) {
      localStorage.removeItem("field");
    } else {
      localStorage.setItem("field", JSON.stringify(field));
    }
  }, [field]);

  return (
    <SelectionContext.Provider
      value={{
        customer,
        setCustomer,
        project,
        setProject,
        projectId,
        setProjectId,
        campaign,
        setCampaign,
        field,
        setField,
        seasons,
      }}
    >
      {children}
    </SelectionContext.Provider>
  );
};

export const useSelection = () => {
  const ctx = useContext(SelectionContext);
  if (!ctx)
    throw new Error("useSelection debe usarse dentro de SelectionProvider");
  return ctx;
};
