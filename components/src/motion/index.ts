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

/**
 * Sidebar 整块淡入 — 用于频繁重渲的侧边栏（如 Wiki Reader）。
 * 不做逐项交错，避免每次路由切换重放序列动画。
 */
export const sidebarBlockFade: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { duration: 0.25, ease },
  },
}

/**
 * Main 区域滑入 — 右侧内容切换时从右向左滑入。
 * 配合 AnimatePresence mode="wait" 使用。
 */
export const mainSlideIn: Variants = {
  hidden: { opacity: 0, x: 16 },
  visible: {
    opacity: 1,
    x: 0,
    transition: { duration: 0.3, ease },
  },
  exit: {
    opacity: 0,
    x: -8,
    transition: { duration: 0.15, ease },
  },
}

/* ─── Landing 页面动画变体 ─────────────────────────────── */

/** 向上淡入 — 标题、描述、按钮等通用入场 */
export const fadeUp: Variants = {
  hidden: { opacity: 0, y: 18 },
  visible: { opacity: 1, y: 0 },
}

/** 纯淡入 — 装饰线等无位移元素 */
export const fadeIn: Variants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1 },
}

/** 缩放淡入 — 代码块、卡片等强调元素 */
export const scaleIn: Variants = {
  hidden: { opacity: 0, scale: 0.92 },
  visible: { opacity: 1, scale: 1 },
}

/** Hero 区域交错容器 — staggerChildren 0.12, delayChildren 0.08 */
export const heroStagger: Variants = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.12, delayChildren: 0.08 },
  },
}

/** Section 区域交错容器 — staggerChildren 0.1 */
export const sectionStagger: Variants = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.1 },
  },
}

/** whileInView 视口配置 — 一次性触发，提前 80px 进入 */
export const viewportOnce = { once: true, margin: '-80px' } as const
