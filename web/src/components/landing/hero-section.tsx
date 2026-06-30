import { Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowRight, Sparkles } from 'lucide-react'

import { Button } from '#/components/ui/button'
import {
  fadeUp,
  fadeIn,
  heroStagger,
  sectionStagger,
  viewportOnce,
} from '#/lib/motion'

const shellBase =
  'border border-line bg-surface shadow-[0_4px_24px_rgba(42,36,32,0.06)] backdrop-blur-sm'

const kickerBase =
  'text-[11px] font-semibold uppercase tracking-[0.2em] text-lagoon-deep'

export function HeroSection() {
  return (
    <motion.section
      className="relative flex min-h-[90vh] items-center justify-center overflow-hidden px-4 pt-4 pb-20 text-center md:min-h-[92vh] md:pt-6 md:pb-24"
      aria-label="主标题区域"
      initial="hidden"
      animate="visible"
      variants={heroStagger}
    >
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.035]"
        style={{
          backgroundImage:
            'linear-gradient(rgba(42,36,32,0.4) 1px, transparent 1px), linear-gradient(90deg, rgba(42,36,32,0.4) 1px, transparent 1px)',
          backgroundSize: '48px 48px',
          maskImage:
            'radial-gradient(circle at 50% 40%, black, transparent 75%)',
        }}
      />

      <div
        className="pointer-events-none absolute top-1/2 left-1/2 h-[600px] w-[600px] -translate-x-1/2 -translate-y-1/2 rounded-full opacity-50 blur-3xl md:h-[900px] md:w-[900px] md:opacity-40"
        style={{
          background:
            'radial-gradient(circle, var(--hero-a), transparent 60%)',
        }}
      />

      <div
        className="pointer-events-none absolute top-[15%] right-[-5%] h-[350px] w-[350px] rounded-full opacity-30 blur-3xl md:h-[500px] md:w-[500px] md:opacity-25"
        style={{
          background:
            'radial-gradient(circle, var(--hero-b), transparent 55%)',
        }}
      />

      <div
        className="pointer-events-none absolute bottom-[10%] left-[-5%] h-[280px] w-[280px] rounded-full opacity-20 blur-3xl md:h-[400px] md:w-[400px]"
        style={{
          background:
            'radial-gradient(circle, rgba(201,136,58,0.25), transparent 60%)',
        }}
      />

      <div className="relative z-10 mx-auto max-w-3xl">
        <motion.h1
          className="display-title mb-6 text-[clamp(4rem,10vw+1rem,7rem)] font-bold leading-[1.02] tracking-tight md:mb-8"
          variants={fadeUp}
        >
          <span className="bg-gradient-to-br from-sea-ink via-lagoon-deep to-palm bg-clip-text text-transparent">
            Lumina
          </span>
        </motion.h1>

        <motion.p
          className="mb-5 text-xl font-semibold text-lagoon-deep md:mb-6 md:text-[1.65rem]"
          variants={fadeUp}
        >
          烛照幽微，知常曰明
        </motion.p>

        <motion.p
          className="mx-auto mb-14 max-w-lg text-base leading-relaxed text-sea-ink-soft md:mb-16 md:max-w-xl md:text-lg"
          variants={fadeUp}
        >
          万象隐于幽微，常理没于流转。微明不求如烈日灼目，只愿燃一寸静烛：为乱麻梳理脉络，使隐晦昭然；将瞬息沉淀为常识，让过往不再是流沙。
        </motion.p>

        <motion.div
          className="flex flex-col items-center justify-center gap-4 sm:flex-row"
          variants={fadeUp}
        >
          <Button
            asChild
            size="lg"
            className="group relative h-12 overflow-hidden rounded-full bg-gradient-to-r from-lagoon to-palm px-8 text-base font-semibold !text-white shadow-lg shadow-lagoon/25 transition-shadow duration-300 hover:shadow-xl hover:shadow-lagoon/40"
          >
            <Link to="/auth/login" aria-label="开始使用 Lumina">
              开始使用
              <ArrowRight
                className="ml-2 h-4 w-4 transition-transform duration-300 group-hover:translate-x-1"
                aria-hidden
              />
            </Link>
          </Button>

          <Button
            asChild
            variant="outline"
            size="lg"
            className="h-12 rounded-full border-line px-8 text-base font-medium text-sea-ink-soft transition-colors duration-300 hover:border-lagoon/30 hover:bg-lagoon/5 hover:text-sea-ink"
          >
            <a href="#features" aria-label="了解更多关于 Lumina 的功能">
              了解更多
            </a>
          </Button>
        </motion.div>

        <motion.div
          className="mx-auto mt-20 h-px w-28 bg-gradient-to-r from-transparent via-lagoon/40 to-transparent md:mt-24"
          variants={fadeIn}
        />
      </div>
    </motion.section>
  )
}

export function CtaSection() {
  return (
    <section
      className="page-wrap px-4 py-24 text-center"
      aria-label="行动号召"
    >
      <motion.div
        className={`${shellBase} mx-auto max-w-xl rounded-2xl px-8 py-14`}
        initial="hidden"
        whileInView="visible"
        viewport={viewportOnce}
        variants={sectionStagger}
      >
        <motion.div variants={fadeUp}>
          <Sparkles
            className="mx-auto mb-5 h-10 w-10 text-lagoon"
            aria-hidden
          />
        </motion.div>
        <motion.h2
          className="display-title mb-4 text-2xl font-bold text-sea-ink sm:text-3xl"
          variants={fadeUp}
        >
          准备好让 AI 真正理解你的代码了吗？
        </motion.h2>
        <motion.p
          className="mb-8 text-sm text-sea-ink-soft"
          variants={fadeUp}
        >
          立即开始使用 Lumina，开启 AI 深度代码认知之旅。
        </motion.p>
        <motion.div variants={fadeUp}>
          <Button asChild size="lg">
            <Link to="/auth/login" aria-label="立即开始使用 Lumina">
              立即开始
              <ArrowRight className="ml-2 h-4 w-4" aria-hidden />
            </Link>
          </Button>
        </motion.div>
      </motion.div>
    </section>
  )
}

export { kickerBase }
