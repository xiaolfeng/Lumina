import type { Question, Session } from "./types";

export const mockSessions: Session[] = [
	{
		id: "session-1",
		title: "项目架构确认",
		agent: "RepoWiki Agent",
		type: "temporary",
		status: "active",
		updatedAt: "2 分钟前",
		expiresAt: "",
		owner: "筱锋",
		onlineDevices: 2,
		questions: [
			{
				id: "q1",
				sessionId: "session-1",
				content: "检测到你的项目使用了多层路由结构，请确认主要使用的路由模式：",
				description: "路由模式决定了页面与 URL 的映射方式，影响项目整体架构。",
				type: "select",
				options: [
					{
						id: "opt-1",
						label: "文件路由",
						description:
							"基于文件系统自动生成路由，如 TanStack Router、Next.js App Router",
					},
					{
						id: "opt-2",
						label: "代码路由",
						description: "通过代码手动注册路由表，灵活但维护成本较高",
					},
					{
						id: "opt-3",
						label: "混合模式",
						description: "文件路由为主，关键路由手动覆盖",
					},
				],
				groupLabel: "路由架构",
				status: "answered",
				answered: true,
				answer: "文件路由",
				createdAt: "",
			},
			{
				id: "q2",
				sessionId: "session-1",
				content: "请描述你期望的 API 错误处理策略：",
				description:
					"错误处理策略决定了异常信息的格式、日志记录方式以及客户端的响应行为。",
				type: "text",
				groupLabel: "错误处理",
				status: "pending",
				answered: false,
				createdAt: "",
			},
			{
				id: "q3",
				sessionId: "session-1",
				content: "是否需要自动生成 Swagger 文档？",
				description:
					"启用后将通过 swaggo/swag 自动扫描 Handler 注解生成 OpenAPI 规范文档。",
				type: "boolean",
				groupLabel: "文档策略",
				status: "pending",
				answered: false,
				createdAt: "",
			},
		],
	},
	{
		id: "session-2",
		title: "编码风格偏好",
		agent: "Memory Agent",
		type: "temporary",
		status: "active",
		updatedAt: "1 小时前",
		expiresAt: "",
		owner: "筱锋",
		onlineDevices: 1,
		questions: [
			{
				id: "q4",
				sessionId: "session-2",
				content: "你偏好哪种代码注释风格？",
				description: "注释风格影响团队协作效率和 IDE 文档生成质量。",
				type: "select",
				options: [
					{
						id: "opt-4",
						label: "行尾注释",
						description: "在每行代码末尾追加简短说明，紧凑且就近",
					},
					{
						id: "opt-5",
						label: "上方块注释",
						description:
							"在函数或逻辑块上方用多行注释详细说明，适合复杂逻辑",
					},
					{
						id: "opt-6",
						label: "JSDoc 风格",
						description: "使用标准 JSDoc 格式，可自动生成 API 文档",
					},
				],
				groupLabel: "注释风格",
				status: "answered",
				answered: true,
				answer: "行尾注释",
				createdAt: "",
			},
		],
	},
	{
		id: "session-3",
		title: "部署环境配置",
		agent: "RepoWiki Agent",
		type: "temporary",
		status: "active",
		updatedAt: "昨天",
		expiresAt: "",
		owner: "筱锋",
		onlineDevices: 1,
		questions: [
			{
				id: "q5",
				sessionId: "session-3",
				content: "你的项目目标部署环境是什么？",
				description: "部署环境决定了构建产物形式、CI/CD 流程以及运维复杂度。",
				type: "select",
				options: [
					{
						id: "opt-7",
						label: "Docker",
						description: "容器化部署，轻量且一致性好",
					},
					{
						id: "opt-8",
						label: "Kubernetes",
						description: "容器编排，适合大规模微服务集群",
					},
					{
						id: "opt-9",
						label: "裸机部署",
						description: "直接运行二进制，最简单直接",
					},
					{
						id: "opt-10",
						label: "Serverless",
						description: "按需弹性伸缩，无需管理基础设施",
					},
				],
				groupLabel: "部署",
				status: "answered",
				answered: true,
				answer: "Docker",
				createdAt: "",
			},
			{
				id: "q6",
				sessionId: "session-3",
				content: "是否需要 CI/CD 自动部署？",
				description:
					"自动化流水线可在代码合并后自动完成测试、构建和部署流程。",
				type: "boolean",
				groupLabel: "部署",
				status: "answered",
				answered: true,
				answer: "是",
				createdAt: "",
			},
		],
	},
];

