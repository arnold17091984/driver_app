import { useIsMobile } from '../../hooks/useIsMobile';

interface ResponsiveTableProps {
  children: React.ReactNode;
}

export function ResponsiveTable({ children }: ResponsiveTableProps) {
  const isMobile = useIsMobile();

  if (!isMobile) {
    return <>{children}</>;
  }

  return (
    <div style={{
      overflowX: 'auto',
      WebkitOverflowScrolling: 'touch',
      margin: '0 -16px',
      padding: '0 16px',
    }}>
      {children}
    </div>
  );
}
