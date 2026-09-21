import React, {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import type { LpwAnnotation } from "./types";

export interface RegisteredAnnotation {
  nodeId: string;
  annotation: LpwAnnotation;
  label?: string;
  anchorText?: string;
}

interface AnnotationContextValue {
  annotations: RegisteredAnnotation[];
  registerAnnotation: (item: RegisteredAnnotation) => () => void;
  hoveredAnnotationId: string | null;
  setHoveredAnnotationId: (id: string | null) => void;
  activeAnnotationId: string | null;
  setActiveAnnotationId: (id: string | null) => void;
  scrollToNode: (nodeId: string) => void;
}

const AnnotationContext = createContext<AnnotationContextValue | null>(null);

export const AnnotationProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [annotationsMap, setAnnotationsMap] = useState<
    Map<string, RegisteredAnnotation>
  >(() => new Map());
  const [hoveredAnnotationId, setHoveredAnnotationId] = useState<string | null>(
    null,
  );
  const [activeAnnotationId, setActiveAnnotationId] = useState<string | null>(
    null,
  );

  const registerAnnotation = useCallback((item: RegisteredAnnotation) => {
    setAnnotationsMap((prev) => {
      const existing = prev.get(item.nodeId);
      if (
        existing &&
        existing.annotation === item.annotation &&
        existing.label === item.label &&
        existing.anchorText === item.anchorText
      ) {
        return prev;
      }
      const next = new Map(prev);
      next.set(item.nodeId, item);
      return next;
    });

    return () => {
      setAnnotationsMap((prev) => {
        if (!prev.has(item.nodeId)) return prev;
        const next = new Map(prev);
        next.delete(item.nodeId);
        return next;
      });
    };
  }, []);

  const scrollToNode = useCallback((nodeId: string) => {
    setActiveAnnotationId(nodeId);
    const el =
      document.getElementById(`frame-${nodeId}`) ??
      document.getElementById(nodeId);
    if (el) {
      el.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, []);

  const annotations = useMemo(
    () => Array.from(annotationsMap.values()),
    [annotationsMap],
  );

  const value = useMemo(
    () => ({
      annotations,
      registerAnnotation,
      hoveredAnnotationId,
      setHoveredAnnotationId,
      activeAnnotationId,
      setActiveAnnotationId,
      scrollToNode,
    }),
    [
      annotations,
      registerAnnotation,
      hoveredAnnotationId,
      activeAnnotationId,
      scrollToNode,
    ],
  );

  return (
    <AnnotationContext.Provider value={value}>
      {children}
    </AnnotationContext.Provider>
  );
};

export function useAnnotationContext(): AnnotationContextValue | null {
  return useContext(AnnotationContext);
}
