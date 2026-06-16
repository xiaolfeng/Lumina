export type QuestionType =
	| "select"
	| "multi-select"
	| "text"
	| "boolean"
	| "code"
	| "image"
	| "file"
	| "diff"
	| "plan"
	| "options"
	| "review"
	| "slider"
	| "rank"
	| "rate";

export interface OptionItem {
	id: string;
	label: string;
	description?: string;
}

export interface Question {
	id: string;
	sessionId: string;
	content: string;
	description?: string;
	type: QuestionType;
	options?: OptionItem[];
	allowOther?: boolean;
	groupLabel: string;
	batch?: { group_id: string; sequence: number; total: number };
	config?: Record<string, any>;
	status: 'pending' | 'answered' | 'skipped' | 'cancelled';
	answered: boolean;
	answer?: any;
	supplements?: SupplementItem[];
	supplement?: boolean; // 是否携带补充内容
	media?: any;
	createdAt: string;
	answeredAt?: string;
}

export interface Session {
  id: string;
  hash: string; // 会话唯一哈希标识
  title: string;
	agent: string;
	type: 'temporary' | 'permanent';
	status: 'active' | 'expired' | 'deleted';
	updatedAt: string;
	expiresAt: string;
	owner: string;
	onlineDevices: number;
	questions: Question[];
}

export interface SupplementItem {
	id: string;
	target_type: 'question' | 'option';
	target_id: string;
	content_type: 'markdown' | 'html';
	content: string;
	created_at: string;
	updated_at: string;
}
