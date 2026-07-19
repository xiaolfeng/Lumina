import { describe, expect, it, beforeEach, afterEach } from 'vitest'
import { render, screen, cleanup } from '@testing-library/react'
import { Toc } from './toc'

class MockIntersectionObserver implements IntersectionObserver {
  callback: IntersectionObserverCallback
  elements: Element[] = []
  root: Element | Document | null = null
  rootMargin = ''
  scrollMargin = ''
  thresholds: ReadonlyArray<number> = []

  constructor(callback: IntersectionObserverCallback) {
    this.callback = callback
  }

  observe(element: Element) {
    this.elements.push(element)
  }

  unobserve(element: Element) {
    this.elements = this.elements.filter((el) => el !== element)
  }

  disconnect() {
    this.elements = []
  }

  takeRecords(): IntersectionObserverEntry[] {
    return []
  }
}

describe('Toc', () => {
  let originalIntersectionObserver: typeof IntersectionObserver

  beforeEach(() => {
    originalIntersectionObserver = global.IntersectionObserver
    global.IntersectionObserver = MockIntersectionObserver
  })

  afterEach(() => {
    global.IntersectionObserver = originalIntersectionObserver
    cleanup()
  })

  it('returns null when content has no headings', () => {
    const { container } = render(<Toc content="Just some plain text without any headings." />)
    expect(container.firstChild).toBeNull()
  })

  it('returns null for empty string', () => {
    const { container } = render(<Toc content="" />)
    expect(container.firstChild).toBeNull()
  })

  it('renders list items for h2 and h3 headings', () => {
    const content = `
## Getting Started

Some text.

### Installation

More text.

### Configuration

Even more.

## API Reference

Final section.
`
    render(<Toc content={content} />)

    const links = screen.getAllByRole('link')
    expect(links).toHaveLength(4)
    expect(links[0].textContent).toBe('Getting Started')
    expect(links[1].textContent).toBe('Installation')
    expect(links[2].textContent).toBe('Configuration')
    expect(links[3].textContent).toBe('API Reference')
  })

  it('renders h4 headings with deeper indent', () => {
    const content = `
## Section A

### Subsection B

#### Deep C

Text.
`
    render(<Toc content={content} />)

    const links = screen.getAllByRole('link')
    expect(links).toHaveLength(3)
    expect(links[0].textContent).toBe('Section A')
    expect(links[1].textContent).toBe('Subsection B')
    expect(links[2].textContent).toBe('Deep C')
  })

  it('ignores h1 headings', () => {
    const content = `
# Page Title

## Section One

### Sub One

## Section Two
`
    render(<Toc content={content} />)

    const links = screen.getAllByRole('link')
    expect(links).toHaveLength(3)
    expect(links[0].textContent).toBe('Section One')
    expect(links[1].textContent).toBe('Sub One')
    expect(links[2].textContent).toBe('Section Two')
  })

  it('renders anchor links with correct href', () => {
    const content = '## Hello World'
    render(<Toc content={content} />)

    const link = screen.getByRole('link')
    expect(link.getAttribute('href')).toBe('#hello-world')
  })
})
