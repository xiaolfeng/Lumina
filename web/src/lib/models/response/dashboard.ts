export interface TokenStats {
  total: number
  active: number
}

export interface QaStats {
  total: number
  active: number
  expired: number
  deleted: number
  pending_questions: number
}

export interface PreviewStats {
  total: number
  active: number
  files: number
}

export interface RepoWikiStats {
  configs: number
  versions: number
  completed: number
  generating: number
}

export interface RecentPreviewItem {
  id: string
  title: string
  hash: string
  file_count: number
  status: string
  updated_at: string
}

export interface DashboardOverview {
  tokens: TokenStats
  projects: number
  qa: QaStats
  preview: PreviewStats
  repowiki: RepoWikiStats
  recent_previews: RecentPreviewItem[]
}
