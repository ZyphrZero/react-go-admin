import { ChevronsUpDownIcon, InfoIcon, RefreshCcwIcon, SearchIcon, WaypointsIcon, XIcon } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'

import api from '@/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useErrorHandler } from '@/hooks/useErrorHandler'

const DEFAULT_PAGE_SIZE = 20
const PAGE_SIZE_OPTIONS = ['20', '50', '100']

const normalizeTagOption = (tag) => {
  if (!tag || typeof tag !== 'object') {
    return null
  }

  return {
    key: String(tag.value),
    label: String(tag.label),
    value: String(tag.value),
    count: typeof tag.count === 'number' ? tag.count : null,
  }
}

const ApiManagement = () => {
  const [loading, setLoading] = useState(false)
  const [apis, setApis] = useState([])
  const [total, setTotal] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [searchValues, setSearchValues] = useState({ path: '', summary: '', tags: [] })
  const [searchParams, setSearchParams] = useState({})
  const [availableTags, setAvailableTags] = useState([])
  const [refreshLoading, setRefreshLoading] = useState(false)

  const { handleError, handleBusinessError, showSuccess } = useErrorHandler()

  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const selectedTags = availableTags.filter((tag) => searchValues.tags.includes(tag.value))

  const fetchApis = useCallback(
    async (page = 1, size = 10, search = {}) => {
      setLoading(true)
      try {
        const response = await api.apis.getList({
          page,
          page_size: size,
          ...search,
        })
        setApis(response.data || [])
        setTotal(response.total || 0)
        setCurrentPage(response.page || page)
        setPageSize(response.page_size || size)
      } catch (error) {
        handleError(error, '获取 API 列表失败')
      } finally {
        setLoading(false)
      }
    },
    [handleError],
  )

  const fetchAllTags = useCallback(async () => {
    try {
      const response = await api.apis.getTags()
      const normalizedTags = Array.isArray(response.data)
        ? response.data.map(normalizeTagOption).filter(Boolean)
        : []

      setAvailableTags(normalizedTags)
    } catch (error) {
      handleError(error, '获取 API 标签失败')
    }
  }, [handleError])

  useEffect(() => {
    void fetchApis(1, DEFAULT_PAGE_SIZE, {})
    void fetchAllTags()
  }, [fetchAllTags, fetchApis])

  const handleSearch = async (event) => {
    event.preventDefault()

    const nextParams = {}
    if (searchValues.path.trim()) nextParams.path = searchValues.path.trim()
    if (searchValues.summary.trim()) nextParams.summary = searchValues.summary.trim()
    if (searchValues.tags.length > 0) nextParams.tags = searchValues.tags.join(',')

    setSearchParams(nextParams)
    await fetchApis(1, pageSize, nextParams)
  }

  const handleClearSearch = async () => {
    setSearchValues({ path: '', summary: '', tags: [] })
    setSearchParams({})
    await fetchApis(1, pageSize, {})
  }

  const handlePageChange = async (page) => {
    await fetchApis(page, pageSize, searchParams)
  }

  const handlePageSizeChange = async (value) => {
    const size = Number(value)
    setPageSize(size)
    await fetchApis(1, size, searchParams)
  }

  const handleRefreshApis = async () => {
    setRefreshLoading(true)
    try {
      await api.apis.refresh()
      showSuccess('API 列表刷新成功')
      await fetchApis(currentPage, pageSize, searchParams)
      await fetchAllTags()
    } catch (error) {
      handleBusinessError(error, 'API 刷新失败')
    } finally {
      setRefreshLoading(false)
    }
  }

  return (
    <div className="flex flex-col gap-5">
      <section className="flex flex-col gap-3 border-b pb-5 md:flex-row md:items-end md:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight">API 管理</h1>
            <InfoIcon className="size-4 text-muted-foreground" />
          </div>
          <p className="text-sm text-muted-foreground">系统根据后端模块注册自动同步接口信息，这里只提供查看与刷新能力</p>
        </div>
        <Button onClick={() => void handleRefreshApis()} disabled={refreshLoading}>
          <RefreshCcwIcon data-icon="inline-start" className={refreshLoading ? 'animate-spin' : undefined} />
          {refreshLoading ? '扫描中...' : '扫描系统 API'}
        </Button>
      </section>

      <Card>
        <CardHeader>
          <CardTitle>筛选条件</CardTitle>
          <CardDescription>按路径、描述和标签筛选接口</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <form className="flex flex-col gap-4" onSubmit={handleSearch}>
            <FieldGroup className="grid gap-3 md:grid-cols-2 xl:grid-cols-[18rem_18rem_18rem_auto] xl:items-end">
              <Field>
                <FieldLabel htmlFor="api-search-path">API 路径</FieldLabel>
                <Input
                  id="api-search-path"
                  placeholder="例如 /api/v1/user/list"
                  value={searchValues.path}
                  onChange={(event) => setSearchValues((current) => ({ ...current, path: event.target.value }))}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="api-search-summary">API 描述</FieldLabel>
                <Input
                  id="api-search-summary"
                  placeholder="例如 查看用户列表"
                  value={searchValues.summary}
                  onChange={(event) => setSearchValues((current) => ({ ...current, summary: event.target.value }))}
                />
              </Field>
              <Field>
                <FieldLabel>API 标签</FieldLabel>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button type="button" variant="outline" className="w-full justify-between font-normal">
                      <span className="truncate text-left">
                        {selectedTags.length > 0
                          ? selectedTags.map((tag) => tag.label).join('、')
                          : '选择标签'}
                      </span>
                      <ChevronsUpDownIcon data-icon="inline-end" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="start" className="w-[var(--radix-dropdown-menu-trigger-width)]">
                    <DropdownMenuLabel>标签筛选</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuGroup>
                      {availableTags.map((tag) => (
                        <DropdownMenuCheckboxItem
                          key={tag.key}
                          checked={searchValues.tags.includes(tag.value)}
                          onSelect={(event) => event.preventDefault()}
                          onCheckedChange={(checked) =>
                            setSearchValues((current) => ({
                              ...current,
                              tags: checked
                                ? [...current.tags, tag.value]
                                : current.tags.filter((value) => value !== tag.value),
                            }))
                          }
                        >
                          <span className="flex-1">{tag.label}</span>
                          {tag.count ? <span className="text-xs text-muted-foreground">{tag.count}</span> : null}
                        </DropdownMenuCheckboxItem>
                      ))}
                    </DropdownMenuGroup>
                  </DropdownMenuContent>
                </DropdownMenu>
              </Field>
            </FieldGroup>
            <div className="flex flex-wrap gap-2">
              <Button type="submit" variant="outline" disabled={loading}>
                <SearchIcon data-icon="inline-start" />
                搜索
              </Button>
              <Button type="button" variant="outline" onClick={handleClearSearch}>
                <XIcon data-icon="inline-start" />
                清空
              </Button>
            </div>
          </form>

          {availableTags.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {(selectedTags.length > 0 ? selectedTags : availableTags).map((tag) => (
                <Badge key={tag.key} variant={selectedTags.length > 0 ? 'secondary' : 'outline'}>
                  {tag.label}
                  {tag.count ? ` (${tag.count})` : ''}
                </Badge>
              ))}
            </div>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>API 列表</CardTitle>
          <CardDescription>共 {total} 条记录</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {apis.length > 0 ? (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>API 路径</TableHead>
                    <TableHead>请求方式</TableHead>
                    <TableHead>API 描述</TableHead>
                    <TableHead>API 标签</TableHead>
                    <TableHead>创建时间</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {apis.map((apiItem) => (
                    <TableRow key={apiItem.id}>
                      <TableCell className="whitespace-normal">
                        <code className="rounded bg-muted px-2 py-1 text-xs">{apiItem.path || '-'}</code>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{apiItem.method}</Badge>
                      </TableCell>
                      <TableCell className="whitespace-normal">{apiItem.summary || '-'}</TableCell>
                      <TableCell>
                        <Badge variant="outline">{apiItem.tags || '未分类'}</Badge>
                      </TableCell>
                      <TableCell>{apiItem.created_at ? new Date(apiItem.created_at).toLocaleString('zh-CN') : '-'}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              <div className="flex flex-col gap-3 border-t pt-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="text-sm text-muted-foreground">
                  第 {currentPage} / {totalPages} 页，共 {total} 条
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-sm text-muted-foreground">每页</span>
                  <Select value={String(pageSize)} onValueChange={(value) => void handlePageSizeChange(value)}>
                    <SelectTrigger className="w-28">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {PAGE_SIZE_OPTIONS.map((value) => (
                          <SelectItem key={value} value={value}>
                            {value} 条
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <Button variant="outline" disabled={currentPage <= 1 || loading} onClick={() => void handlePageChange(currentPage - 1)}>
                    上一页
                  </Button>
                  <Button variant="outline" disabled={currentPage >= totalPages || loading} onClick={() => void handlePageChange(currentPage + 1)}>
                    下一页
                  </Button>
                </div>
              </div>
            </>
          ) : (
            <Empty className="border bg-muted/20 py-10">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <WaypointsIcon />
                </EmptyMedia>
                <EmptyTitle>暂无 API 数据</EmptyTitle>
                <EmptyDescription>请尝试扫描系统 API，或调整筛选条件后重试。</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default ApiManagement
