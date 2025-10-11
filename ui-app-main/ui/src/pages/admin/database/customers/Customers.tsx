import { useEffect, useRef, useState } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { LoaderCircle } from "lucide-react";

import InputField from "../../../../components/Input/InputField";
import Button from "../../../../components/Button/Button";
import { BaseModal } from "../../../../components/Modal/BaseModal";
import AutocompleteSelect from "./AutocompleteSelect";
import Fields, { Field } from "./Fields";
import useOptions from "../../../../hooks/useDatabase/options";
import { Entity, Investor } from "../../../../hooks/useDatabase/options/types";
import useProjects from "../../../../hooks/useDatabase/projects";
import { Project } from "../../../../hooks/useDatabase/projects/types";
import Search from "../../../../components/Input/Search";
import { useKeyboardNavigation } from "./hooks/useKeyboardNavigation";
import { useClickOutside } from "./hooks/useClickOutside";
import { useSelection } from "../../../login/context/SelectionContext";

export default function Customers() {
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const duplicateId = searchParams.get("id");

  const [projectName, setProjectName] = useState("");

  const [suggestions, setSuggestions] = useState<Entity[]>([]);
  const [showCustomerSuggestions, setShowCustomerSuggestions] =
    useState<boolean>(false);
  const [queryCustomer, setQueryCustomer] = useState<string>("");
  const [customer, setCustomer] = useState<Entity>();

  const [queryCampaign, setQueryCampaign] = useState<string>("");
  const [campaign, setCampaign] = useState<Entity>();
  const [campaignSuggestions, setCampaignSuggestions] = useState<Entity[]>([]);
  const [showCampaingSuggestions, setShowCampaignSuggestions] =
    useState<boolean>(false);

  const [projectManagers, setProjectManagers] = useState<Entity[]>([]);

  const [queryManager, setQueryManager] = useState<string>("");

  const [investors, setInvestors] = useState<Investor[]>([]);
  const [investor, setInvestor] = useState<Investor>();

  const [adminCostInvestors, setAdminCostInvestors] = useState<Investor[]>([]);
  const [adminCostInvestor, setAdminCostInvestor] = useState<Investor>();

  const [adminCostInvestment, setAdminCostInvestment] = useState<number | "">("");
  const [investment, setInvestment] = useState<number | "">("");
  const [investmentError, setInvestmentError] = useState<string>("");
  const [adminCostInvestmentError, setAdminCostInvestmentError] = useState<string>("");

  const [queryInvestor, setQueryInvestor] = useState<string>("");
  const [queryAdminCostInvestor, setQueryAdminCostInvestor] = useState<string>("");

  const [admincost, setAdmincost] = useState<string>("");

  const [modalOpen, setModalOpen] = useState(false);
  const [modalAdminOpen, setModalAdminOpen] = useState(false);

  const emptyField: Field = {
    id: 0,
    name: "",
    leaseType: "",
    leaseTypePercent: 0,
    leaseTypeValue: 0,
    plots: [
      {
        id: 0,
        name: "",
        hectares: 0,
        previousCrop: { id: 0, name: "" },
        currentCrop: { id: 0, name: "" },
        season: "",
      },
    ],
  };

  const [fields, setFields] = useState<Field[]>([emptyField]);

  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessages, setErrorMessages] = useState<string[]>([]);
  const { getOptions, options, processing, error } = useOptions();
  const {
    saveProject,
    updateProject,
    getProject,
    deleteProject,
    selectedProject,
    error: projectError,
    result,
  } = useProjects();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalConfig, setModalConfig] = useState({
    title: "",
    message: "",
    primaryButtonText: "",
    secondaryButtonText: "Cancelar",
    onConfirm: () => {},
  });

  const [pendingPayload, setPendingPayload] = useState<Project | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const { seasons } = useSelection();

  useEffect(() => {
    getOptions();
  }, []);

  useEffect(() => {
    if (options) {
      setSuggestions(options.clients);
      setCampaignSuggestions(options.campaigns);
    }
  }, [options]);

  useEffect(() => {
    const projectId = id || duplicateId;
    if (projectId) {
      getProject(Number(projectId));
    }
  }, [id, duplicateId]);

  const convertFields = (
    fields: {
      id: number;
      name: string;
      lease_type_id: number;
      lease_type_percent: number;
      lease_type_value: number;
      lots: {
        id: number;
        name: string;
        hectares: number;
        previous_crop_id: number;
        current_crop_id: number;
        previous_crop_name?: string;
        current_crop_name?: string;
        season: string;
      }[];
    }[]
  ): Field[] => {
    return fields.map((field) => ({
      id: field.id,
      name: field.name,
      leaseType: String(field.lease_type_id),
      leaseTypePercent: field.lease_type_percent,
      leaseTypeValue: field.lease_type_value,
      plots: field.lots.map((lot) => ({
        id: lot.id,
        name: lot.name,
        hectares: lot.hectares,
        previousCrop: {
          id: lot.previous_crop_id,
          name: lot.previous_crop_name || "",
        },
        currentCrop: {
          id: lot.current_crop_id,
          name: lot.current_crop_name || "",
        },
        season: lot.season,
      })),
    }));
  };

  function convertToEntities(
    dataList: { id: number | null; name: string }[]
  ): Entity[] {
    return dataList
      .filter((item): item is { id: number; name: string } => item.id !== null)
      .map((item) => ({
        id: item.id,
        name: item.name,
      }));
  }

  function convertToInvestorEntities(
    dataList: { id: number | null; name: string; percentage: number }[]
  ): Investor[] {
    return dataList
      .filter(
        (item): item is { id: number; name: string; percentage: number } =>
          item.id !== null
      )
      .map((item) => ({
        id: item.id,
        name: item.name,
        percentage: item.percentage,
      }));
  }

  useEffect(() => {
    if (selectedProject && options) {
      setProjectName(selectedProject.name);
      setFields(convertFields(selectedProject.fields || []));
      setQueryCustomer(selectedProject.customer.name);

      if (id) {
        setQueryCampaign(selectedProject.campaign.name);
        setCampaign({
          id: selectedProject.campaign.id || 0,
          name: selectedProject.campaign.name,
        });
      }

      setQueryInvestor("");
      setQueryManager("");
      setAdmincost(String(selectedProject.admin_cost));
      setCustomer({
        id: selectedProject.customer.id || 0,
        name: selectedProject.customer.name,
      });
      setProjectManagers(convertToEntities(selectedProject.managers));
      setInvestors(convertToInvestorEntities(selectedProject.investors));
    }
  }, [selectedProject, options]);

  useEffect(() => {
    if (error && error.trim() !== "") {
      setErrorMessages((prev) =>
        prev.includes(error) ? prev : [...prev, error]
      );
    }
  }, [error]);

  useEffect(() => {
    if (projectError && projectError.trim() !== "") {
      setErrorMessages((prev) =>
        prev.includes(projectError) ? prev : [...prev, projectError]
      );
    }
  }, [projectError]);

  const cleanForm = () => {
    setProjectName("");
    setFields([emptyField]);
    setQueryCustomer("");
    setQueryInvestor("");
    setQueryManager("");
    setQueryCampaign("");
    setAdmincost("");
    setCustomer(undefined);
    setCampaign(undefined);
    setProjectManagers([]);
    setInvestors([]);
    setAdminCostInvestors([]);
  };

  useEffect(() => {
    if (result !== "") {
      if (!id) {
        cleanForm();
      }
      setTimeout(() => {
        document
          .getElementById("main-scroll")
          ?.scrollTo({ top: 0, behavior: "smooth" });
      }, 0);
    }
    setErrorMessages([]);
    setSuccessMessage(result);
  }, [result, id]);

  useEffect(() => {
    if (pendingPayload) {
      setModalConfig({
        title: "Confirmar guardado",
        message: "¿Está seguro que desea guardar los cambios?",
        primaryButtonText: "Aceptar",
        secondaryButtonText: "Cancelar",
        onConfirm: handleSaveConfirmed,
      });
      setIsModalOpen(true);
    }
  }, [pendingPayload]);

  const inputRef = useRef<HTMLInputElement>(null);
  useEffect(() => {
    if (modalOpen) {
      setInvestment("");
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [modalOpen]);

  const inputAdminCostRef = useRef<HTMLInputElement>(null);
  useEffect(() => {
    if (modalAdminOpen) {
      setAdminCostInvestment("");
      setTimeout(() => inputAdminCostRef.current?.focus(), 0);
    }
  }, [modalAdminOpen]);

  const wrapperRef = useRef<HTMLDivElement>(null);
  const campaignRef = useRef<HTMLDivElement>(null);

  useClickOutside(wrapperRef, () => setShowCustomerSuggestions(false));
  useClickOutside(campaignRef, () => setShowCampaignSuggestions(false));

  const handleCustomerChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQueryCustomer(value);
    setCustomer(undefined);

    if (options?.clients) {
      const filtered =
        value.trim() === ""
          ? options.clients
          : options.clients.filter((client) =>
              client.name.toLowerCase().includes(value.toLowerCase())
            );

      setSuggestions(filtered);
      setShowCustomerSuggestions(true);
      setHighlightedIndex(0);
    }
  };

  const handleCampaingChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQueryCampaign(value);
    setCampaign(undefined);

    if (options?.campaigns) {
      const filtered =
        value.trim() === ""
          ? options.campaigns
          : options.campaigns.filter((campaign) =>
              campaign.name.toLowerCase().includes(value.toLowerCase())
            );

      setCampaignSuggestions(filtered);
      setShowCampaignSuggestions(true);
      setCampaignHighlightedIndex(0);
    }
  };

  const handleCustomerSuggestionClick = (customer: Entity) => {
    setQueryCustomer(customer.name);
    setCustomer(customer);
    setShowCustomerSuggestions(false);
  };

  const handleCampaignSuggestionClick = (campaing: Entity) => {
    setQueryCampaign(campaing.name);
    setCampaign(campaing);
    setShowCampaignSuggestions(false);
  };

  const handleManagerSuggestionClick = (manager: Entity) => {
    if (!projectManagers.find((m) => m.name === manager.name)) {
      setProjectManagers([...projectManagers, manager]);
    }
  };

  const handleInvestorSuggestionClick = (investor: Investor) => {
    setInvestor(investor);
    setInvestmentError("");
    setModalOpen(true);
  };

  const handleAdminCostInvestorSuggestionClick = (investor: Investor) => {
    setAdminCostInvestor(investor);
    setAdminCostInvestmentError("");
    setModalAdminOpen(true);
  };

  const { highlightedIndex, handleKeyDown, setHighlightedIndex } =
    useKeyboardNavigation({
      suggestions,
      showSuggestions: showCustomerSuggestions,
      onSelect: handleCustomerSuggestionClick,
      onEscape: () => setShowCustomerSuggestions(false),
    });

  const {
    highlightedIndex: highlightedCampaignIndex,
    handleKeyDown: handleCampaignKeyDown,
    setHighlightedIndex: setCampaignHighlightedIndex,
  } = useKeyboardNavigation({
    suggestions: campaignSuggestions,
    showSuggestions: showCampaingSuggestions,
    onSelect: handleCampaignSuggestionClick,
    onEscape: () => setShowCampaignSuggestions(false),
  });

  const handleAdminCostChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const rawValue = e.target.value;

    if (!/^\d*\.?\d*$/.test(rawValue)) return;

    const parts = rawValue.split(".");
    if (parts.length === 2 && parts[1].length > 2) return;

    setAdmincost(rawValue);
  };

  const handleSaveInvestment = () => {
    if (!investor || investment === null || investment === undefined) {
      return;
    }

    if (
      typeof investment === "number" &&
      investment >= 1 &&
      investment <= 100
    ) {
      let sum = 0;
      for (const inv of investors) {
        sum += inv.percentage || 0;
      }
      if (sum + investment > 100) {
        setInvestmentError(
          "La suma total de los porcentajes de inversores no puede superar 100."
        );
        return;
      }
      const updatedInvestor: Investor = {
        ...investor,
        percentage: Number(investment),
      };

      setInvestors([...investors, updatedInvestor]);
    } else {
      setInvestmentError("Porcentaje debe ser entre 1 y 100");
      return;
    }

    setModalOpen(false);
    setInvestment("");
    setInvestor(undefined);
  };

  const handleSaveAdminCostInvestment = () => {
    if (!adminCostInvestor || adminCostInvestment === null || adminCostInvestment === undefined) {
      return;
    }

    if (
      typeof adminCostInvestment === "number" &&
      adminCostInvestment >= 1 &&
      adminCostInvestment <= 100
    ) {
      let sum = 0;
      for (const inv of adminCostInvestors) {
        sum += inv.percentage || 0;
      }
      if (sum + adminCostInvestment > 100) {
        setAdminCostInvestmentError(
          "La suma total de los porcentajes de inversores no puede superar 100."
        );
        return;
      }
      const updatedInvestor: Investor = {
        ...adminCostInvestor,
        percentage: Number(adminCostInvestment),
      };

      setAdminCostInvestors([...adminCostInvestors, updatedInvestor]);
    } else {
      setAdminCostInvestmentError("Porcentaje debe ser entre 1 y 100");
      return;
    }

    setModalAdminOpen(false);
    setAdminCostInvestment("");
    setAdminCostInvestor(undefined);
  };

  const handlePreSave = () => {
    const errors: string[] = [];
    setErrorMessages(errors);

    if (!customer) {
      if (queryCustomer === "") {
        errors.push("Debe seleccionar un cliente.");
      } else {
        setCustomer({
          id: 0,
          name: queryCustomer,
        });
      }
    }

    if (!projectName || projectName.trim() === "") {
      errors.push("Debe ingresar un nombre de proyecto.");
    }

    if (!campaign) {
      if (queryCampaign === "") {
        errors.push("Debe seleccionar o cargar una campaña.");
      } else {
        setCampaign({
          id: 0,
          name: queryCampaign,
        });
      }
    }

    if (!projectManagers.length) {
      errors.push("Debe agregar al menos un responsable.");
    }

    if (admincost === "") {
      errors.push("Debe agregar el costo administrativo.");
    }

    if (!investors.length) {
      errors.push("Debe agregar al menos un inversor.");
    }

    const totalPercentage = investors.reduce(
      (sum, inv) => sum + Number(inv.percentage || 0),
      0
    );
    const invalidPercentages = investors.some(
      (inv) => inv.percentage <= 0 || inv.percentage > 100
    );

    if (invalidPercentages) {
      errors.push(
        "Cada porcentaje de inversor debe ser mayor a 0 y menor o igual a 100."
      );
    }

    if (totalPercentage > 100) {
      errors.push(
        "La suma total de los porcentajes de inversores no puede superar 100."
      );
    }

    if (totalPercentage < 100) {
      errors.push(
        "La suma total de los porcentajes de inversores debe ser 100."
      );
    }

    if (!fields.length) {
      errors.push("Debe agregar al menos un campo.");
    }

    for (const field of fields) {
      if (!field.name || !field.leaseType) {
        errors.push("Cada campo debe tener nombre y tipo de arriendo.");
        break;
      }

      if (!field.plots.length) {
        errors.push(`El campo "${field.name}" debe tener al menos un lote.`);
      }

      for (const plot of field.plots) {
        if (
          !plot.name ||
          plot.hectares <= 0 ||
          !plot.previousCrop ||
          !plot.currentCrop ||
          !plot.season
        ) {
          errors.push(
            `Todos los lotes del campo "${field.name}" deben estar completos y válidos.`
          );
        }
      }
    }

    if (errors.length) {
      setErrorMessages(errors);
      setSuccessMessage("");
      return;
    }

    const payload = {
      name: projectName,
      updated_at: selectedProject?.updated_at,
      customer: {
        id: customer ? customer.id : 0,
        name: customer ? customer.name : queryCustomer,
      },
      campaign: {
        id: campaign ? campaign.id : 0,
        name: campaign ? campaign.name : queryCampaign,
      },
      admin_cost: parseFloat(admincost),
      managers: projectManagers.map((pm) => ({
        id: pm.id || 0,
        name: pm.name,
      })),
      investors: investors.map((inv) => ({
        id: inv.id || 0,
        name: inv.name,
        percentage: Number(inv.percentage) || 0,
      })),
      admin_cost_investors: adminCostInvestors.map((inv) => ({
        id: inv.id || 0,
        name: inv.name,
        percentage: Number(inv.percentage) || 0,
      })),
      fields: fields.map((field) => ({
        id: id ? field.id || 0 : 0,
        name: field.name,
        lease_type_id: Number(field.leaseType),
        lease_type_percent: field.leaseTypePercent,
        lease_type_value: field.leaseTypeValue,
        lots: field.plots.map((plot) => ({
          id: id ? plot.id || 0 : 0,
          name: plot.name,
          hectares: Number(plot.hectares),
          previous_crop_id: Number(plot.previousCrop.id),
          previous_crop_name: plot.previousCrop.name,
          current_crop_id: Number(plot.currentCrop.id),
          current_crop_name: plot.currentCrop.name,
          season: plot.season || "",
        })),
      })),
    };

    setPendingPayload(payload);
  };

  const handlePreCancel = () => {
    setModalConfig({
      title: "Confirmar cancelación",
      message:
        "¿Está seguro que desea cancelar? Se perderán los cambios no guardados.",
      primaryButtonText: "Sí, cancelar",
      secondaryButtonText: "Volver",
      onConfirm: handleCancelConfirmed,
    });
    setIsModalOpen(true);
  };

  const handleCancelConfirmed = () => {
    if (id || duplicateId) {
      window.location.href = "/admin/customers";
    } else {
      cleanForm();
    }
  };

  const handlePreFinish = () => {
    setModalConfig({
      title: "Confirmar finalización",
      message: "¿Está seguro que desea finalizar la campaña?",
      primaryButtonText: "Sí, finalizar",
      secondaryButtonText: "Volver",
      onConfirm: handleFinishConfirmed,
    });
    setIsModalOpen(true);
  };

  const handleFinishConfirmed = async () => {
    if (!id) return;
    setIsSaving(true);

    try {
      await deleteProject(Number(id));
    } catch (error) {
      console.error("Error al finalizar:", error);
    } finally {
      setModalConfig({
        title: "Confirmar finalización",
        message: "La campaña ha sido finalizada.",
        primaryButtonText: "Volver",
        secondaryButtonText: "Volver",
        onConfirm: () => {
          window.location.href = "/admin/customers";
        },
      });
      setIsModalOpen(true);
      setIsSaving(false);
      cleanForm();
    }
  };

  const handleSaveConfirmed = async () => {
    if (!pendingPayload) return;

    setIsSaving(true);

    try {
      if (!id) {
        await saveProject(pendingPayload);
      } else {
        await updateProject(Number(id), pendingPayload);
      }
    } catch (error) {
      console.error("Error al guardar:", error);
    } finally {
      setIsSaving(false);
      setIsModalOpen(false);
      setPendingPayload(null);
    }
  };

  return (
    <div>
      {errorMessages.length > 0 && (
        <div
          id="alert-2"
          className="flex items-center p-4 mb-4 text-red-800 rounded-lg bg-red-50"
          role="alert"
        >
          <div>
            <div className="ms-3 text-sm font-medium">
              {!id
                ? "Error en la creación de un cliente o sociedad!"
                : "Error en la edición de un cliente o sociedad!"}
            </div>
            <ul className="mt-1.5 list-disc list-inside">
              {errorMessages.map((msg, index) => (
                <li key={index}>{msg}</li>
              ))}
            </ul>
          </div>
          <button
            type="button"
            className="ms-auto -mx-1.5 -my-1.5 bg-red-50 text-red-500 rounded-lg focus:ring-2 focus:ring-red-400 p-1.5 hover:bg-red-200 inline-flex items-center justify-center h-8 w-8 dark:bg-gray-800 dark:text-red-400 dark:hover:bg-gray-700"
            data-dismiss-target="#alert-2"
            aria-label="Close"
            onClick={() => setErrorMessages([])}
          >
            <span className="sr-only">Close</span>
            <svg
              className="w-3 h-3"
              aria-hidden="true"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 14 14"
            >
              <path
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
              />
            </svg>
          </button>
        </div>
      )}
      {successMessage && successMessage !== "" && (
        <div
          className="flex items-center p-4 mb-4 text-sm text-green-800 rounded-lg bg-green-50 dark:bg-gray-800 dark:text-green-400"
          role="alert"
        >
          <svg
            className="shrink-0 inline w-4 h-4 me-3"
            aria-hidden="true"
            xmlns="http://www.w3.org/2000/svg"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z" />
          </svg>
          <span className="sr-only">Info</span>
          <div>
            <span className="font-medium">{successMessage}</span>
            <br /> Para modificar los datos del cliente o los datos relacionados
            a los campos pertenecientes al cliente, edítelos y guarde nuevamente
          </div>
          <button
            type="button"
            className="ms-auto -mx-1.5 -my-1.5 bg-green-50 text-green-500 rounded-lg focus:ring-2 focus:ring-green-400 p-1.5 hover:bg-green-200 inline-flex items-center justify-center h-8 w-8 dark:bg-gray-800 dark:text-green-400 dark:hover:bg-gray-700"
            data-dismiss-target="#alert-3"
            aria-label="Close"
            onClick={() => setSuccessMessage("")}
          >
            <span className="sr-only">Close</span>
            <svg
              className="w-3 h-3"
              aria-hidden="true"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 14 14"
            >
              <path
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
              />
            </svg>
          </button>
        </div>
      )}
      <h1 className="text-custom-text ml-2 font-semibold text-xl leading-none">
        {!id
          ? "Agregar un nuevo cliente o sociedad"
          : `Edición del proyecto: ${selectedProject?.name}`}
      </h1>
      <div className="grid grid-cols-1 md:grid-cols-[1fr_3fr] gap-2 mt-2">
        {processing || isSaving && (
          <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
            <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
          </div>
        )}
        <div className="pt-4 pl-1 pr-2 space-y-4">
          <div ref={wrapperRef} className="relative">
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
              onKeyDown={handleKeyDown}
              className={"w-full"}
              fullWidth
            />
            {showCustomerSuggestions && (
              <div className="flex justify-between items-center">
                <ul className="absolute top-full mb-1 w-full bg-white border rounded-lg shadow-md z-10 max-h-[200px] overflow-y-auto">
                  {suggestions.length > 0 &&
                    suggestions.map((customer, index) => (
                      <li
                        key={index}
                        onClick={() => handleCustomerSuggestionClick(customer)}
                        className={`px-4 py-2 cursor-pointer ${
                          index === highlightedIndex
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

          <div className="flex justify-between items-center">
            <InputField
              label="Proyecto"
              placeholder="Ingrese nombre"
              type="text"
              name="lote"
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              fullWidth
            />
          </div>

          <div ref={campaignRef} className="relative">
            <Search
              label="Campaña"
              placeholder="Ingrese nombre o fecha"
              name="campaign"
              value={queryCampaign}
              onClick={() => {
                if (!showCampaingSuggestions) {
                  setShowCampaignSuggestions(true);
                }
              }}
              onChange={handleCampaingChange}
              onFocus={() => setShowCampaignSuggestions(true)}
              onKeyDown={handleCampaignKeyDown}
              className={"w-full"}
              fullWidth
            />
            {showCampaingSuggestions && (
              <div className="flex justify-between items-center">
                <ul className="absolute top-full mb-1 w-full bg-white border rounded-lg shadow-md z-10 max-h-[200px] overflow-y-auto">
                  {campaignSuggestions.length > 0 &&
                    campaignSuggestions.map((campaign, index) => (
                      <li
                        key={index}
                        onClick={() => handleCampaignSuggestionClick(campaign)}
                        className={`px-4 py-2 cursor-pointer ${
                          index === highlightedCampaignIndex
                            ? "bg-gray-300 font-medium"
                            : "hover:bg-gray-300 hover:font-medium"
                        }`}
                      >
                        {campaign.name}
                      </li>
                    ))}
                </ul>
              </div>
            )}
          </div>

          <AutocompleteSelect<Entity>
            name="manager"
            label="Resp. del proyecto"
            placeholder="Ingrese nombre"
            options={options?.managers}
            selectedItems={projectManagers}
            query={queryManager}
            setQuery={setQueryManager}
            handleSuggestionClick={handleManagerSuggestionClick}
            setItems={setProjectManagers}
            customAddLabel="+ Agregar responsable"
            customAddItem={{ id: 0, name: queryManager }}
          />

          <InputField
            label="Costo administrativo"
            placeholder="$"
            name="costs"
            type="text"
            value={admincost}
            onChange={handleAdminCostChange}
            onPaste={(e) => {
              const paste = e.clipboardData.getData("text");
              if (!/^\d*\.?\d{0,2}$/.test(paste)) {
                e.preventDefault();
              }
            }}
          />

          <AutocompleteSelect<Investor>
            name="adminCostInvestor"
            label="Inversor y % en costo administrativo"
            placeholder="Ingrese nombre"
            options={options?.investors}
            selectedItems={adminCostInvestors}
            query={queryAdminCostInvestor}
            setQuery={setQueryAdminCostInvestor}
            handleSuggestionClick={handleAdminCostInvestorSuggestionClick}
            setItems={setAdminCostInvestors}
            customAddLabel="+ Agregar inversor"
            customAddItem={{ id: 0, name: queryAdminCostInvestor, percentage: 0 }}
            renderTag={(item) => `${item.name} (${item.percentage}%)`}
          />

          <AutocompleteSelect<Investor>
            name="investor"
            label="Inversor y % de participacion en proyecto"
            placeholder="Ingrese nombre"
            options={options?.investors}
            selectedItems={investors}
            query={queryInvestor}
            setQuery={setQueryInvestor}
            handleSuggestionClick={handleInvestorSuggestionClick}
            setItems={setInvestors}
            customAddLabel="+ Agregar inversor"
            customAddItem={{ id: 0, name: queryInvestor, percentage: 0 }}
            renderTag={(item) => `${item.name} (${item.percentage}%)`}
          />
        </div>
        <div>
          <Fields
            fields={fields}
            setFields={setFields}
            seasons={seasons}
            crops={options?.crops}
            rentTypes={options?.rentTypes}
          />
        </div>
      </div>
      <div className="border-b mt-2" style={{ borderColor: "#D1D5DB" }} />
      <div className="flex justify-between pt-6 mb-4">
        <div>
          {id && (
            <Button
              variant="danger"
              className="text-base font-medium"
              onClick={handlePreFinish}
              size="sm"
            >
              Finalizar proyecto
            </Button>
          )}
        </div>
        <div className="flex gap-4">
          <Button
            variant="outlineGray"
            className="text-base font-medium"
            onClick={handlePreCancel}
          >
            Cancelar
          </Button>
          <Button
            variant="success"
            onClick={handlePreSave}
            className="text-base font-medium"
          >
            {!id ? "Guardar" : "Editar"}
          </Button>
        </div>
      </div>
      <BaseModal
        isOpen={isModalOpen}
        isSaving={isSaving}
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
      <BaseModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        title={`Aporte para ${investor?.name || ""}`}
        primaryButtonText="Guardar"
        onPrimaryAction={handleSaveInvestment}
      >
        <div className="flex flex-col items-center gap-2">
          <p>Ingresá el porcentaje de aporte al proyecto:</p>
          <input
            ref={inputRef}
            type="number"
            min="1"
            max="100"
            value={investment}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                handleSaveInvestment();
              }
            }}
            onChange={(e) => {
              const val = Number(e.target.value);
              if (val > 100) return;
              setInvestment(e.target.value === "" ? "" : val);
            }}
            placeholder="Porcentaje"
            className="border rounded px-2 py-1 w-32 text-center"
          />
          <p className="text-sm text-red-700">{investmentError}</p>
        </div>
      </BaseModal>
      <BaseModal
        isOpen={modalAdminOpen}
        onClose={() => setModalAdminOpen(false)}
        title={`Aporte para ${investor?.name || ""}`}
        primaryButtonText="Guardar"
        onPrimaryAction={handleSaveAdminCostInvestment}
      >
        <div className="flex flex-col items-center gap-2">
          <p>Ingresá el porcentaje de aporte al costo administrativo:</p>
          <input
            ref={inputAdminCostRef}
            type="number"
            min="1"
            max="100"
            value={adminCostInvestment}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                handleSaveAdminCostInvestment();
              }
            }}
            onChange={(e) => {
              const val = Number(e.target.value);
              if (val > 100) return;
              setAdminCostInvestment(e.target.value === "" ? "" : val);
            }}
            placeholder="Porcentaje"
            className="border rounded px-2 py-1 w-32 text-center"
          />
          <p className="text-sm text-red-700">{adminCostInvestmentError}</p>
        </div>
      </BaseModal>
    </div>
  );
}
