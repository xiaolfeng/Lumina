import { useState } from 'react'
import { Check, ChevronsUpDown } from 'lucide-react'
import { Button } from '@lumina/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@lumina/components/ui/dialog'
import { Input } from '@lumina/components/ui/input'
import { Label } from '@lumina/components/ui/label'
import { Textarea } from '@lumina/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@lumina/components/ui/select'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@lumina/components/ui/popover'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@lumina/components/ui/command'
import { cn } from '#/lib/utils'
import { useCreatePin } from '#/hooks/usePin'
import { useProjectList } from '#/hooks/useProject'

interface CreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const CATEGORY_OPTIONS = [
  { value: 'notice', label: '注意事项' },
  { value: 'dependency', label: '依赖约束' },
  { value: 'api_change', label: '接口变更' },
  { value: 'other', label: '其他' },
] as const

const PRIORITY_OPTIONS = [
  { value: 'high', label: '高' },
  { value: 'medium', label: '中' },
  { value: 'low', label: '低' },
] as const

function ProjectCombobox({
  value,
  onChange,
  placeholder,
}: {
  value: string
  onChange: (id: string) => void
  placeholder?: string
}) {
  const [open, setOpen] = useState(false)
  const { data } = useProjectList()
  const projects = data?.data?.items ?? []
  const selectedProject = projects.find((p) => p.id === value)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between"
        >
          {selectedProject ? selectedProject.name : (placeholder ?? '选择项目...')}
          <ChevronsUpDown className="ml-2 size-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-full p-0" align="start">
        <Command>
          <CommandInput placeholder="搜索项目..." />
          <CommandList>
            <CommandEmpty>未找到项目</CommandEmpty>
            <CommandGroup>
              {projects.map((project) => (
                <CommandItem
                  key={project.id}
                  value={`${project.name} ${project.alias_name ?? ''}`}
                  onSelect={() => {
                    onChange(project.id)
                    setOpen(false)
                  }}
                >
                  <Check
                    className={cn(
                      'mr-2 size-4',
                      value === project.id ? 'opacity-100' : 'opacity-0',
                    )}
                  />
                  {project.name}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}

export function CreatePinDialog({ open, onOpenChange }: CreateDialogProps) {
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [category, setCategory] = useState('notice')
  const [priority, setPriority] = useState('')
  const [fromProjectId, setFromProjectId] = useState('')
  const [toProjectId, setToProjectId] = useState('')

  const createMutation = useCreatePin()

  const handleSubmit = () => {
    if (!title.trim() || !content.trim() || !priority || !toProjectId) return
    createMutation.mutate(
      {
        title: title.trim(),
        content: content.trim(),
        category,
        priority,
        from_project_id: fromProjectId || undefined,
        to_project_id: toProjectId,
      },
      { onSuccess: () => handleClose() },
    )
  }

  const handleClose = () => {
    setTitle('')
    setContent('')
    setCategory('notice')
    setPriority('')
    setFromProjectId('')
    setToProjectId('')
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>创建 Pin 约束</DialogTitle>
          <DialogDescription>
            创建一条新的 Pin 约束，用于跨项目传递依赖信息。
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4 max-h-[60vh] overflow-y-auto">
          {/* 标题 */}
          <div className="grid gap-2">
            <Label htmlFor="pin-title">标题 *</Label>
            <Input
              id="pin-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="输入约束标题"
            />
          </div>

          {/* 内容 */}
          <div className="grid gap-2">
            <Label htmlFor="pin-content">内容 *</Label>
            <Textarea
              id="pin-content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="输入约束详细内容"
              className="min-h-[80px]"
            />
          </div>

          {/* 分类 + 优先级：同一行 */}
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label>分类</Label>
              <Select value={category} onValueChange={setCategory}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="选择分类" />
                </SelectTrigger>
                <SelectContent>
                  {CATEGORY_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label>优先级 *</Label>
              <Select value={priority} onValueChange={setPriority}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="选择优先级" />
                </SelectTrigger>
                <SelectContent>
                  {PRIORITY_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* 来源项目（可选） */}
          <div className="grid gap-2">
            <Label>来源项目</Label>
            <ProjectCombobox
              value={fromProjectId}
              onChange={setFromProjectId}
              placeholder="选择来源项目（可选）"
            />
          </div>

          {/* 目标项目（必填） */}
          <div className="grid gap-2">
            <Label>目标项目 *</Label>
            <ProjectCombobox
              value={toProjectId}
              onChange={setToProjectId}
              placeholder="选择目标项目"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={
              !title.trim() ||
              !content.trim() ||
              !priority ||
              !toProjectId ||
              createMutation.isPending
            }
          >
            {createMutation.isPending ? '创建中...' : '创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
