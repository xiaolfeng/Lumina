import { createContext, useContext } from "react";

export interface SessionProgress {
	answered: number;
	remaining: number;
}

export const SidebarOpenContext = createContext<{
	open: boolean;
	setOpen: (v: boolean) => void;
	progress: SessionProgress | null;
	setProgress: (v: SessionProgress | null) => void;
}>({ open: false, setOpen: () => {}, progress: null, setProgress: () => {} });

export function useSidebarOpen() {
	return useContext(SidebarOpenContext);
}
