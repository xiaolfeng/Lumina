import { createFileRoute } from '@tanstack/react-router'
import { Bot, Send } from 'lucide-react'

import { Badge } from '#/components/ui/badge'
import { Button } from '#/components/ui/button'
import { Card, CardContent } from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '#/components/ui/tooltip'

export const Route = createFileRoute('/_public/faq/')({
  component: FaqPage,
})

function FaqPage() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-8">
      {/* 页面头部 */}
      <div className="mb-8 space-y-3">
        <div className="flex items-center gap-3">
          <h1 className="text-3xl font-bold tracking-tight">交互问答</h1>
          <Badge variant="secondary">开发中</Badge>
        </div>
        <p className="text-muted-foreground">
          与 Lumina 进行智能问答交互
        </p>
      </div>

      {/* 聊天区域 */}
      <div className="space-y-4">
        {/* 消息列表 */}
        <div className="min-h-[400px] space-y-4 rounded-lg border bg-background p-4">
          {/* 机器人欢迎消息 */}
          <div className="flex items-start gap-3">
            <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground">
              <Bot className="size-4" />
            </div>
            <Card className="max-w-[80%] border-0 bg-muted shadow-none">
              <CardContent className="p-3">
                <p className="text-sm">
                  你好！我是 Lumina 助手，有什么可以帮助你的吗？
                </p>
              </CardContent>
            </Card>
          </div>
        </div>

        {/* 输入区域 */}
        <div className="flex items-center gap-2">
          <Input
            disabled
            placeholder="输入您的问题..."
            className="flex-1"
          />
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button disabled size="icon">
                  <Send className="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>即将上线</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>
    </div>
  )
}
