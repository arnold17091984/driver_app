import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/authStore';
import { useI18nStore } from '../stores/i18nStore';
import type { Locale } from '../i18n';

const LOCALES: { value: Locale; label: string }[] = [
  { value: 'en', label: 'English' },
  { value: 'ja', label: '日本語' },
  { value: 'ko', label: '한국어' },
  { value: 'zh', label: '中文' },
];

export function LoginPage() {
  const [employeeId, setEmployeeId] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { login, isLoading } = useAuthStore();
  const { t, locale, setLocale } = useI18nStore();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await login(employeeId, password);
      navigate('/');
    } catch {
      setError(t('login.error'));
    }
  };

  return (
    <div style={{
      display: 'flex',
      height: '100vh',
      background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0f172a 100%)',
      justifyContent: 'center',
      alignItems: 'center',
    }}>
      <div style={{ width: 400, padding: '0 20px' }}>
        {/* Brand */}
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div style={{
            width: 56, height: 56, borderRadius: 16,
            background: 'linear-gradient(135deg, #3b82f6, #2563eb)',
            display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 26, fontWeight: 700, color: '#fff',
            marginBottom: 16, boxShadow: '0 8px 24px rgba(37,99,235,0.3)',
          }}>V</div>
          <h1 style={{ margin: 0, fontSize: '1.5rem', fontWeight: 700, color: '#f1f5f9' }}>
            {t('common.appName')}
          </h1>
          <p style={{ marginTop: 6, color: '#64748b', fontSize: '0.875rem' }}>
            {t('login.subtitle')}
          </p>
        </div>

        {/* Form card */}
        <form
          onSubmit={handleSubmit}
          style={{
            padding: 32,
            background: '#fff',
            borderRadius: 16,
            boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25)',
          }}
        >
          {error && (
            <div style={{
              padding: '10px 14px',
              marginBottom: 20,
              background: '#fef2f2',
              color: '#dc2626',
              borderRadius: 8,
              fontSize: '0.85rem',
              border: '1px solid #fecaca',
            }}>
              {error}
            </div>
          )}

          <div style={{ marginBottom: 18 }}>
            <label style={{ display: 'block', marginBottom: 6, fontSize: '0.8rem', fontWeight: 600, color: '#475569' }}>
              {t('login.employeeIdLabel')}
            </label>
            <input
              type="text"
              value={employeeId}
              onChange={(e) => setEmployeeId(e.target.value)}
              placeholder={t('login.employeeIdPlaceholder')}
              required
              autoFocus
            />
          </div>

          <div style={{ marginBottom: 28 }}>
            <label style={{ display: 'block', marginBottom: 6, fontSize: '0.8rem', fontWeight: 600, color: '#475569' }}>
              {t('login.passwordLabel')}
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder={t('login.passwordPlaceholder')}
              required
            />
          </div>

          <button
            type="submit"
            disabled={isLoading}
            style={{
              width: '100%',
              padding: '11px',
              background: isLoading ? '#93c5fd' : '#2563eb',
              color: '#fff',
              border: 'none',
              borderRadius: 8,
              fontSize: '0.9rem',
              fontWeight: 600,
              fontFamily: 'inherit',
              cursor: isLoading ? 'wait' : 'pointer',
              transition: 'background 150ms ease',
              boxShadow: '0 1px 3px rgba(37,99,235,0.3)',
            }}
            onMouseEnter={(e) => { if (!isLoading) e.currentTarget.style.background = '#1d4ed8'; }}
            onMouseLeave={(e) => { if (!isLoading) e.currentTarget.style.background = '#2563eb'; }}
          >
            {isLoading ? t('login.signingIn') : t('login.signIn')}
          </button>

          {/* Language selector */}
          <div style={{ marginTop: 16, textAlign: 'center' }}>
            <select
              value={locale}
              onChange={(e) => setLocale(e.target.value as Locale)}
              style={{
                padding: '5px 10px',
                background: '#f8fafc',
                border: '1px solid #e2e8f0',
                borderRadius: 6,
                fontSize: '0.8rem',
                color: '#64748b',
                cursor: 'pointer',
                fontFamily: 'inherit',
              }}
            >
              {LOCALES.map((l) => (
                <option key={l.value} value={l.value}>{l.label}</option>
              ))}
            </select>
          </div>
        </form>
      </div>
    </div>
  );
}
