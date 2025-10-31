const FormContainer: React.FC<{
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
  children: React.ReactNode;
}> = ({ onSubmit, children }) => (
  <div className="max-w-6xl mx-auto mt-8 bg-white rounded-lg shadow-md p-8">
    <form onSubmit={onSubmit}>{children}</form>
  </div>
);

export default FormContainer;
