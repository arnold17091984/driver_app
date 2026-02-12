import { NavLink } from 'react-router-dom';
import { useI18nStore } from '../../stores/i18nStore';
import { useAuthStore } from '../../stores/authStore';

const TAB_ICONS: Record<string, string> = {
  '/': 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0a1 1 0 01-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 01-1 1',
  '/settings': 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
  '/driver': 'M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4',
};

export function MobileNav() {
  const { t } = useI18nStore();
  const user = useAuthStore(s => s.user);

  const tabs: { to: string; label: string }[] = [
    { to: '/', label: t('nav.dashboard') },
  ];

  if (user?.role === 'driver') {
    tabs.push({ to: '/driver', label: t('nav.driverLog') });
  }
  if (user?.role === 'admin' || user?.role === 'dispatcher') {
    tabs.push({ to: '/settings', label: t('nav.settings') });
  }

  return (
    <nav style={{
      position: 'fixed',
      bottom: 0,
      left: 0,
      right: 0,
      height: 56,
      paddingBottom: 'env(safe-area-inset-bottom, 0px)',
      background: '#fff',
      borderTop: '1px solid #e2e8f0',
      display: 'flex',
      justifyContent: 'space-around',
      alignItems: 'center',
      zIndex: 50,
    }}>
      {tabs.map(tab => (
        <NavLink
          key={tab.to}
          to={tab.to}
          end={tab.to === '/'}
          style={({ isActive }) => ({
            display: 'flex',
            flexDirection: 'column' as const,
            alignItems: 'center',
            gap: 2,
            padding: '6px 12px',
            textDecoration: 'none',
            color: isActive ? '#2563eb' : '#94a3b8',
            fontSize: '0.65rem',
            fontWeight: isActive ? 600 : 400,
          })}
        >
          <svg
            width="22" height="22" fill="none"
            stroke="currentColor" strokeWidth="1.8"
            strokeLinecap="round" strokeLinejoin="round"
            viewBox="0 0 24 24"
          >
            <path d={TAB_ICONS[tab.to] || TAB_ICONS['/']} />
          </svg>
          <span>{tab.label}</span>
        </NavLink>
      ))}
    </nav>
  );
}
