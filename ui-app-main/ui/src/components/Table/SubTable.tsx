export type SubColumn<T> = {
  key: keyof T;
  header: string;
  render?: (value: T[keyof T], item: T) => React.ReactNode;
};

type SubTableProps<T> = {
  data: T[];
  columns: SubColumn<T>[];
  className?: string;
};

export const SubTable = <T,>({
  data,
  columns,
  className,
}: SubTableProps<T>) => {
  return (
    <div className={`overflow-x-auto rounded-xl ${className ?? ""}`}>
      <table className="w-full text-sm text-left text-gray-800 bg-white rounded-xl shadow border border-gray-400 overflow-hidden">
        <thead className="bg-gray-300 text-gray-900">
          <tr>
            {columns.map((col) => (
              <th key={String(col.key)} className="px-4 py-2">
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((item, idx) => (
            <tr key={idx} className="border-t border-gray-500">
              {columns.map((col) => (
                <td key={String(col.key)} className="px-4 py-2">
                  {col.render
                    ? col.render(item[col.key], item)
                    : String(item[col.key])}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
