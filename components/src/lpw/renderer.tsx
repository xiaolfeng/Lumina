import React from "react";
import { AnnotationFrame } from "./annotation";
import { hasGroup } from "./contract";
import { DiagnosticCard } from "./diagnostics";
import type { LpwDiagnostic } from "./diagnostics";
import { lpwRegistry } from "./registry";
import { ensureRegistered } from "./register-all";
import { childLocation, rootLocation } from "./render-location";
import type { LpwRenderLocation } from "./render-location";
import type {
  LpwBlock,
  LpwContainer,
  LpwContainerVariantContract,
  LpwContentNode,
  LpwLayout,
  LpwLayoutChild,
  LpwNodeKind,
} from "./types";

interface NodeErrorBoundaryProps {
  diagnostic: Omit<LpwDiagnostic, "reason">;
  children: React.ReactNode;
}

interface NodeErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

class NodeErrorBoundary extends React.Component<
  NodeErrorBoundaryProps,
  NodeErrorBoundaryState
> {
  state: NodeErrorBoundaryState = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): NodeErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    console.error("LPW Node Render Error:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <DiagnosticCard
          diagnostic={{
            ...this.props.diagnostic,
            code: "RENDER_EXCEPTION",
            reason: this.state.error?.message ?? "未知渲染异常",
            stack: this.state.error?.stack,
          }}
        />
      );
    }
    return this.props.children;
  }
}

export interface LpwNodeRendererProps {
  node: LpwContentNode;
  location?: LpwRenderLocation;
  parentKind?: LpwNodeKind;
}

export const LpwNodeRenderer: React.FC<LpwNodeRendererProps> = ({
  node,
  location,
  parentKind,
}) => {
  ensureRegistered();
  const loc = location ?? rootLocation();
  const effectiveKind = (node as { kind?: LpwNodeKind }).kind || "block";

  // 1. 父层级检查
  if (parentKind === "layout" && effectiveKind === "layout") {
    return (
      <DiagnosticCard
        diagnostic={{
          code: "INVALID_LAYOUT_SLOT",
          location: loc,
          nodeId: node.id,
          nodeType: node.type,
          reason: "layout 不允许直接包含 layout 子节点",
          suggestion: "layout 只能直接包含 container 或 block",
        }}
      />
    );
  }

  if (parentKind === "container" && effectiveKind !== "block") {
    return (
      <DiagnosticCard
        diagnostic={{
          code: "INVALID_CONTAINER_CHILD",
          location: loc,
          nodeId: node.id,
          nodeType: node.type,
          reason: `container 不允许包含 kind=${node.kind} 的子节点`,
          suggestion: "container 只能直接包含 block 节点",
        }}
      />
    );
  }

  // 2. 注册表查询
  const entry = lpwRegistry.get(effectiveKind, node.type);
  if (!entry) {
    return (
      <DiagnosticCard
        diagnostic={{
          code: "UNKNOWN_BLOCK",
          location: loc,
          nodeId: node.id,
          nodeType: node.type,
          reason: `未注册的 ${node.kind} 类型: ${node.type}`,
          suggestion: "请确认组件类型是否已注册",
        }}
      />
    );
  }

  // 3. 错误边界包装
  const boundaryDiagnostic: Omit<LpwDiagnostic, "reason"> = {
    code: "RENDER_EXCEPTION",
    location: loc,
    nodeId: node.id,
    nodeType: node.type,
    componentName: entry.displayName,
    propsSummary: node.props,
  };

  const renderComponent = () => {
    if (entry.kind === "layout") {
      const Comp = entry.Component;
      const layout = node as LpwLayout;
      return (
        <Comp
          nodeId={layout.id}
          props={layout.props}
          location={loc}
          children={layout.children}
        />
      );
    }

    if (entry.kind === "container") {
      const Comp = entry.Component;
      const container = node as LpwContainer;
      return (
        <Comp
          nodeId={container.id}
          props={container.props}
          location={loc}
          children={container.children}
        />
      );
    }

    const Comp = entry.Component;
    const block = node as LpwBlock;
    const content = (
      <Comp
        nodeId={block.id}
        props={block.props}
        location={loc}
        annotation={block.annotation}
      />
    );
    return (
      <AnnotationFrame nodeId={block.id} annotation={block.annotation}>
        {content}
      </AnnotationFrame>
    );
  };

  return (
    <NodeErrorBoundary diagnostic={boundaryDiagnostic}>
      {renderComponent()}
    </NodeErrorBoundary>
  );
};

// 兼容别名导出
export const LpwBlockRenderer = LpwNodeRenderer;

export function renderLayoutChildren(
  children: LpwLayoutChild[],
  parent: LpwRenderLocation | undefined,
): React.ReactNode {
  const pLoc = parent || rootLocation();
  return children.map((child, idx) => {
    const loc = childLocation(pLoc, idx, child.id);
    return (
      <LpwNodeRenderer
        key={child.id}
        node={child}
        location={loc}
        parentKind="layout"
      />
    );
  });
}

export function renderContainerBlocks(
  children: LpwBlock[],
  parent: LpwRenderLocation | undefined,
  contract: LpwContainerVariantContract | undefined,
  containerLabel: string,
): React.ReactNode {
  const pLoc = parent || rootLocation();
  return children.map((child, idx) => {
    const loc = childLocation(pLoc, idx, child.id);
    const rawKind = (child as unknown as { kind?: string }).kind;

    // 检查 child 的 kind
    if (rawKind && rawKind !== "block") {
      return (
        <DiagnosticCard
          key={child.id || idx}
          diagnostic={{
            code: "INVALID_CONTAINER_CHILD",
            location: loc,
            nodeId: child.id,
            nodeType: child.type,
            reason: `${containerLabel} 不允许包含 kind=${rawKind} 的子节点；container 只能包含 block`,
            suggestion: "请将该节点移出或更改为 block",
          }}
        />
      );
    }

    // 契约检查
    if (contract) {
      if (contract.DeniedTypes?.includes(child.type)) {
        return (
          <DiagnosticCard
            key={child.id || idx}
            diagnostic={{
              code: "INVALID_CONTAINER_CHILD",
              location: loc,
              nodeId: child.id,
              nodeType: child.type,
              reason: `${containerLabel} 显式禁止包含 ${child.type} 类型的子块`,
              suggestion: `该变体接受的组为: ${contract.AllowedGroups?.join(", ") || "指定列表"}`,
            }}
          />
        );
      }

      let allowed = Boolean(contract.AllowedTypes?.includes(child.type));
      if (!allowed && contract.AllowedGroups) {
        for (const g of contract.AllowedGroups) {
          if (hasGroup(child.type, g)) {
            allowed = true;
            break;
          }
        }
      }

      if (!allowed) {
        return (
          <DiagnosticCard
            key={child.id || idx}
            diagnostic={{
              code: "INVALID_CONTAINER_CHILD",
              location: loc,
              nodeId: child.id,
              nodeType: child.type,
              reason: `${containerLabel} 不允许包含 ${child.type} 类型的子块`,
              suggestion: `该变体接受的策略组: ${contract.AllowedGroups?.join(", ") || contract.AllowedTypes?.join(", ") || "无"}`,
            }}
          />
        );
      }
    }

    return (
      <LpwNodeRenderer
        key={child.id}
        node={child}
        location={loc}
        parentKind="container"
      />
    );
  });
}
