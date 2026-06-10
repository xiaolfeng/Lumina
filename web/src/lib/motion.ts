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

/** 交错子元素入场 — Main 区域从右向左滑入 */
export const staggerItem: Variants = {
  hidden: { opacity: 0, x: 20 },
  visible: {
    opacity: 1,
    x: 0,
    transition: { duration: 0.4, ease },
  },
}

/** 交错子元素入场 — 从左向右滑入（仅标题行使用） */
export const staggerItemLeft: Variants = {
  hidden: { opacity: 0, x: -12 },
  visible: {
    opacity: 1,
    x: 0,
    transition: { duration: 0.4, ease },
  },
}

/** Sidebar 整体交错容器 — header → 各分组 → footer 按顺序载入 */
export const sidebarStaggerContainer: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.08,
      delayChildren: 0.15,
    },
  },
}

/** Sidebar 子项从左向右滑入 */
export const sidebarItem: Variants = {
  hidden: { opacity: 0, x: -16 },
  visible: {
    opacity: 1,
    x: 0,
    transition: { duration: 0.35, ease },
  },
}
