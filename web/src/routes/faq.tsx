import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft, Bot, Send, Sparkles } from 'lucide-react'

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

export const Route = createFileRoute('/faq')({
  component: FaqPage,
})

function FaqPage() {
  return (
    <div className="min-h-screen bg-background">
      {/* 顶部导航栏 */}
      <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="mx-auto flex h-14 max-w-3xl items-center gap-3 px-4">
          <Link
            to="/"
            className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            <ArrowLeft className="size-4" />
            返回首页
          </Link>
          <div className="h-4 w-px bg-border" />
          <div className="flex items-center gap-2">
            <Sparkles className="size-4 text-primary" />
            <span className="text-sm font-medium">微明 Lumina</span>
          </div>
        </div>
      </header>

      {/* 页面内容 */}
      <div className="mx-auto max-w-3xl px-4 py-8">
        <div className="mb-8 space-y-3">
          <div className="flex items-center gap-3">
            <h1 className="text-3xl font-bold tracking-tight">交互问答</h1>
            <Badge variant="secondary">开发中</Badge>
          </div>
          <p className="text-muted-foreground">
            与 Lumina 进行智能问答交互
          </p>
        </div>

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
    </div>
  )
}
