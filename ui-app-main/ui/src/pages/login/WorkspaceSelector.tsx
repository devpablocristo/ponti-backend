import { useNavigate } from "react-router-dom";
import { useEffect, useRef, useState } from "react";

import { AuthProvider, useAuth } from "./context/AuthProvider";
import { Entity } from "../../hooks/useDatabase/options/types";
import useCustomers from "../../hooks/useCustomers";
import Cover from "./Cover";
import useCampaigns from "../../hooks/useCampaigns";
import Search from "../../components/Input/Search";
import useProjects from "../../hooks/useDatabase/projects";
import { ProjectDropdown } from "../../hooks/useDatabase/projects/types";
import { useClickOutside } from "./useClickOutside";
import { SelectionProvider, useSelection } from "./context/SelectionContext";

function WorkspaceSelector() {
  const navigate = useNavigate();
  const auth = useAuth();

  useEffect(() => {
    if (!auth?.loading && !auth?.isAuthenticated) {
      navigate("/login");
    }
  }, [auth?.isAuthenticated, auth?.loading, navigate]);

  const {
    getCustomers,
    customers,
    total,
    processing,
    error: customerError,
  } = useCustomers();

  const {
    getProjectsDropdown,
    projectsDropdown,
    processing: processingProjects,
    error: projectError,
  } = useProjects();

  const {
    getCampaigns,
    campaigns,
    total: totalCampaings,
    processing: processingCampaigns,
    error: campaignError,
  } = useCampaigns();

  const { customer, setCustomer, project, setProject, campaign, setCampaign } =
    useSelection();

  const [isEntering, _] = useState(false);

  const [queryCustomer, setQueryCustomer] = useState<string>("");
  const [queryProject, setQueryProject] = useState<string>("");
  const [queryCampaign, setQueryCampaign] = useState<string>("");

  const [suggestions, setSuggestions] = useState<Entity[]>([]);
  const [showCustomerSuggestions, setShowCustomerSuggestions] =
    useState<boolean>(false);

  const [projectSuggestions, setProjectSuggestions] = useState<
    ProjectDropdown[]
  >([]);
  const [showProjectSuggestions, setShowProjectSuggestions] =
    useState<boolean>(false);

  const [campaignSuggestions, setCampaignSuggestions] = useState<Entity[]>([]);
  const [showCampaignSuggestions, setShowCampaignSuggestions] =
    useState<boolean>(false);

  const [highlightedCustomerIndex, setHighlightedCustomerIndex] = useState(0);
  const [highlightedProjectIndex, setHighlightedProjectIndex] = useState(0);
  const [highlightedCampaignIndex, setHighlightedCampaignIndex] = useState(0);

  useEffect(() => {
    setCustomer(customer);
    setQueryCustomer(customer?.name || "");
    setProject(project);
    setQueryProject(project?.name || "");
    setCampaign(campaign);
    setQueryCampaign(campaign?.name || "");
  }, [customer, project, campaign]);

  useEffect(() => {
    getCustomers("limit=1000");
  }, []);

  useEffect(() => {
    setSuggestions(customers);
  }, [customers]);

  useEffect(() => {
    if (!customer) return;
    getProjectsDropdown(customer.id);
  }, [customer, getProjectsDropdown]);

  useEffect(() => {
    setProjectSuggestions(projectsDropdown);
  }, [projectsDropdown]);

  useEffect(() => {
    if (!customer || !project) return;
    getCampaigns(
      `customer_id=${customer.id}&project_name=${project.name}&limit=100`
    );
  }, [customer, project, getCampaigns]);

  useEffect(() => {
    setCampaignSuggestions(campaigns);
  }, [campaigns]);

  const customerRef = useRef<HTMLDivElement>(null);
  const projectRef = useRef<HTMLDivElement>(null);
  const campaignRef = useRef<HTMLDivElement>(null);

  useClickOutside(customerRef, () => setShowCustomerSuggestions(false));
  useClickOutside(projectRef, () => setShowProjectSuggestions(false));
  useClickOutside(campaignRef, () => setShowCampaignSuggestions(false));

  const handleCampaignChange = async (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = e.target.value.toLowerCase();
    setQueryCampaign(value);
    setCampaign(undefined);

    if (totalCampaings <= 100 && campaigns.length > 0) {
      const filtered =
        value.trim() === ""
          ? campaigns
          : campaigns.filter((campaign) =>
              campaign.name.toLowerCase().includes(value.toLowerCase())
            );

      setCampaignSuggestions(filtered);
    } else {
      await getCampaigns("name=" + value);
    }

    setShowCampaignSuggestions(true);
    setHighlightedCampaignIndex(0);
  };

  const handleCustomerChange = async (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = e.target.value.toLowerCase();
    setQueryCustomer(value);
    setCustomer(undefined);

    if (total <= 100 && customers.length > 0) {
      const filtered =
        value.trim() === ""
          ? customers
          : customers.filter((client) =>
              client.name.toLowerCase().includes(value.toLowerCase())
            );

      setSuggestions(filtered);
    } else {
      await getCustomers("name=" + value);
    }
    setShowCustomerSuggestions(true);
    setHighlightedCustomerIndex(0);
  };

  const handleProjectChange = async (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = e.target.value.toLowerCase();
    setQueryProject(value);
    setProject(undefined);

    if (total <= 100 && projectsDropdown.length > 0) {
      const filtered =
        value.trim() === ""
          ? projectsDropdown
          : projectsDropdown.filter((project) =>
              project.name.toLowerCase().includes(value.toLowerCase())
            );

      setProjectSuggestions(filtered);
    } else {
      await getProjectsDropdown(customer.id);
    }
    setShowProjectSuggestions(true);
    setHighlightedProjectIndex(0);
  };

  const handleCustomerSuggestionClick = (customer: Entity) => {
    setQueryCustomer(customer.name);
    setCustomer(customer);
    setProject(undefined);
    setCampaign(undefined);
    setShowCustomerSuggestions(false);
  };

  const handleProjectSuggestionClick = (project: ProjectDropdown) => {
    setQueryProject(project.name);
    setProject({ id: project.id, name: project.name });
    setCampaign(undefined);
    setShowProjectSuggestions(false);
  };

  const handleCampaignSuggestionClick = (campaign: Entity) => {
    setQueryCampaign(campaign.name);
    setCampaign(campaign);
    setShowCampaignSuggestions(false);
  };

  const createHandleKeyDown =
    (
      suggestions: any[],
      highlightedIndex: number,
      setHighlightedIndex: React.Dispatch<React.SetStateAction<number>>,
      onSelect: (item: any) => void
    ) =>
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (suggestions.length === 0) return;

      if (e.key === "ArrowDown") {
        e.preventDefault();
        setHighlightedIndex((prev) =>
          prev < suggestions.length - 1 ? prev + 1 : 0
        );
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setHighlightedIndex((prev) =>
          prev > 0 ? prev - 1 : suggestions.length - 1
        );
      } else if (e.key === "Enter") {
        e.preventDefault();
        onSelect(suggestions[highlightedIndex]);
      }
    };

  const handleSelection = async (e: React.FormEvent) => {
    e.preventDefault();
    if (customer) setCustomer(customer);
    if (project) setProject(project);
    if (campaign) setCampaign(campaign);

    navigate("/admin/dashboard");
  };

  return (
    <div className="flex flex-col md:flex-row h-screen">
      <Cover />
      <div className="w-full md:w-2/5 flex-1 flex items-center justify-center bg-gray-50 p-4">
        <div className="w-full max-w-sm">
          <h2 className="text-4xl font-bold text-center mb-3">
            Hola {auth.user?.Username}!
          </h2>
          <p className="text-center text-base text-gray-900 mb-6">
            Para comenzar debes elegir un cliente, proyecto y campaña.
          </p>
          <div className="ml-8 w-full max-w-xs">
            {customerError ||
              projectError ||
              (campaignError && (
                <div
                  className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
                  role="alert"
                >
                  <span className="font-medium">Error!</span> {customerError}{" "}
                  {projectError} {campaignError}
                </div>
              ))}
            <form className="space-y-4" onSubmit={handleSelection}>
              <div ref={customerRef} className="relative">
                <Search
                  label="Cliente"
                  placeholder="Ingrese nombre"
                  name="client"
                  value={queryCustomer}
                  onClick={() => {
                    if (!showCustomerSuggestions) {
                      setShowCustomerSuggestions(true);
                    }
                  }}
                  onChange={handleCustomerChange}
                  onFocus={() => setShowCustomerSuggestions(true)}
                  onKeyDown={createHandleKeyDown(
                    suggestions,
                    highlightedCustomerIndex,
                    setHighlightedCustomerIndex,
                    handleCustomerSuggestionClick
                  )}
                  className={"w-full"}
                  fullWidth
                  disabled={processing}
                />
                {suggestions.length === 0 && (
                  <div className="px-4 py-2 text-red-600">
                    No hay clientes activos para operar. Presione el siguientebotón para agregar un nuevo cliente.
                    <button
                      type="button"
                      onClick={() => navigate("/admin/database/customers")}
                      className={`mt-2 text-white font-semibold text-base w-full rounded-xl px-[26px] py-[12px] transition duration-300 bg-custom-btn hover:bg-custom-btn/80 cursor-pointer`}
                    >
                    Crear nuevo cliente
                  </button>
                  </div>
                )}
                {showCustomerSuggestions && (
                  <div className="flex justify-between items-center">
                    <ul className="absolute top-full mb-1 w-full bg-white border rounded-lg shadow-md z-10 max-h-[200px] overflow-y-auto">
                      {suggestions.length > 0 &&
                        suggestions.map((customer, index) => (
                          <li
                            key={index}
                            onClick={() =>
                              handleCustomerSuggestionClick(customer)
                            }
                            className={`px-4 py-2 cursor-pointer ${
                              index === highlightedCustomerIndex
                                ? "bg-gray-300 font-medium"
                                : "hover:bg-gray-300 hover:font-medium"
                            }`}
                          >
                            {customer.name}
                          </li>
                        ))}
                    </ul>
                  </div>
                )}
              </div>
              {suggestions.length > 0 && ( <>
              <div ref={projectRef} className="relative">
                <Search
                  label="Proyecto"
                  placeholder="Seleccione proyecto"
                  name="project"
                  disabled={!customer || processingProjects}
                  value={queryProject}
                  onClick={() => {
                    if (!showProjectSuggestions) {
                      setShowProjectSuggestions(true);
                    }
                  }}
                  onChange={handleProjectChange}
                  onFocus={() => setShowProjectSuggestions(true)}
                  onKeyDown={createHandleKeyDown(
                    projectSuggestions,
                    highlightedProjectIndex,
                    setHighlightedProjectIndex,
                    handleProjectSuggestionClick
                  )}
                  className={"w-full"}
                  fullWidth
                />
                {showProjectSuggestions && (
                  <div className="flex justify-between items-center">
                    <ul className="absolute top-full mb-1 w-full bg-white border rounded-lg shadow-md z-10 max-h-[200px] overflow-y-auto">
                      {projectSuggestions.length > 0 &&
                        projectSuggestions.map((project, index) => (
                          <li
                            key={index}
                            onClick={() =>
                              handleProjectSuggestionClick(project)
                            }
                            className={`px-4 py-2 cursor-pointer ${
                              index === highlightedProjectIndex
                                ? "bg-gray-300 font-medium"
                                : "hover:bg-gray-300 hover:font-medium"
                            }`}
                          >
                            {project.name}
                          </li>
                        ))}
                    </ul>
                  </div>
                )}
                {customer &&
                  !processingProjects &&
                  projectsDropdown.length === 0 && (
                    <div className="px-4 py-2 text-red-600">
                      No hay proyectos disponibles.
                    </div>
                  )}
              </div>
              <div ref={campaignRef} className="relative">
                <Search
                  label="Campaña"
                  placeholder="Seleccione campaña"
                  name="campaing"
                  disabled={!customer || !project || processingCampaigns}
                  value={queryCampaign}
                  onClick={() => {
                    if (!showCampaignSuggestions) {
                      setShowCampaignSuggestions(true);
                    }
                  }}
                  onChange={handleCampaignChange}
                  onFocus={() => setShowCampaignSuggestions(true)}
                  onKeyDown={createHandleKeyDown(
                    campaignSuggestions,
                    highlightedCampaignIndex,
                    setHighlightedCampaignIndex,
                    handleCampaignSuggestionClick
                  )}
                  className={"w-full"}
                  fullWidth
                />
                {showCampaignSuggestions && (
                  <div className="flex justify-between items-center">
                    <ul className="absolute top-full left-0 w-full bg-white border rounded-lg shadow-md z-10">
                      {campaigns.length > 0 &&
                        campaignSuggestions.map((campaign, index) => (
                          <li
                            key={index}
                            onClick={() =>
                              handleCampaignSuggestionClick(campaign)
                            }
                            className="px-4 py-2 cursor-pointer hover:bg-gray-300"
                          >
                            {campaign.name}
                          </li>
                        ))}
                    </ul>
                  </div>
                )}
              </div>
              <button
                type="submit"
                disabled={!project || !campaign || !customer || isEntering}
                className={`text-white font-semibold text-base w-full rounded-xl px-[26px] py-[16px] transition duration-300
                  ${
                    !project || !campaign || !customer || isEntering
                      ? "bg-gray-500 cursor-not-allowed"
                      : "bg-custom-btn hover:bg-custom-btn/80 cursor-pointer"
                  }`}
              >
                {isEntering ? (
                  <>
                    <svg
                      aria-hidden="true"
                      role="status"
                      className="inline w-4 h-4 me-3 text-white animate-spin"
                      viewBox="0 0 100 101"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <path
                        d="M100 50.5908C100 78.2051 77.6142 100.591 50 100.591C22.3858 100.591 0 78.2051 0 50.5908C0 22.9766 22.3858 0.59082 50 0.59082C77.6142 0.59082 100 22.9766 100 50.5908ZM9.08144 50.5908C9.08144 73.1895 27.4013 91.5094 50 91.5094C72.5987 91.5094 90.9186 73.1895 90.9186 50.5908C90.9186 27.9921 72.5987 9.67226 50 9.67226C27.4013 9.67226 9.08144 27.9921 9.08144 50.5908Z"
                        fill="#E5E7EB"
                      />
                      <path
                        d="M93.9676 39.0409C96.393 38.4038 97.8624 35.9116 97.0079 33.5539C95.2932 28.8227 92.871 24.3692 89.8167 20.348C85.8452 15.1192 80.8826 10.7238 75.2124 7.41289C69.5422 4.10194 63.2754 1.94025 56.7698 1.05124C51.7666 0.367541 46.6976 0.446843 41.7345 1.27873C39.2613 1.69328 37.813 4.19778 38.4501 6.62326C39.0873 9.04874 41.5694 10.4717 44.0505 10.1071C47.8511 9.54855 51.7191 9.52689 55.5402 10.0491C60.8642 10.7766 65.9928 12.5457 70.6331 15.2552C75.2735 17.9648 79.3347 21.5619 82.5849 25.841C84.9175 28.9121 86.7997 32.2913 88.1811 35.8758C89.083 38.2158 91.5421 39.6781 93.9676 39.0409Z"
                        fill="currentColor"
                      />
                    </svg>
                    Loading...{" "}
                  </>
                ) : (
                  "Ingresar"
                )}
              </button>
              </>
              )}
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}

const WorkspaceSelectorPage = () => (
  <AuthProvider>
    <SelectionProvider>
      <WorkspaceSelector />
    </SelectionProvider>
  </AuthProvider>
);

export default WorkspaceSelectorPage;
