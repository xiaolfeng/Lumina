import type React from "react";
import { useState } from "react";
import { containerVariants } from "../contract";
import { renderContainerBlocks } from "../renderer";
import type { LpwContainerSlotProps, LpwTabsContainerProps } from "../types";

export const TabsContainer: React.FC<
  LpwContainerSlotProps<LpwTabsContainerProps>
> = ({ nodeId, blockId, props, location, children, childrenBlocks }) => {
  const actualId = nodeId || blockId || "";
  const actualChildren = children || childrenBlocks || [];
  const { items = [], defaultKey, title } = props;
  const variant = props.variant;
  const contract = variant ? containerVariants.tabs[variant] : undefined;

  const initialKey =
    defaultKey && items.some((it) => it.key === defaultKey)
      ? defaultKey
      : (items[0]?.key ?? "");

  const [activeKey, setActiveKey] = useState<string>(initialKey);

  const resolvedKey = items.some((it) => it.key === activeKey)
    ? activeKey
    : (items[0]?.key ?? "");
  const activeIndex = items.findIndex((it) => it.key === resolvedKey);
  const activeChild =
    activeIndex >= 0 ? actualChildren[activeIndex] : undefined;

  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (e.key !== "ArrowLeft" && e.key !== "ArrowRight") return;
    if (items.length <= 1) return;

    e.preventDefault();
    const currentIndex = items.findIndex((it) => it.key === resolvedKey);
    const validIndex = currentIndex >= 0 ? currentIndex : 0;
    const nextIndex =
      e.key === "ArrowRight"
        ? (validIndex + 1) % items.length
        : (validIndex - 1 + items.length) % items.length;

    const nextItem = items[nextIndex];
    setActiveKey(nextItem.key);
    const nextBtn = document.getElementById(`${actualId}-tab-${nextItem.key}`);
    nextBtn?.focus();
  };

  return (
    <div
      id={actualId}
      data-testid="tabs-container"
      data-variant={variant || "reference"}
      className="my-3 font-sans"
    >
      {title && (
        <div className="mb-2 font-serif text-base font-semibold text-sea-ink">
          {title}
        </div>
      )}

      {/* TabList 标头 */}
      <div
        role="tablist"
        onKeyDown={handleKeyDown}
        className="flex overflow-x-auto border-b border-line font-mono text-xs"
      >
        {items.map((item) => {
          const isSelected = item.key === resolvedKey;
          return (
            <button
              key={item.key}
              id={`${actualId}-tab-${item.key}`}
              role="tab"
              type="button"
              aria-selected={isSelected}
              aria-controls={`${actualId}-panel`}
              tabIndex={isSelected ? 0 : -1}
              onClick={() => setActiveKey(item.key)}
              className={`cursor-pointer px-4 py-2 text-xs font-semibold select-none border-b-2 transition-colors uppercase tracking-wider shrink-0 whitespace-nowrap ${
                isSelected
                  ? "border-lagoon text-lagoon-deep font-bold"
                  : "border-transparent text-sea-ink-soft hover:text-sea-ink hover:bg-surface/50"
              }`}
            >
              {item.label}
            </button>
          );
        })}
      </div>

      {/* 面板内容 */}
      <div
        id={`${actualId}-panel`}
        role="tabpanel"
        aria-labelledby={`${actualId}-tab-${resolvedKey}`}
        className="py-6 px-1 [&>[data-testid$='-block']]:my-2 [&>[data-testid$='-block']:first-child]:mt-0 [&>[data-testid$='-block']:last-child]:mb-0"
      >
        {activeChild ? (
          renderContainerBlocks(
            [activeChild],
            location,
            contract,
            `tabs/${variant || "reference"}`,
          )
        ) : (
          <div
            data-testid="tabs-empty-slot"
            className="p-8 text-center text-xs font-serif italic text-sea-ink-soft/70"
          >
            该页签缺少内容块
          </div>
        )}
      </div>
    </div>
  );
};
