import { useNavigate } from "react-router-dom";
import { AuthProvider, useAuth } from "./context/AuthProvider";
import { useEffect, useState } from "react";
import { UserData } from "./types";
import { RequestError } from "../../restclient/types";
import Cover from "./Cover";

function Login() {
  const { login, isAuthenticated, loading } = useAuth();
  const navigate = useNavigate();

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLogin, setIsLogin] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!loading && isAuthenticated) {
      navigate("/workspace");
    }
  }, [isAuthenticated, loading, navigate]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();

    const loginUserData: UserData = {
      username,
      password,
    };

    setIsLogin(true);

    try {
      setError("");
      await login(loginUserData);
    } catch (error) {
      setIsLogin(false);
      if (error instanceof RequestError) {
        setError(error.message);
        return;
      }
      setError("Error en el inicio de sesión. Intenta nuevamente.");
    }
  };

  return (
    <div className="flex flex-col md:flex-row h-screen">
      <Cover />
      <div className="w-full md:w-2/5 flex-1 flex items-center justify-center bg-gray-50 p-4">
        <div className="w-full max-w-xs">
          <h2 className="text-4xl font-bold text-center mb-3">Ingreso</h2>
          <p className="text-center text-base text-gray-900 mb-6">
            Bienvenido a su software de gestión.
            <br />
            Por favor ingrese su usuario y contraseña.
          </p>
          <form className="space-y-4 md:space-y-6" onSubmit={handleLogin}>
            <div className="flex items-center h-[60px] bg-white border border-[#EEEEEE] rounded-xl px-[26px] py-[18px] gap-[10px]">
              <svg
                width="20"
                height="20"
                viewBox="0 0 20 20"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M8.65399 10.9381H11.3454C12.4063 10.9394 13.4209 11.3527 14.1677 12.0826C14.914 12.8121 15.3323 13.7983 15.3337 14.8248V17.4557C15.3337 17.5518 15.2944 17.6468 15.2214 17.7184C15.1477 17.7904 15.0451 17.8335 14.9362 17.8336H5.06415C4.95511 17.8336 4.8527 17.7904 4.77899 17.7184C4.70583 17.6468 4.66669 17.5519 4.66669 17.4557V14.8248L4.67157 14.6334C4.72116 13.6764 5.13214 12.7666 5.83173 12.0826C6.53171 11.3984 7.46755 10.9931 8.45575 10.944L8.65399 10.9381ZM9.99969 2.16663C11.9647 2.16663 13.5387 3.7207 13.5388 5.61389C13.5388 7.50713 11.9647 9.06116 9.99969 9.06116C8.03479 9.06099 6.46161 7.50702 6.46161 5.61389C6.46167 3.72081 8.03483 2.1668 9.99969 2.16663Z"
                  fill="#6B7280"
                  stroke="#6B7280"
                />
              </svg>

              <input
                type="text"
                placeholder="Usuario"
                className="w-full bg-transparent focus:outline-none text-gray-700 text-base placeholder-gray-400"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required={true}
              />
            </div>
            <div className="flex items-center h-[60px] bg-white border border-[#EEEEEE] rounded-xl px-[26px] py-[18px] gap-[10px]">
              <svg
                width="20"
                height="20"
                viewBox="0 0 20 20"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M10.0003 3C10.8781 3.00008 11.7104 3.31471 12.3167 3.86035C12.9214 4.40458 13.2503 5.13052 13.2503 5.875V8.25H15.0003C15.3255 8.25008 15.6284 8.36656 15.8441 8.56055C16.058 8.75314 16.1663 9.0028 16.1663 9.25V16C16.1663 16.2472 16.058 16.4969 15.8441 16.6895C15.6284 16.8834 15.3255 16.9999 15.0003 17H5.00031C4.67488 17 4.37131 16.8836 4.15558 16.6895C3.94159 16.4969 3.83331 16.2472 3.83331 16V9.25C3.83331 9.0028 3.94159 8.75314 4.15558 8.56055C4.37131 8.3664 4.67488 8.25 5.00031 8.25H6.75031V5.875C6.75031 5.13054 7.07825 4.40457 7.68292 3.86035C8.28935 3.31456 9.12234 3 10.0003 3ZM10.0003 10.25C9.66283 10.25 9.32955 10.37 9.07648 10.5977C8.82163 10.827 8.66632 11.1505 8.66632 11.5V13.75C8.66632 14.0995 8.82163 14.423 9.07648 14.6523C9.32955 14.88 9.66283 15 10.0003 15C10.3378 14.9999 10.6711 14.8801 10.9241 14.6523C11.179 14.423 11.3333 14.0995 11.3333 13.75V11.5C11.3333 11.1505 11.179 10.827 10.9241 10.5977C10.6711 10.3699 10.3378 10.2501 10.0003 10.25ZM10.0003 3.5C9.33121 3.5 8.68021 3.739 8.19269 4.17773C7.70342 4.61808 7.41632 5.22709 7.41632 5.875V8.25H12.5833V5.875C12.5833 5.22722 12.297 4.61806 11.8079 4.17773C11.3205 3.73903 10.6693 3.50008 10.0003 3.5Z"
                  fill="#6B7280"
                  stroke="#6B7280"
                />
              </svg>

              <input
                type="password"
                placeholder="Contraseña"
                className="w-full bg-transparent focus:outline-none text-gray-700 text-base placeholder-gray-400"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required={true}
              />
            </div>
            <button
              type="submit"
              disabled={isLogin}
              className="bg-custom-btn text-white font-semibold text-base min-w-80 h-[55px] rounded-xl px-[26px] py-[16px] hover:bg-custom-btn/80 transition duration-300"
            >
              {isLogin ? (
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
                "Siguiente"
              )}
            </button>
          </form>
          {error !== "" && (
            <div
              className="flex items-center p-4 my-4 text-sm text-red-800 rounded-lg bg-red-50 dark:bg-gray-800 dark:text-red-400"
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
                <span className="font-medium">Error!</span> {error}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

const SignInPage = () => (
  <AuthProvider>
    <Login />
  </AuthProvider>
);

export default SignInPage;
