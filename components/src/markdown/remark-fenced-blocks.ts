/**
 * remark-fenced-blocks.ts
 *
 * remark plugin that scans for `:::type{props}` ... `:::` blocks
 * and converts them to custom mdast nodes.
 *
 * Custom mdast node format:
 *   { type: 'callout', data: { hName: 'callout', hProperties: { type: 'info' } }, children: [...] }
 *
 * remark-rehype will convert the custom node to a hast element
 * (tagName='callout'), and react-markdown's `components` prop
 * maps it to a React component.
 *
 * Supported types: callout, card, steps, step
 * Nesting: :::steps can contain multiple :::step blocks.
 * Unclosed fences are left as plain text (no error).
 */

interface MdastNode {
  type: string
  [key: string]: unknown
}

interface MdastParent extends MdastNode {
  children: MdastNode[]
}

function isOpeningFence(node: MdastNode): boolean {
  if (node.type !== 'paragraph') return false
  const parent = node as MdastParent
  if (parent.children.length === 0) return false
  if (parent.children.length > 1) return false
  const firstChild = parent.children[0]
  if (firstChild.type !== 'text') return false
  const value = (firstChild as unknown as { value: string }).value
  return /^:::\w+/.test(value)
}

function parseOpeningFence(node: MdastNode): { type: string; props: string } {
  const parent = node as MdastParent
  const firstChild = parent.children[0] as unknown as { value: string }
  const text = firstChild.value
  const match = text.match(/^:::(\w+)(?:\{([^}]*)\})?\s*$/)
  if (!match) return { type: '', props: '' }
  return { type: match[1], props: match[2] || '' }
}

function isClosingFence(node: MdastNode): boolean {
  if (node.type !== 'paragraph') return false
  const parent = node as MdastParent
  if (parent.children.length === 0) return false
  if (parent.children.length > 1) return false
  const firstChild = parent.children[0]
  if (firstChild.type !== 'text') return false
  const value = (firstChild as unknown as { value: string }).value
  return /^:::\s*$/.test(value)
}

function parseProps(propsStr: string): Record<string, string> {
  const props: Record<string, string> = {}
  if (!propsStr) return props

  const regex = /(\w+)=["']([^"']+)["']/g
  let match: RegExpExecArray | null
  while ((match = regex.exec(propsStr)) !== null) {
    props[match[1]] = match[2]
  }

  return props
}

function collectFencedBlock(
  children: MdastNode[],
  startIndex: number,
): { contentNodes: MdastNode[]; endIndex: number } | null {
  const contentNodes: MdastNode[] = []
  let depth = 1

  for (let i = startIndex; i < children.length; i++) {
    const node = children[i]

    if (isOpeningFence(node)) {
      depth++
      contentNodes.push(node)
    } else if (isClosingFence(node)) {
      depth--
      if (depth === 0) {
        return { contentNodes, endIndex: i }
      }
      contentNodes.push(node)
    } else {
      contentNodes.push(node)
    }
  }

  // No closing found
  return null
}

function processChildren(children: MdastNode[]): MdastNode[] {
  const result: MdastNode[] = []
  let i = 0

  while (i < children.length) {
    const node = children[i]

    if (isOpeningFence(node)) {
      const { type, props } = parseOpeningFence(node)
      const blockResult = collectFencedBlock(children, i + 1)

      if (blockResult) {
        const { contentNodes, endIndex } = blockResult

        // Recursively process content nodes
        const processedContent = processChildren(contentNodes)

        const customNode: MdastNode = {
          type,
          data: {
            hName: type,
            hProperties: parseProps(props),
          },
          children: processedContent,
        }

        result.push(customNode)
        i = endIndex + 1
        continue
      }
    }

    // For non-fence nodes, recursively process their children
    if ('children' in node && Array.isArray((node as MdastParent).children)) {
      const parent = node as MdastParent
      parent.children = processChildren(parent.children)
    }

    result.push(node)
    i++
  }

  return result
}

export function remarkFencedBlocks() {
  return (tree: MdastNode) => {
    if (!('children' in tree) || !Array.isArray((tree as MdastParent).children))
      return

    const parent = tree as MdastParent
    parent.children = processChildren(parent.children)
  }
}
