import type React from 'react'
import { useMemo, useState } from 'react'
import type {
  LpwBlockSlotProps,
  LpwCellValue,
  LpwTableColumn,
  LpwTableProps,
} from '../types'

type SortDirection = 'asc' | 'desc' | null

export const TableBlock: React.FC<LpwBlockSlotProps<LpwTableProps>> = ({
  props,
}) => {
  const [sortKey, setSortKey] = useState<string | null>(null)
  const [sortDir, setSortDir] = useState<SortDirection>(null)

  const handleSort = (key: string) => {
    if (!props.sortable) return
    if (sortKey !== key) {
      setSortKey(key)
      setSortDir('asc')
    } else if (sortDir === 'asc') {
      setSortDir('desc')
    } else if (sortDir === 'desc') {
      setSortKey(null)
      setSortDir(null)
    }
  }

  const sortedData = useMemo(() => {
    if (!sortKey || !sortDir) return props.data

    const list = [...props.data]
    list.sort((a, b) => {
      const valA = a[sortKey]
      const valB = b[sortKey]

      // null 恒在末尾
      if (valA === null) return 1
      if (valB === null) return -1

      if (typeof valA === 'number' && typeof valB === 'number') {
        return sortDir === 'asc' ? valA - valB : valB - valA
      }

      const strA = String(valA)
      const strB = String(valB)
      return sortDir === 'asc'
        ? strA.localeCompare(strB)
        : strB.localeCompare(strA)
    })
    return list
  }, [props.data, sortKey, sortDir])

  const renderCell = (val: LpwCellValue | undefined) => {
    if (val === null || val === undefined) return '—'
    if (typeof val === 'boolean') return val ? '是' : '—'
    return String(val)
  }

  const getAlignClass = (align?: LpwTableColumn['align']) => {
    if (align === 'center') return 'text-center'
    if (align === 'right') return 'text-right'
    return 'text-left'
  }

  return (
    <div
      data-testid="table-block"
      className="my-4 overflow-x-auto border border-line bg-surface"
    >
      <table className="w-full border-collapse text-xs text-sea-ink">
        <thead>
          <tr className="border-b border-line bg-surface-muted/30">
            {props.columns.map((col) => {
              const isCurrent = sortKey === col.key
              const currentDir = isCurrent ? sortDir : null
              const ariaSort =
                currentDir === 'asc'
                  ? 'ascending'
                  : currentDir === 'desc'
                    ? 'descending'
                    : 'none'

              return (
                <th
                  key={col.key}
                  style={col.width ? { width: col.width } : undefined}
                  className={`p-3 font-semibold text-sea-ink ${getAlignClass(
                    col.align,
                  )}`}
                >
                  {props.sortable ? (
                    <button
                      type="button"
                      aria-sort={ariaSort}
                      onClick={() => handleSort(col.key)}
                      className="inline-flex cursor-pointer items-center gap-1 select-none hover:text-lagoon-deep"
                    >
                      <span>{col.title}</span>
                      <span className="font-mono text-[10px] text-sea-ink-soft/60">
                        {currentDir === 'asc'
                          ? '▲'
                          : currentDir === 'desc'
                            ? '▼'
                            : '⇅'}
                      </span>
                    </button>
                  ) : (
                    <span>{col.title}</span>
                  )}
                </th>
              )
            })}
          </tr>
        </thead>
        <tbody className="divide-y divide-line/60">
          {sortedData.length === 0 ? (
            <tr>
              <td
                colSpan={props.columns.length}
                className="p-8 text-center text-xs text-sea-ink-soft/60"
              >
                暂无数据
              </td>
            </tr>
          ) : (
            sortedData.map((row, rowIdx) => (
              <tr
                key={rowIdx}
                className="transition-colors hover:bg-surface-muted/20"
              >
                {props.columns.map((col) => (
                  <td
                    key={col.key}
                    className={`p-3 text-xs ${getAlignClass(col.align)}`}
                  >
                    {renderCell(row[col.key])}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}
