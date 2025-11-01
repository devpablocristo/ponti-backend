const FormButtons: React.FC<{
  onCancel: () => void;
  isSubmitting: boolean;
  submitLabel: string;
}> = ({ onCancel, isSubmitting, submitLabel }) => (
  <div className="flex justify-between mt-6">
    <button
      type="button"
      onClick={onCancel}
      className="bg-gray-200 text-gray-900 font-medium py-3 px-6 rounded-lg hover:bg-gray-300"
    >
      Volver
    </button>
    <button
      type="submit"
      disabled={isSubmitting}
      className="bg-indigo-600 text-white font-medium py-3 px-6 rounded-lg hover:bg-indigo-700"
    >
      {isSubmitting ? "Guardando..." : submitLabel}
    </button>
  </div>
);

export default FormButtons;
