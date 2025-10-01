import { useEffect, useState } from "react";
import Header from "../../../components/Header/Header";
import DataTable from "../../../components/Table/DataTable";
import useUsers from "../../../hooks/useUsers";
import LoadingScreen from "../../../components/LoadingScreen/LoadingScreen";
import { UserData } from "../../../hooks/useUsers/types";
import { useNavigate } from "react-router-dom";
import { BaseModal } from "../../../components/Modal/BaseModal";

type Column<T> = {
  key: keyof T;
  header: string;
  render?: (value: T[keyof T]) => React.ReactNode;
};

const columns: Column<UserData>[] = [
  { key: "id", header: "ID" },
  {
    key: "username",
    header: "Username",
    render: (value) => <strong>{value as string}</strong>,
  },
  { key: "created_at", header: "Created At" },
  { key: "updated_at", header: "Updated At" },
  {
    key: "created_by",
    header: "Created By",
  },
];

export function Users() {
  const navigate = useNavigate();

  const [errorMessage, setErrorMessage] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<UserData | null>(null);
  const [modalMessage, setModalMessage] = useState<string | null>(null);
  const [isDeleteFinished, setIsDeleteFinished] = useState(false);

  const {
    getUsers,
    users,
    processing,
    error,
    deleteUser,
    deleteError,
    isDeleting,
  } = useUsers();

  useEffect(() => {
    setErrorMessage("");
    getUsers("");
  }, []);

  useEffect(() => {
    setErrorMessage(error);
  }, [error]);

  const handleEdit = async (p: UserData) => {
    navigate(`edit/${p.id}`);
  };

  const handleDelete = (user: UserData) => {
    setSelectedUser(user);
    setModalMessage(
      `¿Está seguro que desea eliminar al usuario "${user.username}"?`
    );
    setIsModalOpen(true);
    setIsDeleteFinished(false);
  };

  const confirmDelete = async () => {
    if (!selectedUser) return;

    setModalMessage("⏳ Eliminando usuario...");

    await deleteUser(selectedUser.id);

    if (deleteError) {
      setModalMessage(deleteError);
    } else {
      setModalMessage("✅ Usuario eliminado con éxito");
      getUsers("");
    }

    setIsDeleteFinished(true);
  };

  if (processing)
    return <LoadingScreen title={["Cargando..."]} description={[""]} />;

  return (
    <div>
      <Header
        title=""
        actionButtons={[
          {
            label: "Add User",
            path: "/admin/users/new",
          },
        ]}
      />
      <div className="mt-4">
        {users.length > 0 ? (
          <DataTable
            data={users}
            columns={columns}
            onEdit={(item) => handleEdit(item)}
            onDelete={(item) => handleDelete(item)}
          />
        ) : errorMessage === "" ? (
          <p className="text-gray-500 text-center mt-4">
            No hay usuarios disponibles.
          </p>
        ) : (
          ""
        )}
        {errorMessage && (
          <div
            className="p-4 mb-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
            role="alert"
          >
            <span className="font-medium">Error!</span> {errorMessage}
          </div>
        )}
        <BaseModal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          title="Eliminar usuario"
          message={modalMessage || "¿Está seguro?"}
          primaryButtonText={
            isDeleteFinished
              ? null
              : isDeleting
              ? "Eliminando..."
              : "Sí, eliminar"
          }
          secondaryButtonText="Cerrar"
          onPrimaryAction={isDeleteFinished ? undefined : confirmDelete}
        />
      </div>
    </div>
  );
}
