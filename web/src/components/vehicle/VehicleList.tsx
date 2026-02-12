import type { Vehicle, Dispatch } from '../../types/api';
import { VehicleStatusBadge } from './VehicleStatusBadge';
import { formatRelative } from '../../utils/formatters';
import { useI18nStore } from '../../stores/i18nStore';

interface Props {
  vehicles: Vehicle[];
  dispatches?: Dispatch[];
  selectedId: string | null;
  onSelect: (id: string) => void;
}

function formatEndTime(isoStr: string): string {
  const d = new Date(isoStr);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

export function VehicleList({ vehicles, dispatches = [], selectedId, onSelect }: Props) {
  const { t, locale } = useI18nStore();

  const getActiveDispatch = (vehicleId: string) =>
    dispatches.find(d => d.vehicle_id === vehicleId && ['assigned', 'accepted', 'en_route', 'arrived'].includes(d.status));

  function getTimeRemaining(isoStr: string): { text: string; overdue: boolean } {
    const diff = new Date(isoStr).getTime() - Date.now();
    if (diff <= 0) {
      const overMin = Math.ceil(-diff / 60000);
      return { text: t('format.overdue', { min: overMin }), overdue: true };
    }
    const min = Math.ceil(diff / 60000);
    if (min < 60) return { text: t('format.remainingMinutes', { min }), overdue: false };
    const h = Math.floor(min / 60);
    const m = min % 60;
    if (m > 0) return { text: t('format.remainingHours', { h, m }), overdue: false };
    return { text: t('format.remainingHoursOnly', { h }), overdue: false };
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      {vehicles.map((v) => {
        const isSelected = selectedId === v.id;
        const activeDispatch = getActiveDispatch(v.id);
        const endAt = activeDispatch?.estimated_end_at;
        const remaining = endAt ? getTimeRemaining(endAt) : null;

        return (
          <div
            key={v.id}
            onClick={() => onSelect(v.id)}
            style={{
              padding: '14px 16px',
              background: isSelected ? '#eff6ff' : '#fff',
              border: `1.5px solid ${isSelected ? '#3b82f6' : '#e2e8f0'}`,
              borderRadius: 10,
              cursor: 'pointer',
              transition: 'all 150ms ease',
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontWeight: 600, fontSize: '0.9rem', color: '#0f172a' }}>{v.name}</span>
              <VehicleStatusBadge status={v.status} />
            </div>
            <div style={{
              display: 'flex', alignItems: 'center', gap: 8,
              fontSize: '0.8rem', color: '#64748b', marginTop: 6,
            }}>
              <span>{v.license_plate}</span>
              <span style={{ color: '#cbd5e1' }}>|</span>
              <span>{v.driver_name}</span>
            </div>

            {/* Passenger info + estimated end (read-only) */}
            {activeDispatch?.passenger_name && (
              <div style={{
                marginTop: 8, padding: '6px 10px',
                background: '#f0f9ff', borderRadius: 6, border: '1px solid #bae6fd',
                fontSize: '0.78rem', color: '#0369a1',
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <span style={{ fontWeight: 600 }}>{activeDispatch.passenger_name}</span>
                  {activeDispatch.purpose && activeDispatch.purpose !== '乗車' && (
                    <span style={{ color: '#64748b', marginLeft: 'auto', fontSize: '0.72rem' }}>
                      {activeDispatch.purpose}
                    </span>
                  )}
                </div>
                {endAt && remaining && (
                  <div style={{
                    marginTop: 4, display: 'flex', justifyContent: 'space-between',
                    alignItems: 'center', fontSize: '0.72rem',
                  }}>
                    <span style={{ color: '#64748b' }}>
                      {t('format.estimatedEnd', { time: formatEndTime(endAt) })}
                    </span>
                    <span style={{
                      fontWeight: 600,
                      color: remaining.overdue ? '#dc2626' : '#0369a1',
                    }}>
                      {remaining.text}
                    </span>
                  </div>
                )}
                {!endAt && (
                  <div style={{ marginTop: 4, fontSize: '0.72rem', color: '#94a3b8' }}>
                    {t('format.estimatedEndUndecided')}
                  </div>
                )}
              </div>
            )}

            {v.location_at && !activeDispatch && (
              <div style={{ fontSize: '0.72rem', color: '#94a3b8', marginTop: 4 }}>
                {t('format.lastUpdate', { time: formatRelative(v.location_at, locale) })}
              </div>
            )}
          </div>
        );
      })}
      {vehicles.length === 0 && (
        <div style={{ padding: 32, textAlign: 'center', color: '#94a3b8', fontSize: '0.85rem' }}>
          {t('dashboard.noVehicles')}
        </div>
      )}
    </div>
  );
}
