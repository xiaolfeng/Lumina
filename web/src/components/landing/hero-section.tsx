import { Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowRight } from 'lucide-react'

import { Button } from '@lumina/components/ui/button'
import { fadeUp, heroStagger } from '@lumina/components/motion'

export function HeroSection() {
  return (
    <motion.section
      className="mx-auto max-w-3xl px-6 pt-16 pb-20 text-center"
      initial="hidden"
      animate="visible"
      variants={heroStagger}
      aria-label="主标题区域"
    >
      <motion.p
        variants={fadeUp}
        className="text-[11px] font-semibold uppercase tracking-[0.32em] text-lagoon-deep"
      >
        烛照幽微 · 知识中枢
      </motion.p>

      <motion.h1
        variants={fadeUp}
        className="display-title mt-7 text-5xl leading-[1.1] font-medium tracking-tight text-sea-ink sm:text-6xl"
      >
        赋予 AI 深度代码认知
        <br />
        与<em className="text-lagoon italic">长期记忆</em>
      </motion.h1>

      <motion.p
        variants={fadeUp}
        className="mx-auto mt-7 max-w-xl text-base leading-relaxed text-sea-ink-soft"
      >
        微明 Lumina 是一套面向 AI Agent
        的知识中枢，通过 RepoWiki、Memory、Q&A 与 Pin
        四大模块，为模型注入可检索的代码认知与跨会话记忆。
      </motion.p>

      <motion.div
        variants={fadeUp}
        className="mt-10 flex items-center justify-center gap-4"
      >
        <Button
          asChild
          size="lg"
          className="bg-sea-ink text-foam hover:bg-lagoon-deep"
        >
          <Link to="/auth/login" aria-label="开始使用 Lumina">
            开始使用
          </Link>
        </Button>
        <Button
          asChild
          size="lg"
          variant="outline"
          className="text-sea-ink hover:border-sea-ink hover:bg-transparent"
        >
          <a href="#modules" aria-label="查看核心模块">
            查看文档
            <ArrowRight className="ml-1.5 h-4 w-4" aria-hidden />
          </a>
        </Button>
      </motion.div>
    </motion.section>
  )
}
