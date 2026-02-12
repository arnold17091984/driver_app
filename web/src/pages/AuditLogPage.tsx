import { useState, useEffect } from 'react';
import { listAuditLogs } from '../api/admin';
import type { AuditLog } from '../types/api';
import { formatDateTime } from '../utils/formatters';
import { useI18nStore } from '../stores/i18nStore';
import { useIsMobile } from '../hooks/useIsMobile';
import { ResponsiveTable } from '../components/common/ResponsiveTable';
import type { TranslationKey } from '../i18n';

const actionKeys: Record<string, TranslationKey> = {
  'dispatch.create': 'auditLog.actionDispatchCreate',
  'dispatch.assign': 'auditLog.actionDispatchAssign',
  'dispatch.cancel': 'auditLog.actionDispatchCancel',
  'reservation.create': 'auditLog.actionReservationCreate',
  'reservation.cancel': 'auditLog.actionReservationCancel',
  'conflict.force_assign': 'auditLog.actionForceAssign',
  'user.role_change': 'auditLog.actionRoleChange',
  'vehicle.maintenance_toggle': 'auditLog.actionMaintenanceToggle',
};

const targetKeys: Record<string, TranslationKey> = {
  dispatch: 'auditLog.targetDispatch',
  reservation: 'auditLog.targetReservation',
  conflict: 'auditLog.targetConflict',
  vehicle: 'auditLog.targetVehicle',
  user: 'auditLog.targetUser',
};

export function AuditLogPage() {
  const { t, locale } = useI18nStore();
  const isMobile = useIsMobile();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [filters, setFilters] = useState({ action: '', target_type: '' });

  const fetchLogs = async () => {
    const params: Record<string, string> = {};
    if (filters.action) params.action = filters.action;
    if (filters.target_type) params.target_type = filters.target_type;
    const data = await listAuditLogs(params);
    setLogs(data || []);
  };

  useEffect(() => { fetchLogs(); }, []);

  const tableHeaders: { key: string; label: string }[] = [
    { key: 'timestamp', label: t('auditLog.tableHeaderTimestamp') },
    { key: 'actor', label: t('auditLog.tableHeaderActor') },
    { key: 'action', label: t('auditLog.tableHeaderAction') },
    { key: 'target', label: t('auditLog.tableHeaderTarget') },
    { key: 'reason', label: t('auditLog.tableHeaderReason') },
  ];

  return (
    <div style={{ padding: isMobile ? 16 : 28 }}>
      <div style={{ marginBottom: isMobile ? 16 : 24 }}>
        <h1 style={{ margin: 0, fontSize: '1.25rem', fontWeight: 700, color: '#0f172a' }}>{t('auditLog.title')}</h1>
        <p style={{ margin: '2px 0 0', fontSize: '0.8rem', color: '#94a3b8' }}>
          {t('auditLog.subtitle')}
        </p>
      </div>

      {/* Filters */}
      <div style={{
        display: 'flex', gap: 10, marginBottom: 20, alignItems: 'center', flexWrap: 'wrap',
        padding: '14px 18px', background: '#fff', borderRadius: 10,
        border: '1px solid #e2e8f0',
      }}>
        <span style={{ fontSize: '0.8rem', fontWeight: 600, color: '#64748b', marginRight: 4 }}>{t('auditLog.filters')}</span>
        <select value={filters.action} onChange={e => setFilters(f => ({ ...f, action: e.target.value }))}
          style={{ width: 'auto', padding: '7px 12px', borderRadius: 6, fontSize: '0.82rem' }}>
          <option value="">{t('auditLog.allActions')}</option>
          {Object.entries(actionKeys).map(([val, key]) => (
            <option key={val} value={val}>{t(key)}</option>
          ))}
        </select>
        <select value={filters.target_type} onChange={e => setFilters(f => ({ ...f, target_type: e.target.value }))}
          style={{ width: 'auto', padding: '7px 12px', borderRadius: 6, fontSize: '0.82rem' }}>
          <option value="">{t('auditLog.allTargets')}</option>
          {Object.entries(targetKeys).map(([val, key]) => (
            <option key={val} value={val}>{t(key)}</option>
          ))}
        </select>
        <button onClick={fetchLogs} style={{
          padding: '7px 18px', background: '#2563eb', color: '#fff',
          border: 'none', borderRadius: 6, cursor: 'pointer',
          fontWeight: 600, fontSize: '0.82rem', fontFamily: 'inherit',
        }}>{t('common.search')}</button>
      </div>

      {/* Logs table */}
      <div style={{
        background: '#fff', borderRadius: 12,
        border: '1px solid #e2e8f0', overflow: 'hidden',
        boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
      }}>
        <ResponsiveTable>
        <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: isMobile ? 700 : undefined }}>
          <thead>
            <tr>
              {tableHeaders.map((h) => (
                <th key={h.key} style={{ padding: '12px 16px', textAlign: 'left', fontSize: '0.75rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.04em', color: '#64748b', background: '#f8fafc', borderBottom: '1px solid #e2e8f0' }}>{h.label}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {logs.map((log) => (
              <tr key={log.id} style={{ borderBottom: '1px solid #f1f5f9' }}>
                <td style={{ padding: '12px 16px', fontSize: '0.82rem', whiteSpace: 'nowrap', color: '#64748b' }}>{formatDateTime(log.created_at, locale)}</td>
                <td style={{ padding: '12px 16px', fontSize: '0.85rem', fontWeight: 500 }}>{log.actor_name}</td>
                <td style={{ padding: '12px 16px' }}>
                  <span style={{
                    fontSize: '0.78rem', fontWeight: 600,
                    padding: '3px 10px', borderRadius: 6,
                    background: '#f1f5f9', color: '#475569',
                    fontFamily: 'monospace',
                  }}>{log.action}</span>
                </td>
                <td style={{ padding: '12px 16px', fontSize: '0.82rem', color: '#64748b' }}>
                  <span style={{ textTransform: 'capitalize' }}>{log.target_type}</span>
                  <span style={{ color: '#cbd5e1' }}> / </span>
                  <span style={{ fontFamily: 'monospace', fontSize: '0.78rem' }}>{log.target_id.slice(0, 8)}</span>
                </td>
                <td style={{ padding: '12px 16px', fontSize: '0.82rem', color: '#94a3b8', maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {log.reason || '-'}
                </td>
              </tr>
            ))}
            {logs.length === 0 && (
              <tr><td colSpan={5} style={{ padding: 48, textAlign: 'center', color: '#94a3b8', fontSize: '0.9rem' }}>{t('auditLog.noLogs')}</td></tr>
            )}
          </tbody>
        </table>
        </ResponsiveTable>
      </div>
    </div>
  );
}
