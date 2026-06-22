import {
  Children,
  cloneElement,
  createContext,
  isValidElement,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type CSSProperties,
  type ReactElement,
  type ReactNode,
} from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronDown, FileWarning } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Markdown } from '@/components/ui/markdown'
import { Skeleton } from '@/components/ui/skeleton'
import { CopyButton } from '@/components/copy-button'
import { PublicLayout } from '@/components/layout'
import type { LegalDocumentResponse } from './types'

type LegalDocumentProps = {
  title: string
  queryKey: string
  fetchDocument: () => Promise<LegalDocumentResponse>
  emptyMessage: string
  headerExtra?: ReactNode
}

type TocHeading = {
  id: string
  text: string
  level: number
  line: number
}

type TocSection = {
  id: string
  text: string
  children: TocHeading[]
}

type TableColumnRole =
  | 'name'
  | 'type'
  | 'required'
  | 'endpoint'
  | 'example'
  | 'description'
  | 'default'

type DocsTableMeta = {
  layout: 'semantic' | 'default'
  roles: TableColumnRole[]
  widths: string[]
}

const DocsTableContext = createContext<DocsTableMeta | null>(null)

function isEndpointLike(value: string) {
  const trimmed = value.trim()

  return (
    /^(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)\s+\/\S+$/i.test(trimmed) ||
    /^\/(v\d+|api)\//i.test(trimmed) ||
    /^https?:\/\/\S+$/i.test(trimmed)
  )
}

function isSingleLineEndpointBlock(value: string) {
  const trimmed = value.trim()
  return !trimmed.includes('\n') && isEndpointLike(trimmed)
}

function getErrorMessage(error: unknown) {
  if (!error || typeof error !== 'object') {
    return ''
  }

  const errorWithResponse = error as {
    message?: string
    response?: {
      data?: {
        message?: string
      }
    }
  }

  return (
    errorWithResponse.response?.data?.message || errorWithResponse.message || ''
  )
}

