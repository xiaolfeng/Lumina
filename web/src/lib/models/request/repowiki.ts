export interface RepoWikiConfigListParams {
	page?: number
	size?: number
}

export interface CreateRepoWikiConfigRequest {
	name: string
	repo_url: string
	default_branch?: string
	default_language?: string
	ssh_key_id?: string
	wiki_password?: string
	project_id: string
}

export interface UpdateRepoWikiConfigRequest {
	name?: string
	repo_url?: string
	default_branch?: string
	default_language?: string
	ssh_key_id?: string
	wiki_password?: string
}

export interface AnalyzeRepoWikiRequest {
	language?: string
	branch?: string
}
