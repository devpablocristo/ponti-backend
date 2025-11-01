import { useEffect, useRef, useState } from "react";
import Button from "../../../components/Button/Button";
import InputField from "../../../components/Input/InputField";
import SelectField from "../../../components/Input/SelectField";
import useProducts from "../../../hooks/useProducts";
import { LoaderCircle, Trash } from "lucide-react";
import useProjects from "../../../hooks/useDatabase/projects";
import { Entity } from "../../../hooks/useDatabase/options/types";
import Search from "../../../components/Input/Search";
import { useClickOutside } from "../../login/useClickOutside";
import { useKeyboardNavigation } from "../database/customers/hooks/useKeyboardNavigation";
import useProviders from "../../../hooks/useProviders";
import useSupplyMovements from "../../../hooks/useSupplyMovement";
import {
  Campaign,
  Customer,
  Project,
} from "../../../hooks/useWorkspaceFilters";
import useCampaigns from "../../../hooks/useCampaigns";

const emptyItems = [
  {
    item: "",
    quantity: "",
  },
  {
    item: "",
    quantity: "",
  },
  {
    item: "",
    quantity: "",
  },
  {
    item: "",
    quantity: "",
  },
];

const typeOptions = [
  { id: 1, name: "Stock" },
  { id: 2, name: "Movimiento interno" },
  { id: 3, name: "Remito oficial" },
];

function Drawer({
  open,
  onClose,
  children,
}: {
  open: boolean;
  onClose: () => void;
  children: React.ReactNode;
}) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-40 flex">
      <div
        className="fixed inset-0 bg-black bg-opacity-30 transition-opacity"
        onClick={onClose}
      />
      <div className="ml-auto h-full w-full max-w-3xl bg-white shadow-xl p-6 overflow-y-auto relative animate-slide-in-right">
        <button
          className="absolute top-4 right-4 text-gray-400 hover:text-gray-600"
          onClick={onClose}
          aria-label="Cerrar"
        >
          <svg
            width="24"
            height="24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
        {children}
      </div>
      <style>{`
        @keyframes slide-in-right {
          from { transform: translateX(100%); opacity: 0; }
          to { transform: translateX(0); opacity: 1; }
        }
        .animate-slide-in-right {
          animation: slide-in-right 0.25s cubic-bezier(0.4,0,0.2,1);
        }
      `}</style>
    </div>
  );
}

