export function Table({ columns, data, renderRow }) {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-[720px] w-full border-collapse text-left">
        <thead>
          <tr className="border-b border-slate-200 bg-slate-50">
            {columns.map((column) => (
              <th key={column} className="px-5 py-3 text-xs font-extrabold uppercase text-slate-500">
                {column}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>{data.map(renderRow)}</tbody>
      </table>
    </div>
  )
}
