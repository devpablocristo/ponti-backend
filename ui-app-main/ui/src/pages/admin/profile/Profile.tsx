import { useEffect, useState } from "react";

import Button from "../../../components/Button/Button";
import Header from "../../../components/Header/Header";
import useUsers from "../../../hooks/useUsers";
import { useAuth } from "../../login/context/AuthProvider";
import LoadingScreen from "../../../components/LoadingScreen/LoadingScreen";
import { BaseModal } from "../../../components/Modal/BaseModal";

export function Profile() {
  const auth = useAuth();

  const [errorMessage, setErrorMessage] = useState("");
  const { updateUser, getUser, processing, error, result, selectedUser } =
    useUsers();

  const [id, setId] = useState<number | null>(null);
  const [username, setUser] = useState("");
  const [password, setPassword] = useState("");
  const [passwordConfirm, setPasswordConfirm] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);

  useEffect(() => {
    if (auth?.user && auth?.isAuthenticated) {
      setId(auth.user.ID);
      getUser(auth.user.ID);
    }
  }, [auth?.user, auth?.isAuthenticated]);

  useEffect(() => {
    if (result) {
      setIsModalOpen(true);
    }
  }, [result]);

  useEffect(() => {
    if (selectedUser) {
      setUser(selectedUser.username);
      setPassword("");
      setPasswordConfirm("");
    }
  }, [selectedUser]);

  useEffect(() => {
    setErrorMessage(error);
  }, [error]);

  const updateProfileInfo = (e: React.FormEvent) => {
    setErrorMessage("");
    e.preventDefault();
    if (!username || !password || !passwordConfirm) {
      setErrorMessage("Todos los campos son obligatorios.");
      return;
    }

    if (password !== passwordConfirm) {
      setErrorMessage("Las contraseñas no coinciden.");
      return;
    }

    if (id === null) {
      setErrorMessage("ID inválido");
      return;
    }

    updateUser(id, { username, password, passwordConfirm });

    setPassword("");
    setPasswordConfirm("");
  };

  if (processing || id === null)
    return <LoadingScreen title={["Cargando..."]} description={[""]} />;

  return (
    <div>
      <Header title="Mi perfil" />
      <div className="max-w-3xl mx-auto mt-4 bg-white rounded-lg shadow-md p-8">
        <form onSubmit={updateProfileInfo}>
          <div className="grid gap-6 sm:grid-cols-2 sm:gap-8">
            <div className="sm:col-span-2">
              <label
                htmlFor="name"
                className="block mb-3 text-base font-medium text-gray-900 dark:text-white"
              >
                Username
              </label>
              <input
                autoComplete="off"
                disabled={true}
                type="text"
                name="name"
                id="name"
                value={username}
                onChange={(e) => setUser(e.target.value)}
                className="bg-gray-50 border border-gray-300 text-gray-900 text-base rounded-lg focus:ring-indigo-600 focus:border-indigo-600 block w-full p-4 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white"
                placeholder="Username"
                required={true}
              />
            </div>
            <div className="sm:col-span-2">
              <label
                htmlFor="password"
                className="block mb-3 text-base font-medium text-gray-900 dark:text-white"
              >
                Nueva contraseña
              </label>
              <input
                autoComplete="off"
                type="password"
                name="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="bg-gray-50 border border-gray-300 text-gray-900 text-base rounded-lg focus:ring-indigo-600 focus:border-indigo-600 block w-full p-4 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white"
                placeholder="Password"
                required={true}
              />
            </div>
            <div className="sm:col-span-2">
              <label
                htmlFor="passwordConfirm"
                className="block mb-3 text-base font-medium text-gray-900 dark:text-white"
              >
                Repetir contraseña
              </label>
              <input
                autoComplete="off"
                type="password"
                name="passwordConfirm"
                id="passwordConfirm"
                value={passwordConfirm}
                onChange={(e) => setPasswordConfirm(e.target.value)}
                className="bg-gray-50 border border-gray-300 text-gray-900 text-base rounded-lg focus:ring-indigo-600 focus:border-indigo-600 block w-full p-4 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white"
                placeholder="Password"
                required={true}
              />
            </div>
          </div>
          <div className="flex justify-between mt-6">
            <Button disabled={processing} variant="primary" type="submit">
              {processing ? "Guardando..." : "Editar perfil"}
            </Button>
          </div>
        </form>
        {errorMessage && (
          <div
            className="p-4 mt-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
            role="alert"
          >
            <span className="font-medium">Error!</span> {errorMessage}
          </div>
        )}
      </div>
      <BaseModal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          auth.logout();
        }}
        title="Contraseña modificada con éxito"
        message="Debe iniciar sesión nuevamente para continuar."
        primaryButtonText="Aceptar"
        onPrimaryAction={() => {
          setIsModalOpen(false);
          auth.logout();
        }}
      />
    </div>
  );
}
