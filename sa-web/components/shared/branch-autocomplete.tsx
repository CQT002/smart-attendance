"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Input } from "@/components/ui/input";
import { Branch } from "@/types/branch";
import { Check, X } from "lucide-react";
import { cn } from "@/lib/utils";

interface BranchAutocompleteProps {
  branches: Branch[] | undefined;
  value?: number; // selected branch_id
  onChange: (branchId: number | undefined) => void;
  placeholder?: string;
  className?: string;
}

/**
 * Custom autocomplete cho chi nhánh — dropdown styled tốt hơn native datalist.
 * Filter theo cả tên + mã chi nhánh (case-insensitive).
 */
export function BranchAutocomplete({
  branches,
  value,
  onChange,
  placeholder = "Tìm kiếm chi nhánh...",
  className,
}: BranchAutocompleteProps) {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);

  // Sync query với value prop khi load lại từ ngoài
  useEffect(() => {
    if (value && branches) {
      const b = branches.find((x) => x.id === value);
      if (b) setQuery(`${b.name} (${b.code})`);
    } else if (!value) {
      setQuery("");
    }
  }, [value, branches]);

  // Click outside → close
  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", onClick);
    return () => document.removeEventListener("mousedown", onClick);
  }, []);

  const filtered = useMemo(() => {
    if (!branches) return [];
    const q = query.trim().toLowerCase();
    if (!q) return branches.slice(0, 50);
    return branches
      .filter(
        (b) =>
          b.name.toLowerCase().includes(q) || b.code.toLowerCase().includes(q)
      )
      .slice(0, 50);
  }, [branches, query]);

  const selectBranch = (b: Branch) => {
    onChange(b.id);
    setQuery(`${b.name} (${b.code})`);
    setOpen(false);
    setActiveIndex(-1);
  };

  const clear = () => {
    onChange(undefined);
    setQuery("");
    setActiveIndex(-1);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (!open) {
      if (e.key === "ArrowDown") {
        setOpen(true);
        e.preventDefault();
      }
      return;
    }
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActiveIndex((i) => Math.min(i + 1, filtered.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActiveIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && activeIndex >= 0) {
      e.preventDefault();
      selectBranch(filtered[activeIndex]);
    } else if (e.key === "Escape") {
      setOpen(false);
    }
  };

  return (
    <div ref={containerRef} className={cn("relative", className)}>
      <div className="relative">
        <Input
          placeholder={placeholder}
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setOpen(true);
            setActiveIndex(-1);
            // Nếu user xoá hết → bỏ filter
            if (e.target.value.trim() === "") {
              onChange(undefined);
            }
          }}
          onFocus={() => setOpen(true)}
          onKeyDown={handleKeyDown}
          className="pr-8"
        />
        {query && (
          <button
            type="button"
            onClick={clear}
            className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            aria-label="Xoá"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {open && filtered.length > 0 && (
        <div className="absolute z-50 mt-1 w-full max-h-72 overflow-auto rounded-md border bg-popover text-popover-foreground shadow-md">
          {filtered.map((b, idx) => {
            const isActive = idx === activeIndex;
            const isSelected = value === b.id;
            return (
              <button
                key={b.id}
                type="button"
                onClick={() => selectBranch(b)}
                onMouseEnter={() => setActiveIndex(idx)}
                className={cn(
                  "flex w-full items-center justify-between px-3 py-2 text-sm transition-colors text-left",
                  isActive ? "bg-accent text-accent-foreground" : "hover:bg-accent hover:text-accent-foreground"
                )}
              >
                <div className="min-w-0 flex-1">
                  <div className="truncate font-medium">{b.name}</div>
                  <div className="truncate text-xs text-muted-foreground">{b.code}</div>
                </div>
                {isSelected && <Check className="h-4 w-4 shrink-0 text-primary" />}
              </button>
            );
          })}
        </div>
      )}

      {open && filtered.length === 0 && (
        <div className="absolute z-50 mt-1 w-full rounded-md border bg-popover px-3 py-4 text-sm text-muted-foreground shadow-md">
          Không tìm thấy chi nhánh
        </div>
      )}
    </div>
  );
}
