import { useState, useCallback } from 'react';
import { usePolling } from '../hooks/usePolling';
import { driverBoard, driverAlight, driverCurrentTrip } from '../api/dispatches';
import { driverPendingReservations, driverAcceptReservation, driverDeclineReservation } from '../api/bookings';
import { useI18nStore } from '../stores/i18nStore';
import { useNotificationSound } from '../hooks/useNotificationSound';
import { formatTime } from '../utils/formatters';
import type { Dispatch, Reservation } from '../types/api';
import type { TranslationKey } from '../i18n';

const DURATION_OPTIONS: { key: TranslationKey; value: number }[] = [
  { key: 'driver.duration30m', value: 30 },
  { key: 'driver.duration1h', value: 60 },
  { key: 'driver.duration2h', value: 120 },
  { key: 'driver.duration3h', value: 180 },
  { key: 'driver.durationHalfDay', value: 240 },
  { key: 'driver.durationUndecided', value: 0 },
];

function formatEstimatedEnd(isoStr: string): string {
  const d = new Date(isoStr);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export function DriverPage() {
  const { t } = useI18nStore();
  const [currentTrip, setCurrentTrip] = useState<Dispatch | null>(null);
  const [pendingRes, setPendingRes] = useState<Reservation[]>([]);
  const [passengerName, setPassengerName] = useState('');
  const [passengerCount, setPassengerCount] = useState(1);
  const [purpose, setPurpose] = useState('');
  const [estimatedMinutes, setEstimatedMinutes] = useState(60);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { checkForNew: checkNewReservations } = useNotificationSound<Reservation>('urgent');

  const fetchTrip = useCallback(async () => {
    const [trip, pending] = await Promise.all([
      driverCurrentTrip(),
      driverPendingReservations().catch(() => []),
    ]);
    setCurrentTrip(trip);
    const pendingList = pending || [];
    setPendingRes(pendingList);
    checkNewReservations(pendingList);
  }, [checkNewReservations]);

  usePolling(fetchTrip, 5000);

  const handleAccept = async (id: string) => {
    setLoading(true);
    try {
      await driverAcceptReservation(id);
      await fetchTrip();
    } catch {
      setError(t('driver.boardError'));
    } finally {
      setLoading(false);
    }
  };

  const handleDecline = async (id: string) => {
    const reason = prompt(t('booking.declineReason'));
    if (reason === null) return;
    setLoading(true);
    try {
      await driverDeclineReservation(id, reason);
      await fetchTrip();
    } catch {
      setError(t('driver.boardError'));
    } finally {
      setLoading(false);
    }
  };

  function getTimeRemaining(isoStr: string): { text: string; overdue: boolean } {
    const diff = new Date(isoStr).getTime() - Date.now();
    if (diff <= 0) {
      const overMin = Math.ceil(-diff / 60000);
      return { text: t('format.overdue', { min: overMin }), overdue: true };
    }
    const min = Math.ceil(diff / 60000);
    if (min < 60) return { text: t('driver.timeRemainingMinutes', { min }), overdue: false };
    const h = Math.floor(min / 60);
    const m = min % 60;
    if (m > 0) return { text: t('driver.timeRemainingHours', { h, m }), overdue: false };
    return { text: t('driver.timeRemainingHoursOnly', { h }), overdue: false };
  }

  const handleBoard = async () => {
    if (!passengerName.trim()) return;
    setLoading(true);
    setError('');
    try {
      await driverBoard(passengerName.trim(), purpose.trim() || undefined, passengerCount, estimatedMinutes);
      setPassengerName('');
      setPassengerCount(1);
      setPurpose('');
      setEstimatedMinutes(60);
      await fetchTrip();
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } };
      setError(err?.response?.data?.message || t('driver.boardError'));
    } finally {
      setLoading(false);
    }
  };

  const handleAlight = async () => {
    if (!currentTrip) return;
    setLoading(true);
    setError('');
    try {
      await driverAlight(currentTrip.id);
      setCurrentTrip(null);
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } };
      setError(err?.response?.data?.message || t('driver.alightError'));
    } finally {
      setLoading(false);
    }
  };

  const hasTrip = currentTrip && ['assigned', 'accepted', 'en_route', 'arrived'].includes(currentTrip.status);

  return (
    <div style={{
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      height: '100%', padding: 24,
    }}>
      <div style={{ width: '100%', maxWidth: 420 }}>
        <h1 style={{ fontSize: '1.3rem', fontWeight: 700, color: '#0f172a', margin: '0 0 24px' }}>
          {t('driver.title')}
        </h1>

        {error && (
          <div style={{
            padding: '10px 14px', marginBottom: 16, borderRadius: 8,
            background: '#fef2f2', border: '1px solid #fecaca', color: '#dc2626',
            fontSize: '0.85rem',
          }}>
            {error}
          </div>
        )}

        {hasTrip ? (
          <div style={{
            background: '#fff', borderRadius: 12, padding: 24,
            border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}>
            <div style={{
              padding: '12px 16px', background: '#eff6ff', borderRadius: 8,
              marginBottom: 20, border: '1px solid #bfdbfe',
            }}>
              <div style={{ fontSize: '0.75rem', color: '#3b82f6', fontWeight: 600, marginBottom: 4 }}>
                {t('driver.inTrip')}
              </div>
              <div style={{ fontSize: '1.1rem', fontWeight: 700, color: '#1e40af' }}>
                {currentTrip!.passenger_name || t('driver.nameNotSet')}
              </div>
              {currentTrip!.passenger_count > 1 && (
                <div style={{ fontSize: '0.8rem', color: '#3b82f6', marginTop: 2 }}>
                  {t('driver.passengerCount', { count: currentTrip!.passenger_count })}
                </div>
              )}
              {currentTrip!.purpose && currentTrip!.purpose !== '乗車' && (
                <div style={{ fontSize: '0.8rem', color: '#64748b', marginTop: 4 }}>
                  {currentTrip!.purpose}
                </div>
              )}
              {currentTrip!.estimated_end_at && (() => {
                const remaining = getTimeRemaining(currentTrip!.estimated_end_at!);
                return (
                  <div style={{
                    marginTop: 8, padding: '6px 10px', background: '#fff', borderRadius: 6,
                    fontSize: '0.8rem', color: '#0f172a',
                    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                  }}>
                    <span>{t('driver.estimatedEndLabel', { time: formatEstimatedEnd(currentTrip!.estimated_end_at!) })}</span>
                    <span style={{
                      color: remaining.overdue ? '#dc2626' : '#3b82f6',
                      fontWeight: 600,
                    }}>
                      {remaining.text}
                    </span>
                  </div>
                );
              })()}
            </div>

            <button
              onClick={handleAlight}
              disabled={loading}
              style={{
                width: '100%', padding: '16px 0', border: 'none', borderRadius: 10,
                background: '#dc2626', color: '#fff', fontWeight: 700,
                fontSize: '1.1rem', cursor: loading ? 'wait' : 'pointer',
                opacity: loading ? 0.6 : 1, fontFamily: 'inherit',
              }}
            >
              {loading ? t('common.processing') : t('driver.alight')}
            </button>
          </div>
        ) : (
          <div style={{
            background: '#fff', borderRadius: 12, padding: 24,
            border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}>
            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#374151', marginBottom: 6 }}>
                {t('driver.passengerNameLabel')} <span style={{ color: '#dc2626' }}>*</span>
              </label>
              <input
                type="text"
                value={passengerName}
                onChange={e => setPassengerName(e.target.value)}
                placeholder={t('driver.passengerNamePlaceholder')}
                style={{
                  width: '100%', padding: '12px 14px', border: '1px solid #d1d5db',
                  borderRadius: 8, fontSize: '1rem', outline: 'none',
                  boxSizing: 'border-box',
                }}
                onFocus={e => e.target.style.borderColor = '#3b82f6'}
                onBlur={e => e.target.style.borderColor = '#d1d5db'}
              />
            </div>

            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#374151', marginBottom: 6 }}>
                {t('driver.countLabel')}
              </label>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <button
                  onClick={() => setPassengerCount(Math.max(1, passengerCount - 1))}
                  style={{
                    width: 40, height: 40, border: '1px solid #d1d5db', borderRadius: 8,
                    background: '#fff', fontSize: '1.2rem', cursor: 'pointer', fontFamily: 'inherit',
                  }}
                >-</button>
                <span style={{ fontSize: '1.1rem', fontWeight: 600, minWidth: 24, textAlign: 'center' }}>
                  {passengerCount}
                </span>
                <button
                  onClick={() => setPassengerCount(passengerCount + 1)}
                  style={{
                    width: 40, height: 40, border: '1px solid #d1d5db', borderRadius: 8,
                    background: '#fff', fontSize: '1.2rem', cursor: 'pointer', fontFamily: 'inherit',
                  }}
                >+</button>
              </div>
            </div>

            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#374151', marginBottom: 6 }}>
                {t('driver.estimatedDurationLabel')}
              </label>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
                {DURATION_OPTIONS.map(opt => (
                  <button
                    key={opt.value}
                    onClick={() => setEstimatedMinutes(opt.value)}
                    style={{
                      padding: '8px 14px', border: '1px solid',
                      borderColor: estimatedMinutes === opt.value ? '#3b82f6' : '#d1d5db',
                      borderRadius: 8,
                      background: estimatedMinutes === opt.value ? '#eff6ff' : '#fff',
                      color: estimatedMinutes === opt.value ? '#2563eb' : '#374151',
                      fontWeight: estimatedMinutes === opt.value ? 600 : 400,
                      fontSize: '0.85rem', cursor: 'pointer', fontFamily: 'inherit',
                    }}
                  >
                    {t(opt.key)}
                  </button>
                ))}
              </div>
            </div>

            <div style={{ marginBottom: 24 }}>
              <label style={{ display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#374151', marginBottom: 6 }}>
                {t('driver.purposeLabel')}
              </label>
              <input
                type="text"
                value={purpose}
                onChange={e => setPurpose(e.target.value)}
                placeholder={t('driver.purposePlaceholder')}
                style={{
                  width: '100%', padding: '12px 14px', border: '1px solid #d1d5db',
                  borderRadius: 8, fontSize: '1rem', outline: 'none',
                  boxSizing: 'border-box',
                }}
                onFocus={e => e.target.style.borderColor = '#3b82f6'}
                onBlur={e => e.target.style.borderColor = '#d1d5db'}
              />
            </div>

            <button
              onClick={handleBoard}
              disabled={loading || !passengerName.trim()}
              style={{
                width: '100%', padding: '16px 0', border: 'none', borderRadius: 10,
                background: !passengerName.trim() ? '#94a3b8' : '#2563eb',
                color: '#fff', fontWeight: 700, fontSize: '1.1rem',
                cursor: loading || !passengerName.trim() ? 'not-allowed' : 'pointer',
                opacity: loading ? 0.6 : 1, fontFamily: 'inherit',
              }}
            >
              {loading ? t('common.processing') : t('driver.board')}
            </button>
          </div>
        )}

        {/* Pending reservations */}
        {pendingRes.length > 0 && (
          <div style={{ marginTop: 20 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, margin: '0 0 12px' }}>
              <div style={{
                width: 22, height: 22, borderRadius: '50%', background: '#fef3c7',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <span style={{ fontSize: '0.7rem' }}>
                  {pendingRes.length}
                </span>
              </div>
              <h2 style={{ fontSize: '0.95rem', fontWeight: 700, color: '#0f172a', margin: 0 }}>
                {t('booking.driverPending')}
              </h2>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {pendingRes.map((res) => (
                <div key={res.id} style={{
                  background: '#fffbeb', borderRadius: 12, padding: 16,
                  border: '1px solid #fde68a', boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
                }}>
                  <div style={{ fontWeight: 600, fontSize: '0.9rem', color: '#0f172a', marginBottom: 4 }}>
                    {res.purpose}
                  </div>
                  <div style={{ fontSize: '0.8rem', color: '#64748b', marginBottom: 2 }}>
                    {res.requester_name}
                    {res.passenger_name ? ` \u00b7 ${res.passenger_name}` : ''}
                  </div>
                  <div style={{ fontSize: '0.78rem', color: '#94a3b8', marginBottom: 10 }}>
                    {formatTime(res.start_time)} - {formatTime(res.end_time)}
                    {res.pickup_address ? ` \u00b7 ${res.pickup_address}` : ''}
                  </div>
                  <div style={{ display: 'flex', gap: 8 }}>
                    <button onClick={() => handleAccept(res.id)} disabled={loading} style={{
                      flex: 1, padding: '11px', background: '#16a34a', color: '#fff',
                      border: 'none', borderRadius: 10, fontWeight: 700, fontSize: '0.85rem',
                      cursor: loading ? 'wait' : 'pointer', fontFamily: 'inherit',
                      opacity: loading ? 0.6 : 1,
                    }}>{t('booking.driverAccept')}</button>
                    <button onClick={() => handleDecline(res.id)} disabled={loading} style={{
                      flex: 1, padding: '11px', background: '#fff', color: '#dc2626',
                      border: '1px solid #fecaca', borderRadius: 10, fontWeight: 600, fontSize: '0.85rem',
                      cursor: loading ? 'wait' : 'pointer', fontFamily: 'inherit',
                      opacity: loading ? 0.6 : 1,
                    }}>{t('booking.driverDecline')}</button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
