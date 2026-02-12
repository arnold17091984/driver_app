import { useState, useRef, useEffect, useCallback } from 'react';

interface BottomSheetProps {
  children: React.ReactNode;
  peekHeight?: number;
  maxHeight?: string;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  header?: React.ReactNode;
  style?: React.CSSProperties;
  fitContent?: boolean;
}

export function BottomSheet({
  children,
  peekHeight = 120,
  maxHeight = '70vh',
  open = false,
  onOpenChange,
  header,
  style,
  fitContent = false,
}: BottomSheetProps) {
  const [expanded, setExpanded] = useState(open);
  const [dragY, setDragY] = useState(0);
  const [isDragging, setIsDragging] = useState(false);
  const startY = useRef(0);

  useEffect(() => { setExpanded(open); }, [open]);

  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    startY.current = e.touches[0].clientY;
    setIsDragging(true);
  }, []);

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (!isDragging) return;
    const delta = startY.current - e.touches[0].clientY;
    setDragY(delta);
  }, [isDragging]);

  const handleTouchEnd = useCallback(() => {
    setIsDragging(false);
    const threshold = 50;
    if (dragY > threshold && !expanded) {
      setExpanded(true);
      onOpenChange?.(true);
    } else if (dragY < -threshold && expanded) {
      setExpanded(false);
      onOpenChange?.(false);
    }
    setDragY(0);
  }, [dragY, expanded, onOpenChange]);

  const toggle = () => {
    const next = !expanded;
    setExpanded(next);
    onOpenChange?.(next);
  };

  const useFit = fitContent && expanded;
  const height = useFit ? 'auto' : expanded ? maxHeight : `${peekHeight}px`;

  return (
    <div
      style={{
        position: 'absolute',
        bottom: 0,
        left: 0,
        right: 0,
        height,
        maxHeight: useFit ? maxHeight : undefined,
        background: '#fff',
        borderTopLeftRadius: 16,
        borderTopRightRadius: 16,
        boxShadow: '0 -4px 20px rgba(0,0,0,0.12)',
        transition: isDragging || useFit ? 'none' : 'height 300ms cubic-bezier(0.4, 0, 0.2, 1)',
        display: 'flex',
        flexDirection: 'column',
        zIndex: 20,
        overflow: 'hidden',
        ...style,
      }}
    >
      {/* Drag handle */}
      <div
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
        onClick={toggle}
        style={{
          padding: '10px 0 6px',
          cursor: 'pointer',
          display: 'flex',
          justifyContent: 'center',
          flexShrink: 0,
        }}
      >
        <div style={{
          width: 36,
          height: 4,
          borderRadius: 2,
          background: '#cbd5e1',
        }} />
      </div>

      {/* Sticky header */}
      {header && (
        <div style={{ padding: '0 16px 10px', flexShrink: 0 }}>
          {header}
        </div>
      )}

      {/* Scrollable content */}
      <div style={{
        flex: 1,
        overflow: 'auto',
        padding: '0 16px 16px',
        WebkitOverflowScrolling: 'touch',
      }}>
        {children}
      </div>
    </div>
  );
}
