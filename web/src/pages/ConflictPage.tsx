import { useState, useEffect } from 'react';
import { listConflicts, getConflict, resolveConflictCancel, forceAssign } from '../api/reservations';
import type { ReservationConflict, ConflictDetail } from '../types/api';
import { formatDateTime } from '../utils/formatters';
import { usePermission } from '../hooks/usePermission';
import { useI18nStore } from '../stores/i18nStore';
import { useIsMobile } from '../hooks/useIsMobile';

export function ConflictPage() {
  const { t, locale } = useI18nStore();
  const isMobile = useIsMobile();
  const [conflicts, setConflicts] = useState<ReservationConflict[]>([]);
  const [detail, setDetail] = useState<ConflictDetail | null>(null);
  const isAdmin = usePermission('admin');

  const fetchConflicts = async () => {
    const data = await listConflicts();
    setConflicts(data || []);
  };

  useEffect(() => { fetchConflicts(); }, []);

  const handleViewDetail = async (id: string) => {
    const data = await getConflict(id);
    setDetail(data);
  };

  const handleCancelLosing = async (conflictId: string) => {
    const reason = prompt(t('conflict.cancelReason'));
    if (reason) {
      await resolveConflictCancel(conflictId, reason);
      setDetail(null);
      fetchConflicts();
    }
  };

  const handleForceAssign = async (conflictId: string) => {
    const reason = prompt(t('conflict.forceAssignReason'));
    if (reason) {
      await forceAssign(conflictId, reason);
      setDetail(null);
      fetchConflicts();
    }
  };

  return (
    <div style={{ padding: isMobile ? 16 : 28 }}>
      <div style={{ marginBottom: isMobile ? 16 : 24 }}>
        <h1 style={{ margin: 0, fontSize: '1.25rem', fontWeight: 700, color: '#0f172a' }}>
          {t('conflict.title')}
          {conflicts.length > 0 && (
            <span style={{
              marginLeft: 10, padding: '2px 10px', borderRadius: 9999,
              background: '#fef3c7', color: '#92400e', fontSize: '0.8rem', fontWeight: 600,
            }}>{conflicts.length}</span>
          )}
        </h1>
        <p style={{ margin: '2px 0 0', fontSize: '0.8rem', color: '#94a3b8' }}>
          {t('conflict.subtitle')}
        </p>
      </div>

      {conflicts.length === 0 ? (
        <div style={{
          padding: 64, textAlign: 'center', background: '#fff',
          borderRadius: 12, border: '1px solid #e2e8f0',
        }}>
          <div style={{ fontSize: '2rem', marginBottom: 12 }}>&#10003;</div>
          <div style={{ fontSize: '1rem', fontWeight: 600, color: '#0f172a', marginBottom: 4 }}>{t('conflict.allClear')}</div>
          <div style={{ color: '#94a3b8', fontSize: '0.85rem' }}>{t('conflict.noConflicts')}</div>
        </div>
      ) : (
        <div style={{ display: 'grid', gap: 10 }}>
          {conflicts.map((c) => (
            <div key={c.id}
              onClick={() => handleViewDetail(c.id)}
              style={{
                padding: '16px 20px', background: '#fff',
                borderRadius: 10, border: '1px solid #fbbf24',
                borderLeft: '4px solid #f59e0b',
                cursor: 'pointer',
                transition: 'box-shadow 150ms',
                boxShadow: '0 1px 2px rgba(0,0,0,0.04)',
              }}
              onMouseEnter={(e) => e.currentTarget.style.boxShadow = '0 4px 12px rgba(245,158,11,0.15)'}
              onMouseLeave={(e) => e.currentTarget.style.boxShadow = '0 1px 2px rgba(0,0,0,0.04)'}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <span style={{
                    width: 8, height: 8, borderRadius: '50%',
                    background: '#f59e0b', animation: 'pulse 2s infinite',
                  }} />
                  <span style={{ fontWeight: 600, color: '#0f172a' }}>{t('conflict.conflictId', { id: c.id.slice(0, 8) })}</span>
                </div>
                <span style={{ fontSize: '0.8rem', color: '#94a3b8' }}>{formatDateTime(c.created_at, locale)}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Detail modal */}
      {detail && (
        <div style={{
          position: 'fixed', inset: 0,
          background: 'rgba(15,23,42,0.4)', backdropFilter: 'blur(4px)',
          display: 'flex', justifyContent: 'center', alignItems: 'center', zIndex: 100,
        }}>
          <div style={{
            background: '#fff', borderRadius: 16, padding: isMobile ? 20 : 28, width: isMobile ? '95vw' : 600,
            maxHeight: '85vh', overflow: 'auto',
            boxShadow: '0 20px 50px rgba(0,0,0,0.15)',
          }}>
            <h3 style={{ margin: '0 0 6px', fontSize: '1.1rem', fontWeight: 700 }}>{t('conflict.detailTitle')}</h3>
            <p style={{ margin: '0 0 24px', fontSize: '0.82rem', color: '#64748b' }}>
              {t('conflict.detailSubtitle')}
            </p>

            <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: 16, marginBottom: isMobile ? 20 : 28 }}>
              {/* Winner */}
              <div style={{
                padding: 16, background: '#f0fdf4',
                borderRadius: 10, border: '1px solid #bbf7d0',
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 10 }}>
                  <span style={{
                    padding: '2px 8px', borderRadius: 9999,
                    background: '#16a34a', color: '#fff',
                    fontSize: '0.7rem', fontWeight: 700,
                  }}>{t('conflict.winner')}</span>
                </div>
                <div style={{ fontSize: '0.9rem', fontWeight: 600, marginBottom: 4 }}>{detail.winning_reservation?.purpose}</div>
                <div style={{ fontSize: '0.82rem', color: '#475569', marginBottom: 2 }}>
                  {t('conflict.priority')} <strong>{detail.winning_reservation?.priority_level}</strong>
                </div>
                <div style={{ fontSize: '0.78rem', color: '#64748b' }}>
                  {detail.winning_reservation?.start_time && formatDateTime(detail.winning_reservation.start_time, locale)}
                  {' - '}
                  {detail.winning_reservation?.end_time && formatDateTime(detail.winning_reservation.end_time, locale)}
                </div>
              </div>

              {/* Loser */}
              <div style={{
                padding: 16, background: '#fef2f2',
                borderRadius: 10, border: '1px solid #fecaca',
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 10 }}>
                  <span style={{
                    padding: '2px 8px', borderRadius: 9999,
                    background: '#dc2626', color: '#fff',
                    fontSize: '0.7rem', fontWeight: 700,
                  }}>{t('conflict.conflicting')}</span>
                </div>
                <div style={{ fontSize: '0.9rem', fontWeight: 600, marginBottom: 4 }}>{detail.losing_reservation?.purpose}</div>
                <div style={{ fontSize: '0.82rem', color: '#475569', marginBottom: 2 }}>
                  {t('conflict.priority')} <strong>{detail.losing_reservation?.priority_level}</strong>
                </div>
                <div style={{ fontSize: '0.78rem', color: '#64748b' }}>
                  {detail.losing_reservation?.start_time && formatDateTime(detail.losing_reservation.start_time, locale)}
                  {' - '}
                  {detail.losing_reservation?.end_time && formatDateTime(detail.losing_reservation.end_time, locale)}
                </div>
              </div>
            </div>

            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              <button onClick={() => handleCancelLosing(detail.conflict.id)} style={{
                padding: '9px 20px', background: '#dc2626', color: '#fff',
                border: 'none', borderRadius: 8, cursor: 'pointer',
                fontWeight: 600, fontSize: '0.85rem', fontFamily: 'inherit',
              }}>{t('conflict.cancelConflicting')}</button>
              {isAdmin && (
                <button onClick={() => handleForceAssign(detail.conflict.id)} style={{
                  padding: '9px 20px', background: '#7c3aed', color: '#fff',
                  border: 'none', borderRadius: 8, cursor: 'pointer',
                  fontWeight: 600, fontSize: '0.85rem', fontFamily: 'inherit',
                }}>{t('conflict.forceAssign')}</button>
              )}
              <button onClick={() => setDetail(null)} style={{
                padding: '9px 20px', background: '#f1f5f9', color: '#475569',
                border: 'none', borderRadius: 8, cursor: 'pointer',
                fontWeight: 500, fontSize: '0.85rem', fontFamily: 'inherit',
                marginLeft: 'auto',
              }}>{t('common.close')}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
