export interface RepoWikiConfigItem {
	id: string
	name: string
	project_id: string
	repo_url: string
	default_branch: string
	default_language: string
	status: string
	has_ssh_key: boolean
	has_password: boolean
	last_accessed_at?: string
	latest_version?: RepoWikiVersionItem
	created_at: string
	updated_at: string
}

export interface RepoWikiVersionItem {
	id: string
	config_id: string
	commit_hash: string
	branch: string
	language: string
	status: string // pending | cloning | scanning | analyzing | assembling | completed | failed | cancelled
	current_stage?: string // scan | dep_extract | pass1 | pass2 | pass3 | pass4 | assemble
	progress_percent: number
	error_msg?: string
	file_count: number
	token_count: number
	duration_ms: number
	started_at?: string
	completed_at?: string
	created_at: string
}

export interface RepoWikiConfigListResponse {
	items: RepoWikiConfigItem[]
	total: number
}

export interface RepoWikiVersionListResponse {
	items: RepoWikiVersionItem[]
	total: number
}
