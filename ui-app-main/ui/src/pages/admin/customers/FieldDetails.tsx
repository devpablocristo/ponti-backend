import { useState } from "react";
import { cropColors } from "../colors";

type FieldData = {
  name: string;
  hectares: string;
  crops: any[];
  lease_type: string;
};

type Props = {
  field: FieldData;
};

export default function FieldDetails({ field }: Props) {
  const [showModal, setShowModal] = useState(false);

  const items = [
    { label: "fields", value: field.name },
    { label: "hectares", value: String(field.hectares) + " Hectáreas" },
    { label: "crops", value: field.crops },
    { label: "leasetype", value: field.lease_type },
  ];

  return (
    <div className="flex flex-wrap gap-4">
      {items.map(({ label, value }) => (
        <div
          key={label}
          className="bg-gray-100 rounded-xl p-4 min-w-[180px] flex-1 my-1"
        >
          <div className="font-medium text-[14px] leading-[150%] text-black">
            {label === "crops" && Array.isArray(value) ? (
              <>
                <div className="flex flex-wrap gap-2">
                  {value
                    .slice(0, 4)
                    .map((crop: { id: number; name: string }) => (
                      <span
                        key={crop.id}
                        className={`px-2 py-1 text-[14px] rounded-md ${
                          cropColors[crop.name] || "bg-[#E5E7EB] text-[#000000] border border-[#000000]"
                        }`}
                      >
                        {crop.name}
                      </span>
                    ))}
                  {value.length > 4 && (
                    <button
                      className="ml-2 text-blue-600 underline text-sm"
                      onClick={() => setShowModal(true)}
                      type="button"
                    >
                      Ver todos
                    </button>
                  )}
                </div>
                {showModal && (
                  <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-40 z-50">
                    <div className="bg-white p-6 rounded-lg shadow-lg max-w-md w-full">
                      <h2 className="text-lg font-bold mb-4">
                        Todos los cultivos
                      </h2>
                      <div className="flex flex-wrap gap-2 mb-4">
                        {value.map((crop: { id: number; name: string }) => (
                          <span
                            key={crop.id}
                            className={`px-4 py-1 rounded-xl text-base font-medium ${
                              cropColors[crop.id] ||
                              "bg-[#E5E7EB] text-[#000000] border border-[#000000]"
                            }`}
                          >
                            {crop.name}
                          </span>
                        ))}
                      </div>
                      <button
                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                        onClick={() => setShowModal(false)}
                        type="button"
                      >
                        Cerrar
                      </button>
                    </div>
                  </div>
                )}
              </>
            ) : (
              value
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
