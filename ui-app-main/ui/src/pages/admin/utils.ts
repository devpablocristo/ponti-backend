export const formatNumberAr = (value: number | string) =>
  new Intl.NumberFormat("es-AR", { maximumFractionDigits: 2 }).format(
    Number(value) || 0
  );