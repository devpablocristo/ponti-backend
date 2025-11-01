export default function Cover() {
  return (
    <div
      className="hidden md:block md:w-3/5 relative bg-cover bg-center"
      style={{
        backgroundImage: "url('/LeftContent.png')",
      }}
    >
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <div className="flex items-center space-x-4 text-white px-4 text-left">
          <div className="flex flex-col justify-center mt-10">
            <h1 className="text-8xl font-bold whitespace-nowrap">Ponti</h1>
          </div>
        </div>
        <p className="text-lg leading-tight text-white ml-64">
          Es gestión. Es trazabilidad. <br />
          <span className="font-semibold">Es simple.</span>
        </p>
      </div>
      <p className="w-full absolute bottom-4 text-sm text-white text-center">
        Ponti Software versión Beta 1.0
      </p>
    </div>
  );
}