export default function CreateItem({
  drawerOpen,
  setDrawerOpen,
  projectId,
  customers,
  onProductCreated,
}: {
  drawerOpen: boolean;
  setDrawerOpen: (open: boolean) => void;
  projectId: number;
  customers: Customer[];
  onProductCreated: () => void;
}) {
  const {
    resultCreation,
    errorCreation,
    processingCreation,
    saveSupplyMovement,
  } = useSupplyMovements();
  const { getProject, selectedProject, processing } = useProjects();
  const { getProviders, providers } = useProviders();

  const [error, setError] = useState<string | null>(null);
  const [errorMessages, setErrorMessages] = useState<string[]>([]);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const [customer, setCustomer] = useState<Customer | null>(null);
  const [campaign, setCampaign] = useState<Campaign | null>(null);
  const [project, setProject] = useState<Project | null>(null);

  const [selectedProjectDestination, setSelectedProjectDestination] = useState<
    number | null
  >(null);

  const { getSupplies, supplies } = useProducts();
  const { projectsDropdown, getProjectsDropdown } = useProjects();
  const { campaigns, getCampaigns } = useCampaigns();

  const [orderNumber, setOrderNumber] = useState("");
  const [date, setDate] = useState("");
  const [investor, setInvestor] = useState<{ id: number; name: string } | null>(
    null
  );
  const [investors, setInvestors] = useState<{ id: number; name: string }[]>(
    []
  );

  const [queryProvider, setQueryProvider] = useState<string>("");
  const [provider, setProvider] = useState<Entity>();
  const [providerSuggestions, setProviderSuggestions] = useState<Entity[]>([]);
  const [showProviderSuggestions, setShowProviderSuggestions] =
    useState<boolean>(false);

  const [type, setType] = useState<{ id: number; name: string } | null>(null);

  const [items, setItems] = useState<
    {
      item: string;
      quantity: string;
    }[]
  >(emptyItems);

  const clearForm = () => {
    setError(null);
    setErrorMessages([]);
    setProvider(undefined);
    setQueryProvider("");
    setShowProviderSuggestions(false);
    setInvestor(null);
    setItems(emptyItems);
    setOrderNumber("");
    setDate("");
  };

  useEffect(() => {
    setSuccessMessage(null);
    setError(null);
    setErrorMessages([]);
  }, [drawerOpen]);

  useEffect(() => {
    getProviders("");
  }, []);

  useEffect(() => {
    if (!customer) return;
    getProjectsDropdown(customer.id);
  }, [customer]);

  useEffect(() => {
    if (customer && project) {
      getCampaigns(
        `customer_id=${customer.id}&project_name=${project.name}&limit=100`
      );
    }
  }, [customer, project]);

  useEffect(() => {
    if (providers) {
      setProviderSuggestions(providers);
    }
  }, [providers]);

  useEffect(() => {
    if (errorCreation) {
      setError(errorCreation);
      setSuccessMessage(null);
    }
  }, [errorCreation]);

  useEffect(() => {
    if (resultCreation.supply_movements.length > 0) {
      const errors: string[] = [];
      resultCreation.supply_movements.forEach((movement) => {
        if (movement.error_detail !== "") {
          errors.push(movement.error_detail.replace("VALIDATION_ERROR: ", ""));
        }
      });

      if (errors.length > 0) {
        setError(errors.join("\n"));
        setSuccessMessage(null);
        return;
      }

      setSuccessMessage("Movimiento guardado correctamente");
      onProductCreated();
      clearForm();
    }
  }, [resultCreation]);

  useEffect(() => {
    if (projectId) {
      getSupplies(projectId);
      getProject(projectId);
    }
  }, [projectId]);

  useEffect(() => {
    if (!selectedProject) return;
    setInvestors(
      selectedProject.investors
        .filter((i) => i.id !== null)
        .map((i) => ({ id: i.id!, name: i.name }))
    );
  }, [selectedProject]);

  const handleProviderSuggestionClick = (provider: Entity) => {
    setQueryProvider(provider.name);
    setProvider(provider);
    setShowProviderSuggestions(false);
  };

  const {
    highlightedIndex: highlightedProviderIndex,
    handleKeyDown: handleProviderKeyDown,
    setHighlightedIndex: setProviderHighlightedIndex,
  } = useKeyboardNavigation({
    suggestions: providerSuggestions,
    showSuggestions: showProviderSuggestions,
    onSelect: handleProviderSuggestionClick,
    onEscape: () => setShowProviderSuggestions(false),
  });

  const providerRef = useRef<HTMLDivElement>(null);
  useClickOutside(providerRef, () => setShowProviderSuggestions(false));

  const handleProviderChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQueryProvider(value);
    setProvider(undefined);

    if (providers) {
      const filtered =
        value.trim() === ""
          ? providers
          : providers.filter((provider) =>
              provider.name.toLowerCase().includes(value.toLowerCase())
            );

      setProviderSuggestions(filtered);
      setShowProviderSuggestions(true);
      setProviderHighlightedIndex(0);
    }
  };

  const handleItemChange = (i: number, field: string, value: string) => {
    setItems((prev) =>
      prev.map((item, idx) => (idx === i ? { ...item, [field]: value } : item))
    );
  };

  const handlePreSave = () => {
    const errors: string[] = [];
    setErrorMessages(errors);

    if (!provider) {
      if (queryProvider === "") {
        errors.push("Debe seleccionar un proveedor.");
      } else {
        setProvider({
          id: 0,
          name: queryProvider,
        });
      }
    }

    if (!investor) {
      errors.push("Debe seleccionar un inversor.");
    }

    if (!orderNumber) {
      errors.push("Debe seleccionar un número de orden.");
    }

    if (!date) {
      errors.push("Debe seleccionar una fecha.");
    }

    if (!type) {
      errors.push("Debe seleccionar un tipo.");
    }

    if (type && type.id === 2 && !selectedProjectDestination) {
      errors.push("Debe seleccionar un proyecto destino.");
    }

    const itemsWithAnyValue = items.filter(
      (item) => item.item || item.quantity
    );

    if (itemsWithAnyValue.length === 0) {
      errors.push("Debe cargar al menos un insumo");
      return;
    }

    const hasPartial = itemsWithAnyValue.some(
      (item) => !item.item || !item.quantity
    );

    if (hasPartial) {
      errors.push("No se completaron todos los campos de los items cargados");
      return;
    }

    if (errors.length > 0) {
      setErrorMessages(errors);
      return;
    }

    saveSupplyMovement(projectId, {
      items: itemsWithAnyValue.map((item) => ({
        supply_id: Number(item.item),
        quantity: Number(item.quantity),
        movement_type: type?.name || "",
        movement_date: new Date(date),
        reference_number: orderNumber,
        project_destination_id: selectedProjectDestination || 0,
        investor_id: investor?.id || 0,
        provider: {
          id: provider?.id || 0,
          name: provider ? provider.name : queryProvider,
        },
      })),
    });
  };

  return (
    <Drawer open={drawerOpen} onClose={() => setDrawerOpen(false)}>
      <div className="flex flex-col h-full">
        <h2 className="text-lg font-semibold mb-2">Ingreso de insumo</h2>
        {processing || processingCreation ? (
          <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
            <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
          </div>
        ) : (
          <>
            <form className="space-y-4 flex-1">
              <div className="grid grid-cols-3 gap-4">
                <SelectField
                  label="Tipo de ingreso"
                  name="type"
                  options={typeOptions}
                  value={type?.id?.toString() || ""}
                  onChange={(e) => {
                    const selectedType = typeOptions.find(
                      (type) => type.id === Number(e.target.value)
                    );
                    if (selectedType) {
                      setType(selectedType);
                    }
                  }}
                  disabled={processing}
                  size="sm"
                />
                <InputField
                  label="Fecha"
                  name="date"
                  type="date"
                  value={date || ""}
                  onChange={(e) => {
                    const inputValue = e.target.value;
                    if (inputValue) {
                      const dateParts = inputValue.split("-");
                      if (dateParts[0] && dateParts[0].length > 4) {
                        dateParts[0] = dateParts[0].slice(0, 4);
                        const formattedDate = dateParts.join("-");
                        setDate(formattedDate);
                      } else {
                        setDate(inputValue);
                      }
                    } else {
                      setDate("");
                    }
                  }}
                  size="sm"
                />
                <InputField
                  label="Numero / Nombre"
                  placeholder="Numero / Nombre"
                  name="nroName"
                  type="text"
                  value={orderNumber || ""}
                  onChange={(e) => setOrderNumber(e.target.value)}
                  size="sm"
                />
              </div>

              <div className="grid grid-cols-3 gap-4">
                <InputField
                  label="Proyecto"
                  name="project"
                  type="text"
                  value={selectedProject?.name || ""}
                  onChange={() => {}}
                  disabled
                  size="sm"
                />

                <div ref={providerRef} className="relative">
                  <Search
                    label="Proveedor"
                    placeholder="Ingrese nombre o fecha"
                    name="provider"
                    value={queryProvider}
                    onClick={() => {
                      if (!showProviderSuggestions) {
                        setShowProviderSuggestions(true);
                      }
                    }}
                    onChange={handleProviderChange}
                    onFocus={() => setShowProviderSuggestions(true)}
                    onKeyDown={handleProviderKeyDown}
                    className={"w-full"}
                    size="sm"
                    fullWidth
                  />
                  {showProviderSuggestions && (
                    <div className="flex justify-between items-center">
                      <ul className="absolute top-full mb-1 w-full bg-white border rounded-lg shadow-md z-10 max-h-[200px] overflow-y-auto">
                        {providerSuggestions.length > 0 &&
                          providerSuggestions.map((provider, index) => (
                            <li
                              key={index}
                              onClick={() =>
                                handleProviderSuggestionClick(provider)
                              }
                              className={`px-4 py-2 cursor-pointer ${
                                index === highlightedProviderIndex
                                  ? "bg-gray-300 font-medium"
                                  : "hover:bg-gray-300 hover:font-medium"
                              }`}
                            >
                              {provider.name}
                            </li>
                          ))}
                      </ul>
                    </div>
                  )}
                </div>
                <SelectField
                  label="Inversor"
                  placeholder="Selecciona el inversor"
                  name="investor"
                  options={investors}
                  value={investor?.id?.toString() || ""}
                  onChange={(e) => {
                    const selectedInvestor = investors.find(
                      (i) => i.id === Number(e.target.value)
                    );
                    if (selectedInvestor) {
                      setInvestor(selectedInvestor);
                    }
                  }}
                  size="sm"
                />
              </div>

              {type?.id === 2 && (
                <div className="grid grid-cols-3 gap-4">
                  <SelectField
                    label="Cliente destino"
                    name="customer"
                    options={customers}
                    value={customer?.id?.toString() || ""}
                    onChange={(e) => {
                      const selectedCustomer = customers.find(
                        (customer) => customer.id === Number(e.target.value)
                      );
                      if (selectedCustomer) {
                        setCustomer(selectedCustomer);
                      }
                    }}
                    size="sm"
                  />
                  <SelectField
                    label="Proyecto destino"
                    name="projectDestination"
                    options={projectsDropdown}
                    value={project?.id?.toString() || ""}
                    onChange={(e) => {
                      const selectedProject = projectsDropdown.find(
                        (project) => project.id === Number(e.target.value)
                      );
                      if (selectedProject) {
                        setProject(selectedProject);
                      }
                    }}
                    disabled={processing || !customers}
                    size="sm"
                  />
                  <SelectField
                    label="Campaña"
                    name="campaign"
                    options={campaigns}
                    value={campaign?.id?.toString() || ""}
                    onChange={(e) => {
                      const selectedCampaign = campaigns.find(
                        (campaign) => campaign.id === Number(e.target.value)
                      );
                      if (selectedCampaign) {
                        setCampaign(selectedCampaign);
                        setSelectedProjectDestination(
                          selectedCampaign.project_id
                        );
                      }
                    }}
                    size="sm"
                  />
                </div>
              )}
              <div>
                <div className="hidden sm:grid grid-cols-[1.5fr_1fr_1.5fr] gap-4 mb-2">
                  <span className="font-sm text-gray-900">Insumo</span>
                  <span className="font-sm text-gray-900">Cantidad</span>
                  <div></div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-[1.5fr_1fr_1.5fr] gap-4">
                  {items.map((item, i) => (
                    <div
                      key={i}
                      className="sm:contents border sm:border-0 p-4 sm:p-0 rounded-md sm:rounded-none mb-4 sm:mb-0 shadow-sm sm:shadow-none"
                    >
                      <div className="sm:col-span-1">
                        <SelectField
                          label=""
                          name={`item-${i}`}
                          options={supplies}
                          value={item.item}
                          onChange={(e) =>
                            handleItemChange(i, "item", e.target.value)
                          }
                          size="sm"
                        />
                      </div>
                      <div className="sm:col-span-1">
                        <InputField
                          label=""
                          placeholder="Lts/kg"
                          name={`quantity${i}`}
                          type="text"
                          value={item.quantity}
                          onChange={(e) => {
                            let value = e.target.value.replace(/,/g, ".");
                            if (/^\d*\.?\d{0,3}$/.test(value)) {
                              handleItemChange(i, "quantity", value);
                            }
                          }}
                          size="sm"
                        />
                      </div>
                      <div>
                        <Button
                          variant="danger"
                          size="xs"
                          onClick={() => {
                            const newItems = [...items];
                            newItems.splice(i, 1);
                            setItems(newItems);
                          }}
                          className="text-blue-500 hover:underline max-w-fit"
                        >
                          <Trash size={12} />
                        </Button>
                      </div>
                    </div>
                  ))}
                  <Button
                    variant="outlinePonti"
                    size="sm"
                    onClick={() => {
                      setItems([...items, { item: "", quantity: "" }]);
                    }}
                    className="max-w-fit"
                  >
                    Agregar insumo +
                  </Button>
                </div>
              </div>
              {errorMessages.length > 0 && (
                <div
                  id="alert-2"
                  className="flex items-center p-4 mb-4 text-red-800 rounded-lg bg-red-50"
                  role="alert"
                >
                  <div>
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
              {error && error !== "" && (
                <div
                  className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
                  role="alert"
                >
                  <span className="font-medium">Error!</span> {error}
                  <button
                    type="button"
                    className="ms-auto -mx-1 -my-1 bg-red-50 text-red-500 rounded-lg focus:ring-2 focus:ring-red-400 p-1.5 hover:bg-red-200 inline-flex items-center justify-center h-8 w-8 dark:bg-gray-800 dark:text-red-400 dark:hover:bg-gray-700"
                    aria-label="Close"
                    onClick={() => setError("")}
                  >
                    <span className="sr-only">Close</span>
                    <svg
                      className="w-2 h-2"
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
            </form>
            <div className="flex justify-end gap-2 mt-auto pt-6 pb-2 bg-white">
              <div className="flex gap-2">
                <Button
                  variant="outlineGray"
                  className="text-base font-medium"
                  onClick={() => setDrawerOpen(false)}
                >
                  Cancelar
                </Button>
                <Button
                  variant="success"
                  className="text-base font-medium"
                  onClick={handlePreSave}
                  disabled={processing || processingCreation}
                >
                  Guardar
                </Button>
              </div>
            </div>
          </>
        )}
      </div>
    </Drawer>
  );
}
