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
      // 运行时 JSON 数据行可能缺列（Record 索引访问实际可返回 undefined），断言补齐类型缺口
      const valA = a[sortKey] as LpwCellValue | undefined
      const valB = b[sortKey] as LpwCellValue | undefined

      const isNullA = valA === null || valA === undefined
      const isNullB = valB === null || valB === undefined

      // null / 缺列恒在末尾，且满足严格弱序自反性
      if (isNullA && isNullB) return 0
      if (isNullA) return 1
      if (isNullB) return -1

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
      className="my-8 overflow-x-auto border-t-2 border-b-2 border-sea-ink bg-surface/30 shadow-2xs font-sans"
    >
      <table className="w-full border-collapse text-xs text-sea-ink">
        <thead>
          <tr className="border-b border-sea-ink bg-surface/50">
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
                  className={`p-3.5 font-mono text-[11px] font-bold uppercase tracking-wider text-sea-ink ${getAlignClass(
                    col.align,
                  )}`}
                >
                  {props.sortable ? (
                    <button
                      type="button"
                      aria-sort={ariaSort}
                      onClick={() => handleSort(col.key)}
                      className="inline-flex cursor-pointer items-center gap-1.5 select-none hover:text-lagoon transition-colors"
                    >
                      <span>{col.title}</span>
                      <span className="font-mono text-[10px] text-sea-ink-soft">
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
                className="p-8 text-center text-xs font-serif italic text-sea-ink-soft/70"
              >
                暂无数据
              </td>
            </tr>
          ) : (
            sortedData.map((row, rowIdx) => (
              <tr
                key={rowIdx}
                className="transition-colors hover:bg-surface/60"
              >
                {props.columns.map((col) => (
                  <td
                    key={col.key}
                    className={`p-3.5 text-xs text-sea-ink/90 ${getAlignClass(col.align)}`}
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
