import { Link, useLocation } from "@tanstack/react-router";
import {
	ExternalLink,
	FileKey,
	FolderKanban,
	KeyRound,
	Pin,
	LayoutDashboard,
	MessageCircle,
	MessageCircleQuestion,
	MonitorPlay,
	Settings,
	User,
} from "lucide-react";
import { motion } from "motion/react";
import {
	Sidebar,
	SidebarContent,
	SidebarFooter,
	SidebarGroup,
	SidebarGroupContent,
	SidebarGroupLabel,
	SidebarHeader,
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
} from "@lumina/components/ui/sidebar";
import { useAuth } from "#/hooks/useAuth";
import { sidebarItem, sidebarStaggerContainer } from "@lumina/components/motion";

interface NavItem {
	title: string;
	to: string;
	icon: React.ComponentType<{ className?: string }>;
	external?: boolean;
}

interface NavGroup {
	label: string;
	items: NavItem[];
}

const navGroups: NavGroup[] = [
	{
		label: "导航",
		items: [
			{ title: "看板", to: "/console/dashboard", icon: LayoutDashboard },
			{
				title: "交互问答",
				to: "/interact",
				icon: MessageCircle,
				external: true,
			},
		],
	},
		{
			label: "功能",
			items: [
				{ title: "项目管理", to: "/console/project", icon: FolderKanban },
				{ title: "Pin 管理", to: "/console/pin", icon: Pin },
				{ title: "问答管理", to: "/console/qa", icon: MessageCircleQuestion },
				{ title: "预览管理", to: "/console/preview", icon: MonitorPlay },
			],
		},
	{
		label: "系统",
		items: [
			{ title: "令牌管理", to: "/console/apikey", icon: KeyRound },
			{ title: "SSH 密钥", to: "/console/ssh", icon: FileKey },
			{ title: "个人信息", to: "/console/profile", icon: User },
			{ title: "系统设置", to: "/console/settings", icon: Settings },
		],
	},
];

export function AppSidebar() {
	const location = useLocation();
	const { currentUser } = useAuth();

	const user = currentUser.data?.data;
	const displayName = user?.username || "管理员";
	const subtitle = user?.email || "Lumina Console";
	const fallbackInitial = displayName.slice(0, 1) || "管";

	return (
		<Sidebar variant="inset">
			<motion.div
				className="flex h-full flex-col"
				initial="hidden"
				animate="visible"
				variants={sidebarStaggerContainer}
			>
				<SidebarHeader>
					<SidebarMenu>
						<SidebarMenuItem>
							<motion.div variants={sidebarItem}>
								<SidebarMenuButton
									size="lg"
									asChild
									className="hover:bg-link-bg-hover"
								>
									<Link to="/console/dashboard" className="flex items-baseline gap-2 px-1">
										<span className="size-[9px] shrink-0 self-center rounded-full bg-lagoon shadow-[0_0_0_3px_rgba(201,136,58,0.16)]" />
										<span className="display-title text-[17px] font-semibold tracking-[0.5px] text-sea-ink">微明</span>
										<span className="h-[13px] w-px shrink-0 self-center bg-sea-ink-soft/30" />
										<span className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sea-ink-soft">
											Lumina Console
										</span>
									</Link>
								</SidebarMenuButton>
							</motion.div>
						</SidebarMenuItem>
					</SidebarMenu>
				</SidebarHeader>
				<SidebarContent>
					{navGroups.map((group) => (
						<SidebarGroup key={group.label}>
							<SidebarGroupLabel>
								<motion.span variants={sidebarItem}>{group.label}</motion.span>
							</SidebarGroupLabel>
							<SidebarGroupContent>
								<SidebarMenu>
									{group.items.map((item) => {
										const isActive =
											location.pathname === item.to ||
											location.pathname.startsWith(item.to + "/");
										return (
											<motion.div key={item.to} variants={sidebarItem}>
												<SidebarMenuItem>
													<SidebarMenuButton
														asChild
														isActive={isActive}
														tooltip={item.title}
														className={
															isActive
																? "relative font-medium text-sea-ink after:absolute after:-left-4 after:top-1/2 after:h-4 after:w-[3px] after:-translate-y-1/2 after:bg-lagoon"
																: "text-sea-ink-soft hover:bg-link-bg-hover hover:text-sea-ink"
														}
													>
														{item.external ? (
															<a
																href={item.to}
																target="_blank"
																rel="noopener noreferrer"
															>
																<item.icon />
																<span>{item.title}</span>
																<ExternalLink className="ml-auto size-3.5 text-muted-foreground" />
															</a>
														) : (
															<Link to={item.to}>
																<item.icon />
																<span>{item.title}</span>
															</Link>
														)}
													</SidebarMenuButton>
												</SidebarMenuItem>
											</motion.div>
										);
									})}
								</SidebarMenu>
							</SidebarGroupContent>
						</SidebarGroup>
					))}
				</SidebarContent>
				<SidebarFooter className="border-t border-line">
					<SidebarMenu>
						<SidebarMenuItem>
							<motion.div variants={sidebarItem}>
								<SidebarMenuButton
									size="lg"
									className="hover:bg-link-bg-hover"
								>
									<div className="flex size-8 shrink-0 items-center justify-center border border-line text-sm font-bold text-lagoon-deep">
										{fallbackInitial}
									</div>
									<div className="flex flex-col gap-0.5 leading-none">
										<span className="text-sm font-medium text-sea-ink">
											{displayName}
										</span>
										<span className="text-xs text-sea-ink-soft">
											{subtitle}
										</span>
									</div>
								</SidebarMenuButton>
							</motion.div>
						</SidebarMenuItem>
					</SidebarMenu>
				</SidebarFooter>
			</motion.div>
		</Sidebar>
	);
}
