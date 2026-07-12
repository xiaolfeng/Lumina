export interface SshKeyItem {
	id: string
	name: string
	description: string
	key_type: string // 'ed25519' | 'rsa'
	public_key: string
	fingerprint: string
	source: 'generated' | 'imported'
	created_at: string
	updated_at: string
}

export interface CreateSshKeyResponse extends SshKeyItem {
	public_key_download_url: string
}

export interface SshKeyListResponse {
	total: number
	items: SshKeyItem[]
}
