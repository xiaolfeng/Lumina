export interface CreateSshKeyRequest {
	source: 'generated' | 'imported'
	name: string
	description?: string
	private_key?: string // 仅 imported 时必填
}

export interface UpdateSshKeyRequest {
	name?: string
	description?: string
}

export interface SshKeyListParams {
	page?: number
	size?: number
}