export const debugMotionQuestion: Question = {
	id: "debug-motion",
	sessionId: "debug-session",
	content: "请选择你偏好的组件载入动画方案：",
	type: "select",
	options: [
		{
			id: "dm-opt-1",
			label: "从上到下滑入",
			description: "组件从上方滑入视口，适合列表项依次出现",
		},
		{
			id: "dm-opt-2",
			label: "淡入",
			description: "组件透明度从 0 渐变到 1，最柔和的入场方式",
		},
		{
			id: "dm-opt-3",
			label: "从左到右滑入",
			description: "组件从左侧滑入，适合导航类组件",
		},
		{
			id: "dm-opt-4",
			label: "缩放弹入",
			description: "组件从小到大弹性出现，带轻微弹跳效果",
		},
	],
	allowOther: true,
	groupLabel: "动画方案",
	status: "pending",
	answered: false,
	createdAt: "",
};

export const debugMultiQuestion: Question = {
	id: "debug-multi-codereview",
	sessionId: "debug-session",
	content: "你认为代码审查应关注哪些方面？（可多选）",
	description:
		"选择你认为最重要的审查维度，Agent 将根据你的偏好调整审查策略的优先级和详细程度。",
	type: "multi-select",
	options: [
		{
			id: "dm-opt-5",
			label: "类型安全",
			description: "检查类型定义完整性、泛型约束、`any` 使用等",
		},
		{
			id: "dm-opt-6",
			label: "性能影响",
			description: "关注渲染性能、内存泄漏、不必要的重渲染和大列表优化",
		},
		{
			id: "dm-opt-7",
			label: "错误处理",
			description: "异常边界、错误恢复、空值防护和用户反馈",
		},
		{
			id: "dm-opt-8",
			label: "可访问性",
			description: "ARIA 属性、键盘导航、颜色对比度和语义化 HTML",
		},
		{
			id: "dm-opt-9",
			label: "测试覆盖",
			description: "单元测试、集成测试的完整性和边界条件覆盖",
		},
	],
	allowOther: true,
	groupLabel: "代码审查",
	status: "pending",
	answered: false,
	createdAt: "",
};

export const debugBooleanQuestion: Question = {
	id: "debug-boolean",
	sessionId: "debug-session",
	content: "是否启用交互式代码预览功能？",
	description:
		"开启后，用户在回答问题时可以实时查看代码片段的渲染效果，增强理解与决策效率。",
	type: "boolean",
	groupLabel: "Debug 模拟",
	status: "pending",
	answered: false,
	createdAt: "",
};

export const debugTextQuestion: Question = {
	id: "debug-text",
	sessionId: "debug-session",
	content: "请描述你对项目日志策略的期望：",
	description:
		"日志策略包括日志级别、输出格式、持久化方式等，直接影响问题排查和系统监控体验。",
	type: "text",
	groupLabel: "Debug 模拟",
	status: "pending",
	answered: false,
	createdAt: "",
};

export const mockMarkdownContent = `# 项目架构分析报告

## 路由结构

检测到项目使用 **TanStack Start** 文件路由模式：

\`\`\`text
routes/
├── __root.tsx          # 根布局
├── _public.tsx         # 公开页面布局
├── _public/
│   ├── index.tsx       # 首页
│   └── start.tsx       # 快速开始
├── auth.tsx            # 认证布局
├── console.tsx         # 控制台布局
└── console/
    ├── dashboard.tsx   # 仪表盘
    └── ...
\`\`\`

## 当前决策

| 项目 | 选择 |
|------|------|
| 路由模式 | 文件路由 |
| 错误处理 | _待确认_ |
| 文档策略 | _待确认_ |

> Agent 正在等待你对剩余问题的回答，完成后将生成完整的架构报告。
`;
