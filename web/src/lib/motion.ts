import type { Variants } from 'motion/react'

/** 共享缓动 — 与登录页完全一致 */
export const ease: [number, number, number, number] = [0.16, 1, 0.3, 1]

/** 交错子元素入场（供子页面自驱动使用） */
export const staggerContainer: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.06,
      delayChildren: 0.05,
    },
  },
}

/** 交错子元素入场（y 方向轻滑） */
export const staggerItem: Variants = {
  hidden: { opacity: 0, y: 12 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4, ease },
  },
}
