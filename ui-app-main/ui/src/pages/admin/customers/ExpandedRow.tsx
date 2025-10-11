import { useEffect, useState } from "react";
import FieldDetails from "./FieldDetails";
import useProjects from "../../../hooks/useDatabase/projects";
import { Field } from "../../../hooks/useDatabase/projects/types";

function mapFieldToFieldData(field: Field) {
  const totalHectares = field.lots.reduce(
    (sum, plot) => sum + (plot.hectares || 0),
    0
  );

  const allCrops = [
    ...field.lots
      .map((plot) => ({
        id: plot.current_crop_id,
        name: plot.current_crop_name,
      }))
      .filter(Boolean),
  ];

  const cropMap = new Map();
  allCrops.forEach((crop) => {
    if (crop && crop.id != null) {
      cropMap.set(crop.id, crop);
    }
  });
  const crops = Array.from(cropMap.values());

  let leaseTypeParts = field.lease_type_name || "";
  leaseTypeParts += " (";
  if (field.lease_type_value > 0) {
    leaseTypeParts += `${field.lease_type_value}USD. `;
  }

  if (field.lease_type_percent > 0) {
    leaseTypeParts += `${field.lease_type_percent}%`;
  }

  leaseTypeParts += ")";

  return {
    name: field.name,
    hectares: totalHectares.toLocaleString("es-AR", {
      maximumFractionDigits: 2,
    }),
    crops,
    lease_type: leaseTypeParts,
  };
}

const ExpandedRow = ({ projectId }: { projectId: number }) => {
  const [loading, setLoading] = useState(true);

  const { getProject, selectedProject } = useProjects();

  useEffect(() => {
    setLoading(true);
    getProject(projectId);
  }, [projectId, getProject]);

  useEffect(() => {
    if (selectedProject) {
      setLoading(false);
    }
  }, [selectedProject]);

  if (loading) return <div>Cargando detalles...</div>;
  if (
    !selectedProject ||
    !selectedProject.fields ||
    selectedProject.fields.length === 0
  )
    return <div>No field data available for this project.</div>;

  const items = [
    { label: "CAMPOS" },
    { label: "HECTÁREAS" },
    { label: "CULTIVOS" },
    { label: "TIPO DE ARRIENDO" },
  ];

  return (
    <div>
      <div className="flex flex-wrap gap-4">
        {items.map(({ label }) => (
          <div
            key={label}
            className="rounded-xl pt-1 pl-4 min-w-[180px] flex-1"
          >
            <div className="text-xs font-semibold text-gray-600 tracking-wide uppercase mb-1">
              {label}
            </div>
          </div>
        ))}
      </div>
      {selectedProject.fields.map((fd, index) => (
        <FieldDetails key={index} field={mapFieldToFieldData(fd)} />
      ))}
    </div>
  );
};

export default ExpandedRow;
