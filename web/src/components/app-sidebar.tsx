import { Link, useLocation } from "@tanstack/react-router";
import {
	ExternalLink,
	FolderKanban,
	KeyRound,
	LayoutDashboard,
	MessageCircle,
	MessageCircleQuestion,
	Settings,
	Sparkles,
} from "lucide-react";
import { motion } from "motion/react";
import { Avatar, AvatarFallback } from "#/components/ui/avatar";
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
} from "#/components/ui/sidebar";
import { useAuth } from "#/hooks/useAuth";
import { sidebarItem, sidebarStaggerContainer } from "#/lib/motion";

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
			{ title: "问答管理", to: "/console/qa", icon: MessageCircleQuestion },
		],
	},
	{
		label: "系统",
		items: [
			{ title: "令牌管理", to: "/console/apikey", icon: KeyRound },
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
									<Link to="/console/dashboard">
										<div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-lagoon text-foam shadow-sm shadow-hero-a">
											<Sparkles className="size-4" />
										</div>
										<div className="flex flex-col gap-0.5 leading-none">
											<span className="font-semibold text-sea-ink">
												微明 Lumina
											</span>
											<span className="text-xs text-sea-ink-soft">
												管理后台
											</span>
										</div>
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
																? "bg-chip-bg text-lagoon border border-chip-line font-medium"
																: "hover:bg-link-bg-hover"
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
									<Avatar className="size-8 rounded-lg">
										<AvatarFallback className="rounded-lg bg-(--accent) text-lagoon text-sm font-medium">
											{fallbackInitial}
										</AvatarFallback>
									</Avatar>
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