function isValidUrl(value: string) {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

function flattenNodeText(node: ReactNode): string {
  if (typeof node === 'string' || typeof node === 'number') {
    return String(node)
  }

  if (Array.isArray(node)) {
    return node.map(flattenNodeText).join('')
  }

  if (!isValidElement(node)) {
    return ''
  }

  const element = node as React.ReactElement<{ children?: ReactNode }>
  return flattenNodeText(element.props.children)
}

function slugifyHeading(text: string): string {
  const normalized = text
    .trim()
    .toLowerCase()
    .replace(/[`*_~]/g, '')
    .replace(/[()[\]{}<>]/g, '')
    .replace(/[^\w\u4e00-\u9fff\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')

  return normalized || 'section'
}

function createHeadingId(text: string, counts: Map<string, number>): string {
  const baseId = slugifyHeading(text)
  const nextCount = (counts.get(baseId) ?? 0) + 1
  counts.set(baseId, nextCount)
  return nextCount === 1 ? baseId : `${baseId}-${nextCount}`
}

function extractTocHeadings(markdown: string): TocHeading[] {
  const headings: TocHeading[] = []
  const lines = markdown.split(/\r?\n/)
  const headingCounts = new Map<string, number>()
  let inCodeFence = false

  for (const [index, line] of lines.entries()) {
    if (/^\s*(```|~~~)/.test(line)) {
      inCodeFence = !inCodeFence
      continue
    }

    if (inCodeFence) {
      continue
    }

    const match = line.match(/^(#{1,4})\s+(.+?)\s*#*\s*$/)
    if (!match) {
      continue
    }

    const level = match[1].length
    const text = match[2].trim()
    if (!text) {
      continue
    }

    headings.push({
      id: createHeadingId(text, headingCounts),
      text,
      level,
      line: index + 1,
    })
  }

  return headings
}

function extractMarkdownNodeText(node: unknown): string {
  if (!node || typeof node !== 'object') {
    return ''
  }

  const markdownNode = node as {
    type?: string
    value?: string
    children?: unknown[]
  }

  if (markdownNode.type === 'text' && typeof markdownNode.value === 'string') {
    return markdownNode.value
  }

  return (markdownNode.children ?? [])
    .map((child) => extractMarkdownNodeText(child))
    .join('')
}

function getTableHeaderLabels(node: unknown): string[] {
  if (!node || typeof node !== 'object') {
    return []
  }

  const tableNode = node as {
    children?: Array<{
      tagName?: string
      children?: Array<{
        tagName?: string
        children?: Array<{
          tagName?: string
          children?: unknown[]
        }>
      }>
    }>
  }

  const thead = tableNode.children?.find((child) => child.tagName === 'thead')
  const headerRow = thead?.children?.find((child) => child.tagName === 'tr')
  const headerCells = headerRow?.children?.filter(
    (child) => child.tagName === 'th'
  )

  return (
    headerCells
      ?.map((cell) => extractMarkdownNodeText(cell).trim())
      .filter(Boolean) ?? []
  )
}

function getTableColumnCount(node: unknown): number {
  if (!node || typeof node !== 'object') {
    return 0
  }

  const tableNode = node as {
    children?: Array<{
      children?: Array<{
        tagName?: string
        children?: unknown[]
      }>
    }>
  }

  return (
    tableNode.children?.reduce((maxCount, section) => {
      const rowCounts =
        section.children
          ?.filter((child) => child.tagName === 'tr')
          .map((row) => row.children?.length ?? 0) ?? []

      return Math.max(maxCount, ...rowCounts, 0)
    }, 0) ?? 0
  )
}

function normalizeHeaderLabel(label: string) {
  return label.trim().toLowerCase().replace(/\s+/g, '')
}

function classifyColumnRole(
  label: string,
  index: number,
  totalColumns: number
): TableColumnRole {
  const normalized = normalizeHeaderLabel(label)

  if (
    /^(?:\u53c2\u6570|\u5b57\u6bb5|\u5c5e\u6027|\u540d\u79f0|\u53c2\u6570\u540d|\u5b57\u6bb5\u540d|header|headers|name|field|param|parameter|query|path)$/u.test(
      normalized
    )
  ) {
    return 'name'
  }

  if (/^(?:\u7c7b\u578b|\u683c\u5f0f|datatype|type)$/u.test(normalized)) {
    return 'type'
  }

  if (
    /^(?:\u5fc5\u586b|\u662f\u5426\u5fc5\u586b|required|required\?)$/u.test(
      normalized
    )
  ) {
    return 'required'
  }

  if (
    /^(?:\u63a5\u53e3|\u8c03\u7528\u63a5\u53e3|\u63a5\u53e3\u5730\u5740|\u8bf7\u6c42\u5730\u5740|endpoint|api|apiurl|url|uri|route)$/u.test(
      normalized
    )
  ) {
    return 'endpoint'
  }

  if (
    /^(?:\u793a\u4f8b|\u6837\u4f8b|example|sample|\u9ed8\u8ba4\u503c|default|value|\u53d6\u503c|\u679a\u4e3e|enum)$/u.test(
      normalized
    )
  ) {
    return 'example'
  }

  if (
    /^(?:\u8bf4\u660e|\u63cf\u8ff0|\u5907\u6ce8|\u542b\u4e49|\u7528\u9014|description|desc|note|notes|comment)$/u.test(
      normalized
    )
  ) {
    return 'description'
  }

  if (totalColumns === 5) {
    return (['name', 'type', 'required', 'example', 'description'][index] ??
      'default') as TableColumnRole
  }

  if (totalColumns === 4) {
    return (['name', 'type', 'endpoint', 'description'][index] ??
      'default') as TableColumnRole
  }

  if (totalColumns === 3) {
    return (['name', 'type', 'description'][index] ??
      'default') as TableColumnRole
  }

  return 'default'
}

function getSemanticColumnWidths(roles: TableColumnRole[]): string[] {
  const columnWeights: Record<TableColumnRole, number> = {
    name: 16,
    type: 11,
    required: 8,
    endpoint: 22,
    example: 24,
    description: 36,
    default: 18,
  }

  const totalWeight = roles.reduce(
    (sum, role) => sum + (columnWeights[role] ?? columnWeights.default),
    0
  )

  return roles.map((role) => {
    const weight = columnWeights[role] ?? columnWeights.default
    return `${((weight / totalWeight) * 100).toFixed(2)}%`
  })
}

function getDocsTableMeta(node: unknown): DocsTableMeta {
  const headerLabels = getTableHeaderLabels(node)
  const columnCount = headerLabels.length || getTableColumnCount(node)

  if (columnCount === 0) {
    return {
      layout: 'default',
      roles: [],
      widths: [],
    }
  }

  const roles = Array.from({ length: columnCount }, (_, index) =>
    classifyColumnRole(headerLabels[index] ?? '', index, columnCount)
  )
  const hasSemanticLayout = roles.some((role) => role !== 'default')

  return {
    layout: hasSemanticLayout ? 'semantic' : 'default',
    roles,
    widths: hasSemanticLayout ? getSemanticColumnWidths(roles) : [],
  }
}

function getColumnStyle(width?: string): CSSProperties | undefined {
  if (!width) {
    return undefined
  }

  return {
    width,
  }
}

function TocNav({
  sections,
  activeSectionId,
  activeHeadingId,
  title,
  expandedSectionId,
  onToggleSection,
  onNavigate,
}: {
  sections: TocSection[]
  activeSectionId: string
  activeHeadingId: string
  title: string
  expandedSectionId: string | null
  onToggleSection: (id: string) => void
  onNavigate: (id: string) => void
}) {
  return (
    <nav aria-label={title} className='space-y-1.5'>
      {sections.map((section) => {
        const isExpanded = expandedSectionId === section.id
        const isActiveSection = activeSectionId === section.id
        const hasChildren = section.children.length > 0

        return (
          <div key={section.id} className='space-y-1'>
            <div
              className='docs-toc-link docs-toc-link--section'
              data-active={isActiveSection}
            >
              <button
                type='button'
                onClick={() => {
                  if (hasChildren) {
                    onToggleSection(section.id)
                  }
                  onNavigate(section.id)
                }}
                className='flex min-w-0 flex-1 items-start gap-2 text-left'
              >
                <span className='docs-toc-badge'>H2</span>
                <span className='min-w-0 break-words'>{section.text}</span>
              </button>

              {hasChildren && (
                <button
                  type='button'
                  aria-label={section.text}
                  onClick={(event) => {
                    event.stopPropagation()
                    onToggleSection(section.id)
                  }}
                  className='text-muted-foreground hover:text-foreground rounded-lg p-1 transition-colors'
                >
                  <ChevronDown
                    className={cn(
                      'h-4 w-4 transition-transform',
                      isExpanded && 'rotate-180'
                    )}
                  />
                </button>
              )}
            </div>

            {isExpanded && hasChildren && (
              <div className='border-border/50 space-y-1 border-l pl-3'>
                {section.children.map((heading) => {
                  const isActive = heading.id === activeHeadingId

                  return (
                    <a
                      key={heading.id}
                      href={`#${heading.id}`}
                      onClick={(event) => {
                        event.preventDefault()
                        onNavigate(heading.id)
                      }}
                      className='docs-toc-link docs-toc-link--child'
                      data-active={isActive}
                    >
                      <span className='docs-toc-badge'>H3</span>
                      <span className='min-w-0 break-words'>
                        {heading.text}
                      </span>
                    </a>
                  )
                })}
              </div>
            )}
          </div>
        )
      })}
    </nav>
  )
}

function DocsTableRow({
  children,
  node: _node,
  ...props
}: {
  children?: ReactNode
  node?: unknown
  className?: string
}) {
  const tableMeta = useContext(DocsTableContext)

  return (
    <tr {...props}>
      {Children.map(children, (child, index) => {
        if (!isValidElement(child)) {
          return child
        }

        return cloneElement(child as ReactElement<Record<string, unknown>>, {
          'data-col-index': index + 1,
          'data-col-role': tableMeta?.roles[index] ?? 'default',
        })
      })}
    </tr>
  )
}

export function LegalDocument({
  title,
  queryKey,
  fetchDocument,
  emptyMessage,
  headerExtra,
}: LegalDocumentProps) {
  const { t } = useTranslation()
  const { data, isError, error, isLoading } = useQuery({
    queryKey: [queryKey],
    queryFn: fetchDocument,
    staleTime: 10 * 60 * 1000,
  })

  const rawContent = data?.data?.trim() ?? ''
  const hasContent = rawContent.length > 0
  const isUrl = hasContent && isValidUrl(rawContent)
  const success = data?.success ?? false
  const tocHeadings = useMemo(
    () => extractTocHeadings(rawContent),
    [rawContent]
  )
  const [observedActiveHeadingId, setObservedActiveHeadingId] = useState('')
  const activeHeadingId =
    tocHeadings.length === 0 ? '' : observedActiveHeadingId || tocHeadings[0].id
  const tocSections = useMemo(() => {
    const sections: TocSection[] = []
    let currentSection: TocSection | null = null

    for (const heading of tocHeadings) {
      if (heading.level === 2) {
        currentSection = {
          id: heading.id,
          text: heading.text,
          children: [],
        }
        sections.push(currentSection)
        continue
      }

      if (heading.level === 3 && currentSection) {
        currentSection.children.push(heading)
      }
    }

    return sections
  }, [tocHeadings])
  const activeSectionId = useMemo(() => {
    for (const section of tocSections) {
      if (
        section.id === activeHeadingId ||
        section.children.some((child) => child.id === activeHeadingId)
      ) {
        return section.id
      }
    }

    return tocSections[0]?.id ?? ''
  }, [activeHeadingId, tocSections])
  const [expandedSectionId, setExpandedSectionId] = useState<string | null>(
    null
  )
  const derivedExpandedSectionId = useMemo(() => {
    const activeChildSection = tocSections.find((section) =>
      section.children.some((child) => child.id === activeHeadingId)
    )

    return expandedSectionId ?? activeChildSection?.id ?? null
  }, [activeHeadingId, expandedSectionId, tocSections])

  const scrollToHeading = useCallback((headingId: string) => {
    const element = document.getElementById(headingId)
    if (!element) {
      return
    }

    const top = element.getBoundingClientRect().top + window.scrollY - 112
    window.scrollTo({
      top: Math.max(0, top),
      behavior: 'smooth',
    })
    setObservedActiveHeadingId(headingId)
    window.history.replaceState(null, '', `#${headingId}`)
  }, [])

  useEffect(() => {
    if (tocHeadings.length === 0) {
      return
    }

    const updateActiveHeading = () => {
      const offset = 140
      let nextActiveId = tocHeadings[0].id

      for (const heading of tocHeadings) {
        const element = document.getElementById(heading.id)
        if (!element) {
          continue
        }

        if (element.getBoundingClientRect().top - offset <= 0) {
          nextActiveId = heading.id
        }
      }

      setObservedActiveHeadingId(nextActiveId)
    }

    const frame = window.requestAnimationFrame(updateActiveHeading)
    window.addEventListener('scroll', updateActiveHeading, { passive: true })
    window.addEventListener('resize', updateActiveHeading)

    return () => {
      window.cancelAnimationFrame(frame)
      window.removeEventListener('scroll', updateActiveHeading)
      window.removeEventListener('resize', updateActiveHeading)
    }
  }, [tocHeadings])

  useEffect(() => {
    if (tocHeadings.length === 0 || !window.location.hash) {
      return
    }

    const hash = decodeURIComponent(window.location.hash.slice(1))
    const headingExists = tocHeadings.some((heading) => heading.id === hash)
    if (!headingExists) {
      return
    }

    const frame = window.requestAnimationFrame(() => {
      scrollToHeading(hash)
    })

    return () => window.cancelAnimationFrame(frame)
  }, [scrollToHeading, tocHeadings])

  const renderedDocument = useMemo(() => {
    const renderedHeadingCounts = new Map<string, number>()
    const resolveHeading = (
      expectedLevel: number,
      children: ReactNode,
      node?: unknown
    ): TocHeading => {
      const headingLine = (node as { position?: { start?: { line?: number } } })
        ?.position?.start?.line

      if (typeof headingLine === 'number') {
        const matchedHeading = tocHeadings.find(
          (heading) =>
            heading.level === expectedLevel && heading.line === headingLine
        )

        if (matchedHeading) {
          return matchedHeading
        }
      }

      const fallbackText = flattenNodeText(children).trim() || 'Section'
      return {
        id: createHeadingId(fallbackText, renderedHeadingCounts),
        text: fallbackText,
        level: expectedLevel,
        line: -1,
      }
    }

    const markdownComponents = {
      h1: ({
        children,
        node,
        ...props
      }: {
        children?: ReactNode
        node?: unknown
      }) => {
        const heading = resolveHeading(1, children, node)

        return (
          <h1
            {...props}
            id={heading.id}
            className='group/heading border-border/40 scroll-mt-28 border-b pb-4'
          >
            <a
              href={`#${heading.id}`}
              onClick={(event) => {
                event.preventDefault()
                scrollToHeading(heading.id)
              }}
              className='inline-flex items-center gap-2 no-underline'
            >
              <span>{children}</span>
              <span className='text-muted-foreground/0 group-hover/heading:text-muted-foreground font-mono text-sm transition-colors'>
                #
              </span>
            </a>
          </h1>
        )
      },
      h2: ({
        children,
        node,
        ...props
      }: {
        children?: ReactNode
        node?: unknown
      }) => {
        const heading = resolveHeading(2, children, node)

        return (
          <h2
            {...props}
            id={heading.id}
            className='group/heading border-border/25 scroll-mt-28 border-b pb-3'
          >
            <a
              href={`#${heading.id}`}
              onClick={(event) => {
                event.preventDefault()
                scrollToHeading(heading.id)
              }}
              className='inline-flex items-center gap-2 no-underline'
            >
              <span>{children}</span>
              <span className='text-muted-foreground/0 group-hover/heading:text-muted-foreground font-mono text-sm transition-colors'>
                #
              </span>
            </a>
          </h2>
        )
      },
      h3: ({
        children,
        node,
        ...props
      }: {
        children?: ReactNode
        node?: unknown
      }) => {
        const heading = resolveHeading(3, children, node)

        return (
          <h3 {...props} id={heading.id} className='group/heading scroll-mt-28'>
            <a
              href={`#${heading.id}`}
              onClick={(event) => {
                event.preventDefault()
                scrollToHeading(heading.id)
              }}
              className='inline-flex items-center gap-2 no-underline'
            >
              <span>{children}</span>
              <span className='text-muted-foreground/0 group-hover/heading:text-muted-foreground font-mono text-xs transition-colors'>
                #
              </span>
            </a>
          </h3>
        )
      },
      h4: ({
        children,
        node,
        ...props
      }: {
        children?: ReactNode
        node?: unknown
      }) => {
        const heading = resolveHeading(4, children, node)

        return (
          <h4 {...props} id={heading.id} className='group/heading scroll-mt-28'>
            <a
              href={`#${heading.id}`}
              onClick={(event) => {
                event.preventDefault()
                scrollToHeading(heading.id)
              }}
              className='inline-flex items-center gap-2 no-underline'
            >
              <span>{children}</span>
              <span className='text-muted-foreground/0 group-hover/heading:text-muted-foreground font-mono text-xs transition-colors'>
                #
              </span>
            </a>
          </h4>
        )
      },
      code: ({
        children,
        className,
        ...props
      }: {
        children?: ReactNode
        className?: string
        inline?: boolean
      }) => {
        const text = flattenNodeText(children).trim()
        const isInlineCode = !className

        if (isInlineCode && isEndpointLike(text)) {
          return (
            <span className='docs-endpoint-inline'>
              <code {...props} className='docs-endpoint-inline__code'>
                {text}
              </code>
              <CopyButton
                value={text}
                variant='ghost'
                size='icon'
                tooltip={t('Copy to clipboard')}
                successTooltip={t('Copied!')}
                className='docs-endpoint-inline__copy'
                iconClassName='h-3.5 w-3.5'
                aria-label={t('Copy to clipboard')}
              />
            </span>
          )
        }

        if (!isInlineCode && isSingleLineEndpointBlock(text)) {
          return (
            <span className='docs-endpoint-block'>
              <code {...props} className='docs-endpoint-block__code'>
                {text}
              </code>
              <CopyButton
                value={text}
                variant='outline'
                size='icon'
                tooltip={t('Copy to clipboard')}
                successTooltip={t('Copied!')}
                className='docs-endpoint-block__copy'
                iconClassName='h-3.5 w-3.5'
                aria-label={t('Copy to clipboard')}
              />
            </span>
          )
        }

        if (!isInlineCode) {
          return (
            <pre className='docs-code-block'>
              <code
                {...props}
                className={cn(className, 'docs-code-block__code')}
              >
                {children}
              </code>
            </pre>
          )
        }

        return (
          <code {...props} className={className}>
            {children}
          </code>
        )
      },
      pre: ({ children }: { children?: ReactNode }) => <>{children}</>,
      table: ({ children, node }: { children?: ReactNode; node?: unknown }) => {
        const tableMeta = getDocsTableMeta(node)

        return (
          <DocsTableContext.Provider value={tableMeta}>
            <div
              className='docs-table-wrap'
              data-docs-table-layout={tableMeta.layout}
              data-docs-table-columns={tableMeta.roles.length || undefined}
            >
              <table
                className='docs-table'
                data-docs-table-layout={tableMeta.layout}
                data-docs-table-columns={tableMeta.roles.length || undefined}
              >
                {tableMeta.layout === 'semantic' && (
                  <colgroup>
                    {tableMeta.roles.map((role, index) => (
                      <col
                        key={`${role}-${index}`}
                        className={`docs-table-col docs-table-col--${role}`}
                        style={getColumnStyle(tableMeta.widths[index])}
                      />
                    ))}
                  </colgroup>
                )}
                {children}
              </table>
            </div>
          </DocsTableContext.Provider>
        )
      },
      thead: ({ children }: { children?: ReactNode }) => (
        <thead>{children}</thead>
      ),
      tbody: ({ children }: { children?: ReactNode }) => (
        <tbody>{children}</tbody>
      ),
      tr: DocsTableRow,
      th: ({
        children,
        className,
        node: _node,
        ...props
      }: {
        children?: ReactNode
        className?: string
        node?: unknown
      }) => (
        <th {...props} className={cn('docs-table-head', className)}>
          {children}
        </th>
      ),
      td: ({
        children,
        className,
        node: _node,
        ...props
      }: {
        children?: ReactNode
        className?: string
        node?: unknown
      }) => (
        <td {...props} className={cn('docs-table-cell', className)}>
          {children}
        </td>
      ),
    }

    return (
      <Markdown
        className='docs-markdown prose-neutral dark:prose-invert max-w-none'
        components={markdownComponents}
      >
        {rawContent}
      </Markdown>
    )
  }, [rawContent, scrollToHeading, t, tocHeadings])

  if (isLoading) {
    return (
      <PublicLayout>
        <div className='mx-auto flex max-w-4xl flex-col gap-4 py-12'>
          <Skeleton className='h-8 w-[45%]' />
          <Skeleton className='h-4 w-full' />
          <Skeleton className='h-4 w-[90%]' />
          <Skeleton className='h-4 w-[80%]' />
        </div>
      </PublicLayout>
    )
  }

  if (isError) {
    const errorMessage = getErrorMessage(error)

    return (
      <PublicLayout>
        <div className='mx-auto max-w-2xl py-12'>
          <Card className='border-dashed'>
            <CardHeader className='flex flex-row items-center gap-4'>
              <div className='bg-muted rounded-full p-2'>
                <FileWarning className='text-muted-foreground h-5 w-5' />
              </div>
              <div className='space-y-1'>
                <CardTitle className='text-lg font-semibold'>{title}</CardTitle>
                <p className='text-muted-foreground text-sm'>
                  {errorMessage
                    ? `${t('Request failed')}: ${errorMessage}`
                    : `${t('Request failed')} ${t('Please try again later.')}`}
                </p>
              </div>
            </CardHeader>
          </Card>
        </div>
      </PublicLayout>
    )
  }

  if (!success || !hasContent) {
    return (
      <PublicLayout>
        <div className='mx-auto max-w-2xl py-12'>
          <Card className='border-dashed'>
            <CardHeader className='flex flex-row items-center gap-4'>
              <div className='bg-muted rounded-full p-2'>
                <FileWarning className='text-muted-foreground h-5 w-5' />
              </div>
              <div className='space-y-1'>
                <CardTitle className='text-lg font-semibold'>{title}</CardTitle>
                <p className='text-muted-foreground text-sm'>
                  {data?.message || emptyMessage}
                </p>
              </div>
            </CardHeader>
          </Card>
        </div>
      </PublicLayout>
    )
  }

  if (isUrl) {
    return (
      <PublicLayout>
        <div className='mx-auto max-w-2xl py-12'>
          <Card>
            <CardHeader>
              <CardTitle>{title}</CardTitle>
            </CardHeader>
            <CardContent className='space-y-4'>
              <p className='text-muted-foreground text-sm'>
                {t(
                  'The administrator configured an external link for this document.'
                )}
              </p>
              <Button asChild>
                <a href={rawContent} target='_blank' rel='noopener noreferrer'>
                  {t('View document')}
                </a>
              </Button>
            </CardContent>
          </Card>
        </div>
      </PublicLayout>
    )
  }

  return (
    <PublicLayout>
      <div className='docs-page'>
        <div className='docs-page__hero'>
          <h1 className='docs-page__title'>{title}</h1>
          {headerExtra}
        </div>

        <div className='relative'>
          <div className='docs-page__content'>
            <div className='docs-page__article'>{renderedDocument}</div>
          </div>

          {tocSections.length > 0 && (
            <aside className='hidden xl:block'>
              <div className='docs-page__rail'>
                <div className='docs-page__rail-card'>
                  <div>
                    <p className='docs-page__toc-heading'>
                      {t('On this page')}
                    </p>
                    <p className='docs-page__toc-subtitle'>
                      {t('Jump between document sections')}
                    </p>
                  </div>
                </div>

                <div className='docs-page__rail-scroll hover-scrollbar'>
                  <TocNav
                    sections={tocSections}
                    activeSectionId={activeSectionId}
                    activeHeadingId={activeHeadingId}
                    title={t('On this page')}
                    expandedSectionId={derivedExpandedSectionId}
                    onToggleSection={(sectionId) => {
                      setExpandedSectionId((current) =>
                        current === sectionId ? null : sectionId
                      )
                    }}
                    onNavigate={scrollToHeading}
                  />
                </div>
              </div>
            </aside>
          )}
        </div>
      </div>
    </PublicLayout>
  )
}
