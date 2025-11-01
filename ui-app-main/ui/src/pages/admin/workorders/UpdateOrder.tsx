import { useEffect, useState } from "react";
import Button from "../../../components/Button/Button";
import InputField from "../../../components/Input/InputField";
import SelectField from "../../../components/Input/SelectField";
import useTasks from "../../../hooks/useTasks";
import { TaskInfo } from "../../../hooks/useTasks/types";
import useWorkOrders from "../../../hooks/useWorkOrders";
import { LoaderCircle } from "lucide-react";
import useProjects from "../../../hooks/useDatabase/projects";
import { Plot } from "../../../hooks/useDatabase/projects/types";
import { WorkorderData } from "../../../hooks/useWorkOrders/types";
import useProducts from "../../../hooks/useProducts";

const emptyItems = [
  {
    item: "",
    totalUsed: "",
    dose: "",
  },
  {
    item: "",
    totalUsed: "",
    dose: "",
  },
  {
    item: "",
    totalUsed: "",
    dose: "",
  },
  {
    item: "",
    totalUsed: "",
    dose: "",
  },
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
      {/* Fondo oscuro para cerrar */}
      <div
        className="fixed inset-0 bg-black bg-opacity-30 transition-opacity"
        onClick={onClose}
      />
      {/* Drawer */}
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

export default function UpdateOrder({
  orderId,
  drawerOpen,
  setDrawerOpen,
  onOrderUpdated,
  onOrderDuplicated,
}: {
  orderId: number;
  drawerOpen: boolean;
  setDrawerOpen: (open: boolean) => void;
  onOrderUpdated: () => void;
  onOrderDuplicated: (order: WorkorderData) => void;
}) {
  const {
    updateOrder,
    getWorkorder,
    selectedOrder,
    resultCreation,
    errorCreation,
    processingCreation,
  } = useWorkOrders();

  const { getProject, selectedProject, processing } = useProjects();

  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const { getSupplies, supplies } = useProducts();
  const { getLabors, labors } = useTasks();
  const [lots, setLots] = useState<Plot[]>([]);
  const [lot, setLot] = useState<Plot | null>(null);
  const [labor, setLabor] = useState<TaskInfo | null>(null);
  const [contractor, setContractor] = useState("");
  const [observations, setObservations] = useState("");
  const [orderNumber, setOrderNumber] = useState("");
  const [surface, setSurface] = useState("");
  const [date, setDate] = useState("");
  const [investor, setInvestor] = useState<{ id: number; name: string } | null>(
    null
  );
  const [investors, setInvestors] = useState<{ id: number; name: string }[]>(
    []
  );

  const [items, setItems] = useState<
    {
      item: string;
      totalUsed: string;
      dose: string;
    }[]
  >(emptyItems);

  useEffect(() => {
    if (orderId) {
      getWorkorder(orderId);
    }
  }, [orderId]);

  useEffect(() => {
    if (selectedOrder) {
      getProject(selectedOrder.project_id);
      getSupplies(selectedOrder.project_id);
      getLabors(selectedOrder.project_id);
    }
  }, [selectedOrder]);

  useEffect(() => {
    if (!selectedProject || !selectedOrder) return;
    setInvestors(
      selectedProject.investors
        .filter((i) => i.id !== null)
        .map((i) => ({ id: i.id!, name: i.name }))
    );

    const foundField = selectedProject.fields.find(
      (f) => String(f.id) === String(selectedOrder.field_id)
    );

    if (foundField?.lots) {
      setLots(foundField.lots);
    } else {
      setLots([]);
    }
  }, [selectedProject, selectedOrder]);

  useEffect(() => {
    if (
      selectedOrder &&
      investors.length > 0 &&
      labors.length > 0 &&
      lots.length > 0
    ) {
      setOrderNumber(selectedOrder.number);
      setSurface(selectedOrder.effective_area.toString());

      const isoDate = selectedOrder.date; // "2025-08-06T00:00:00Z"
      const formattedDate = isoDate.split("T")[0]; // "2025-08-06"
      setDate(formattedDate);

      const investorObj = investors.find(
        (i) => i.id === selectedOrder.investor_id
      );
      setInvestor(investorObj || null);

      const laborObj = labors.find((l) => l.id === selectedOrder.labor_id);
      setLabor(laborObj || null);

      const lotObj = lots.find((l) => l.id === selectedOrder.lot_id);
      setLot(lotObj || null);

      setContractor(selectedOrder.contractor);
      setObservations(selectedOrder.observations);

      let loadedItems = selectedOrder.items.map((item) => ({
        item: item.supply_id.toString(),
        totalUsed: item.total_used.toString(),
        dose: item.final_dose.toString(),
      }));

      while (loadedItems.length < 4) {
        loadedItems.push({ item: "", totalUsed: "", dose: "" });
      }

      setItems(loadedItems);
    }
  }, [selectedOrder, investors, labors, lots]);

  useEffect(() => {
    if (errorCreation) {
      setError(errorCreation);
      setSuccessMessage(null);
    }
  }, [errorCreation]);

  useEffect(() => {
    if (resultCreation) {
      setSuccessMessage(resultCreation);
      onOrderUpdated();
    }
  }, [resultCreation]);

  useEffect(() => {
    setSuccessMessage(null);
  }, [drawerOpen]);

  useEffect(() => {
    if (surface && surface !== "" && surface !== "0") {
      items.forEach((item, i) => {
        if (item.totalUsed && item.totalUsed !== "") {
          handleItemChange(
            i,
            "dose",
            (Number(item.totalUsed) / Number(surface)).toFixed(3)
          );
        }
      });
    }
  }, [surface]);

  const handleItemChange = (i: number, field: string, value: string) => {
    setItems((prev) =>
      prev.map((item, idx) => (idx === i ? { ...item, [field]: value } : item))
    );
  };

  const handleSaveOrder = () => {
    setError(null);
    setSuccessMessage(null);
    if (
      !selectedOrder ||
      !lot ||
      !labor ||
      !contractor ||
      !investor ||
      !surface ||
      !orderNumber ||
      !date ||
      processing
    ) {
      setError("Campos obligatorios incompletos");
      return;
    }

    const itemsWithAnyValue = items.filter(
      (item) => item.item || item.totalUsed || item.dose
    );

    if (itemsWithAnyValue.length > 0) {
      const hasPartial = itemsWithAnyValue.some(
        (item) => !item.item || !item.totalUsed || !item.dose
      );

      if (hasPartial) {
        setError("No se completaron todos los campos de los items cargados");
        return;
      }
    }

    updateOrder(orderId, {
      number: orderNumber,
      date,
      project_id: selectedOrder.project_id,
      field_id: selectedOrder.field_id,
      lot_id: lot.id,
      crop_id: lot.current_crop_id,
      labor_id: labor.id,
      contractor,
      investor_id: investor.id,
      effective_area: Number(surface),
      items: itemsWithAnyValue.map((item) => ({
        supply_id: Number(item.item),
        total_used: Number(item.totalUsed),
        final_dose: Number(item.dose),
      })),
      observations,
    });
  };

  return (
    <Drawer open={drawerOpen} onClose={() => setDrawerOpen(false)}>
      <div className="flex flex-col h-full">
        <h2 className="text-lg font-semibold mb-2">
          Edición de Orden de Trabajo:{" "}
          <span className="text-gray-700">{selectedProject?.name}</span>
        </h2>
        {processing || processingCreation ? (
          <div className="absolute inset-0 bg-white bg-opacity-70 backdrop-blur-sm flex items-center justify-center z-10">
            <LoaderCircle className="w-10 h-10 text-blue-600 animate-spin" />
          </div>
        ) : (
          <>
            <form className="space-y-4 flex-1">
              <div className="grid grid-cols-4 gap-4">
                <InputField
                  label="Nro. Orden"
                  placeholder="000-001"
                  name="order"
                  type="text"
                  value={orderNumber || ""}
                  onChange={(e) => setOrderNumber(e.target.value)}
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
              </div>

              <div className="grid grid-cols-3 gap-4">
                <InputField
                  label="Campo"
                  name="field"
                  type="text"
                  value={
                    selectedProject?.fields.find(
                      (f) => f.id === selectedOrder?.field_id
                    )?.name || ""
                  }
                  onChange={() => {}}
                  disabled
                  size="sm"
                />
                <SelectField
                  label="Lote"
                  name="lot"
                  options={lots.map((lot) => ({
                    id: lot.id,
                    name: lot.name,
                  }))}
                  value={lot?.id?.toString() || ""}
                  onChange={(e) => {
                    const selectedLot = lots.find(
                      (l) => l.id === Number(e.target.value)
                    );
                    if (selectedLot) {
                      setLot(selectedLot);
                    }
                  }}
                  disabled={!selectedOrder || processing}
                  size="sm"
                />

                <div>
                  <InputField
                    label="Cultivo actual"
                    placeholder="Selecciona el lote"
                    name="crop"
                    type="text"
                    value={lot?.current_crop_name || ""}
                    onChange={() => {}}
                    disabled
                    size="sm"
                  />
                </div>
              </div>

              <div className="grid grid-cols-4 gap-4">
                <div>
                  <SelectField
                    label="Labor"
                    placeholder="Selecciona el labor"
                    name="labor"
                    options={labors}
                    value={labor?.id?.toString() || ""}
                    onChange={(e) => {
                      const selectedLabor = labors.find(
                        (l) => l.id === Number(e.target.value)
                      );
                      if (selectedLabor) {
                        setLabor(selectedLabor);
                        setContractor(selectedLabor.contractor_name);
                      }
                    }}
                    size="sm"
                  />
                </div>
                <div>
                  <InputField
                    label="Superficie realizada"
                    placeholder="Ingresar superficie"
                    name="surface"
                    type="text"
                    value={surface}
                    onChange={(e) => {
                      let value = e.target.value.replace(/,/g, ".");
                      if (/^\d*\.?\d{0,2}$/.test(value)) {
                        setSurface(value);
                      }
                    }}
                    size="sm"
                  />
                </div>
                <InputField
                  label="Contratista"
                  placeholder="Selecciona el labor"
                  name="contractor"
                  type="text"
                  value={contractor}
                  disabled
                  onChange={() => {}}
                  size="sm"
                />
                <SelectField
                  label="Inversor del labor"
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

              {/* Tabla de insumos */}
              <div>
                <div className="hidden sm:grid grid-cols-[1.5fr_1fr_1fr_0.5fr] gap-4 mb-2">
                  <span className="font-sm text-gray-900">Insumo</span>
                  <span className="font-sm text-gray-900">Total utilizado</span>
                  <span className="font-sm text-gray-900">Dosis final</span>
                  <div></div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-[1.5fr_1fr_1fr_0.5fr] gap-4">
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
                          placeholder="lts/kg"
                          name={`totalUsed${i}`}
                          type="text"
                          value={item.totalUsed}
                          onChange={(e) => {
                            let value = e.target.value.replace(/,/g, ".");
                            if (/^\d*\.?\d{0,3}$/.test(value)) {
                              handleItemChange(i, "totalUsed", value);
                              if (
                                surface &&
                                surface !== "" &&
                                surface !== "0"
                              ) {
                                handleItemChange(
                                  i,
                                  "dose",
                                  (Number(value) / Number(surface)).toFixed(3)
                                );
                              }
                            }
                          }}
                          size="sm"
                        />
                      </div>
                      <div className="sm:col-span-1">
                        <InputField
                          label=""
                          placeholder="Total/superficie"
                          name={`dose${i}`}
                          type="text"
                          value={item.dose}
                          onChange={(e) => {
                            let value = e.target.value.replace(/,/g, ".");
                            if (/^\d*\.?\d{0,3}$/.test(value)) {
                              handleItemChange(i, "dose", value);
                            }
                          }}
                          size="sm"
                        />
                      </div>
                      <div>
                        <Button
                          variant="outlineGray"
                          size="xs"
                          onClick={() => {
                            const newItems = [...items];
                            newItems.splice(i, 1);
                            setItems(newItems);
                          }}
                          className="text-blue-500 hover:underline max-w-fit"
                        >
                          Eliminar
                        </Button>
                      </div>
                    </div>
                  ))}
                  <Button
                    variant="primary"
                    size="xs"
                    onClick={() => {
                      setItems([
                        ...items,
                        { item: "", totalUsed: "", dose: "" },
                      ]);
                    }}
                    className="text-blue-500 hover:underline max-w-fit"
                  >
                    + Agregar nuevo insumo
                  </Button>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Observaciones
                </label>
                <textarea
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 min-h-[80px]"
                  placeholder="Escriba observaciones"
                  name="observaciones"
                  value={observations}
                  onChange={(e) => setObservations(e.target.value)}
                />
              </div>
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
            <div className="flex justify-between gap-2 mt-auto pt-6 pb-2 bg-white">
              {selectedOrder && (
                <Button
                  onClick={() => onOrderDuplicated(selectedOrder)}
                  variant="outlineGray"
                  className="text-base font-medium"
                >
                  Duplicar orden
                </Button>
              )}
              <div className="flex gap-2">
                <Button
                  variant="outlineGray"
                  className="text-base font-medium"
                  onClick={() => setDrawerOpen(false)}
                >
                  Cancelar
                </Button>
                <Button
                  onClick={handleSaveOrder}
                  variant="success"
                  className="text-base font-medium"
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
