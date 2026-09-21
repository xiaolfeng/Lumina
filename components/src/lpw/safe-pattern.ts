/**
 * 批注划线正则的静态安全检查。
 *
 * 后端（Go RE2）按线性时间校验 pattern 并持久化；前端渲染在浏览器主线程
 * 用回溯型引擎（V8 Irregexp）执行同一段 pattern，存在指数回溯（ReDoS）风险。
 * 这里在构造 RegExp 之前做保守的语法子集检查，命中危险构造即回退原文渲染：
 *
 * 1. 长度超过 128 字符直接拒绝（与后端 Schema 上限一致）；
 * 2. 拒绝嵌套量词（star height > 1），如 (a+)+、(\d{2,3})+、((ab)+)?；
 * 3. 拒绝量词化的含交替组，如 (a|a)*、(?:x|y)+；
 * 4. 拒绝未闭合的括号（与 RegExp 构造报错对齐，双保险）。
 *
 * 字符类 [...] 内的量词字符是字面量，不参与上述判定。
 */

const MAX_PATTERN_LENGTH = 128

interface GroupFrame {
  /** 组内（含嵌套组传递）出现过交替 | */
  hasAlternation: boolean
  /** 组内存在被量词修饰过的原子 */
  hasQuantifiedAtom: boolean
  /** atom=捕获/非捕获组（可被量词化）；assert=前瞻/后顾断言（不可被量词化） */
  kind: 'atom' | 'assert'
}

export function isSafeAnnotationPattern(pattern: string): boolean {
  if (pattern.length === 0 || pattern.length > MAX_PATTERN_LENGTH) {
    return false
  }

  const stack: GroupFrame[] = []
  let lastAtom: 'none' | 'char' | 'group' = 'none'
  let lastGroupHasQuantifiedAtom = false
  let lastGroupHasAlternation = false

  /** 量词作用于最近一个原子；返回 false 表示应拒绝该 pattern */
  const markQuantified = (): boolean => {
    const top = stack.length > 0 ? stack[stack.length - 1] : null
    if (lastAtom === 'group') {
      if (lastGroupHasQuantifiedAtom) return false // 嵌套量词：star height > 1
      if (lastGroupHasAlternation) return false // 量词化交替组
      if (top) top.hasQuantifiedAtom = true
      lastAtom = 'none'
      return true
    }
    if (lastAtom === 'char') {
      if (top) top.hasQuantifiedAtom = true
      lastAtom = 'none'
      return true
    }
    return false // 量词前没有原子，交给 RegExp 构造报错路径，这里直接拒绝
  }

  let i = 0
  while (i < pattern.length) {
    const ch = pattern[i]
    switch (ch) {
      case '\\': {
        const next = pattern[i + 1]
        // Q-09 修复：Unicode 属性转义 \p{...}/\P{...} 与码点转义 \u{...} 整体作为单个字符原子消费，
        // 防止内部花括号被误判为量词导致 (\p{L})+ 等安全模式被判为嵌套量词
        if (
          (next === 'p' || next === 'P' || next === 'u') &&
          pattern[i + 2] === '{'
        ) {
          const closeIdx = pattern.indexOf('}', i + 3)
          if (closeIdx !== -1) {
            const inner = pattern.slice(i + 3, closeIdx)
            if (/^[a-zA-Z0-9_=-]+$/.test(inner)) {
              i = closeIdx + 1
              lastAtom = 'char'
              continue
            }
          }
        }
        i += 2
        lastAtom = 'char'
        continue
      }
      case '[': {
        i++
        if (pattern[i] === '^') i++
        if (pattern[i] === ']') i++
        while (i < pattern.length && pattern[i] !== ']') {
          if (pattern[i] === '\\') i++
          i++
        }
        i++ // 消费闭合 ]
        lastAtom = 'char'
        continue
      }
      case '(': {
        const frame: GroupFrame = {
          hasAlternation: false,
          hasQuantifiedAtom: false,
          kind: 'atom',
        }
        stack.push(frame)
        i++
        // 消费组修饰前缀：(?: (?= (?! (?<= (?<! (?<name>
        if (pattern[i] === '?') {
          i++
          if (pattern[i] === '<') {
            i++
            if (pattern[i] === '=' || pattern[i] === '!') {
              frame.kind = 'assert'
              i++
            } else {
              // 命名捕获组 (?<name>，其余 (?<= (?<! 已在上面处理
              while (i < pattern.length && pattern[i] !== '>') i++
              i++ // 消费 >
            }
          } else if (pattern[i] === '=' || pattern[i] === '!' || pattern[i] === ':') {
            if (pattern[i] !== ':') frame.kind = 'assert'
            i++
          }
          // 其余字符不构成合法组前缀，交给末尾的 RegExp 构造报错路径
        }
        lastAtom = 'none'
        continue
      }
      case ')': {
        const frame = stack.pop()
        if (!frame) return false
        // 断言组不是可量词化原子；量词化的断言在正则语法中非法
        lastAtom = frame.kind === 'atom' ? 'group' : 'none'
        lastGroupHasQuantifiedAtom = frame.hasQuantifiedAtom
        lastGroupHasAlternation = frame.hasAlternation
        i++
        continue
      }
      case '|': {
        if (stack.length > 0) {
          stack[stack.length - 1].hasAlternation = true
        }
        lastAtom = 'none'
        i++
        continue
      }
      case '*':
      case '+':
      case '?': {
        if (!markQuantified()) return false
        i++
        if (pattern[i] === '?') i++ // 懒惰量词
        continue
      }
      case '{': {
        if (lastAtom === 'none') {
          // 非法量词位置（如 ^{2}）：按字面量字符处理
          i++
          lastAtom = 'char'
          continue
        }
        if (!markQuantified()) return false
        i++
        while (i < pattern.length && pattern[i] !== '}') i++
        i++ // 消费闭合 }
        continue
      }
      default: {
        i++
        lastAtom = 'char'
        continue
      }
    }
  }

  return stack.length === 0
}
