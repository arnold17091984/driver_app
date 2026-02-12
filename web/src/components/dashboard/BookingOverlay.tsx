import { useState, useEffect } from 'react';
import { useBookingStore, type BookingLocation } from '../../stores/bookingStore';
import { useI18nStore } from '../../stores/i18nStore';
import { calculateETAs } from '../../api/dispatches';
import { listDispatches } from '../../api/dispatches';
import { listReservations } from '../../api/reservations';
import type { Dispatch, Reservation } from '../../types/api';

/**
 * Booking flow rendered inside BottomSheet / side panel.
 * Steps: destination → pickup-map → vehicle-select → status
 */
export function BookingOverlay() {
  const step = useBookingStore(s => s.step);

  switch (step) {
    case 'destination':
      return <DestinationPanel />;
    case 'pickup-map':
      return <PickupPanel />;
    case 'vehicle-select':
      return <VehicleSelectPanel />;
    case 'status':
      return <StatusPanel />;
    default:
      return null;
  }
}

/* ─── Destination Panel ─── */
type ActiveField = 'origin' | 'destination';

function DestinationPanel() {
  const { t } = useI18nStore();
  const {
    origin, setOrigin, setDestination, setStep, reset,
  } = useBookingStore();

  const [activeField, setActiveField] = useState<ActiveField>('destination');
  const [originQuery, setOriginQuery] = useState('');
  const [destQuery, setDestQuery] = useState('');
  const [useCurrentLocation, setUseCurrentLocation] = useState(true);
  const [recentLocations, setRecentLocations] = useState<BookingLocation[]>([]);

  // Geolocation (only when using current location)
  useEffect(() => {
    if (!origin && useCurrentLocation) {
      navigator.geolocation?.getCurrentPosition(
        (pos) => {
          setOrigin({
            name: t('bookingFlow.currentLocation'),
            address: t('bookingFlow.currentLocation'),
            lat: pos.coords.latitude,
            lng: pos.coords.longitude,
          });
        },
        () => {},
        { enableHighAccuracy: false, timeout: 5000 },
      );
    }
  }, [origin, useCurrentLocation, setOrigin, t]);

  // Recent locations
  useEffect(() => {
    const locs: BookingLocation[] = [];
    const seen = new Set<string>();
    Promise.all([
      listDispatches().catch(() => [] as Dispatch[]),
      listReservations().catch(() => [] as Reservation[]),
    ]).then(([dispatches, reservations]) => {
      for (const d of (dispatches || []).slice(0, 10)) {
        if (d.pickup_address && !seen.has(d.pickup_address)) {
          seen.add(d.pickup_address);
          locs.push({ name: d.purpose || d.pickup_address, address: d.pickup_address, lat: d.pickup_lat || 0, lng: d.pickup_lng || 0 });
        }
        if (d.dropoff_address && !seen.has(d.dropoff_address)) {
          seen.add(d.dropoff_address);
          locs.push({ name: d.dropoff_address, address: d.dropoff_address, lat: d.dropoff_lat || 0, lng: d.dropoff_lng || 0 });
        }
      }
      for (const r of (reservations || []).slice(0, 10)) {
        if (r.pickup_address && !seen.has(r.pickup_address)) {
          seen.add(r.pickup_address);
          locs.push({ name: r.purpose || r.pickup_address, address: r.pickup_address, lat: r.pickup_lat || 0, lng: r.pickup_lng || 0 });
        }
      }
      setRecentLocations(locs.slice(0, 6));
    });
  }, []);

  const query = activeField === 'origin' ? originQuery : destQuery;
  const filtered = query.trim()
    ? recentLocations.filter(l =>
        l.name.toLowerCase().includes(query.toLowerCase()) ||
        l.address.toLowerCase().includes(query.toLowerCase()))
    : recentLocations;

  const handleSelect = (loc: BookingLocation) => {
    if (activeField === 'origin') {
      setOrigin(loc);
      setOriginQuery(loc.name);
      setUseCurrentLocation(false);
      setActiveField('destination');
      return;
    }
    // Destination selected — check if origin is ready
    setDestination(loc);
    setDestQuery(loc.name);
    if (origin || useCurrentLocation) {
      setStep('pickup-map');
    } else {
      setActiveField('origin');
    }
  };

  const handleFreeText = () => {
    if (!query.trim()) return;
    const loc: BookingLocation = { name: query.trim(), address: query.trim(), lat: 0, lng: 0 };
    if (activeField === 'origin') {
      setOrigin(loc);
      setUseCurrentLocation(false);
      setActiveField('destination');
      return;
    }
    setDestination(loc);
    if (origin || useCurrentLocation) {
      setStep('pickup-map');
    } else {
      setActiveField('origin');
    }
  };

  const handleOriginFocus = () => {
    setActiveField('origin');
    if (useCurrentLocation) setOriginQuery('');
  };

  const handleUseCurrentLocation = () => {
    setUseCurrentLocation(true);
    setOriginQuery('');
    setOrigin(null);
    setActiveField('destination');
  };

  const isOriginActive = activeField === 'origin';
  const isDestActive = activeField === 'destination';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
      {/* Back + Title */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <button onClick={() => reset()} style={{
          width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: '#f1f5f9', border: 'none', borderRadius: 8, cursor: 'pointer',
        }}>
          <svg width="16" height="16" fill="none" stroke="#475569" strokeWidth="2" viewBox="0 0 24 24"><path d="M19 12H5m7-7l-7 7 7 7" /></svg>
        </button>
        <span style={{ fontWeight: 700, fontSize: '0.95rem', color: '#0f172a' }}>
          {t('bookingFlow.destination')}
        </span>
      </div>

      {/* Dual input: Origin + Destination */}
      <div style={{
        display: 'flex', gap: 10,
        background: '#f8fafc', borderRadius: 12, padding: '10px 10px 10px 14px',
      }}>
        {/* Dots + connector line */}
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', paddingTop: 12, gap: 0 }}>
          <div style={{ width: 10, height: 10, borderRadius: '50%', background: '#16a34a', flexShrink: 0 }} />
          <div style={{ width: 2, flex: 1, background: '#cbd5e1', margin: '4px 0', minHeight: 12 }} />
          <div style={{ width: 10, height: 10, borderRadius: '50%', background: '#ef4444', flexShrink: 0 }} />
        </div>

        {/* Input fields */}
        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
          <input
            value={useCurrentLocation && origin && !isOriginActive
              ? origin.name
              : originQuery}
            onChange={e => { setOriginQuery(e.target.value); setUseCurrentLocation(false); }}
            onFocus={handleOriginFocus}
            placeholder={t('bookingFlow.originPlaceholder')}
            style={{
              padding: '10px 12px',
              border: `2px solid ${isOriginActive ? '#16a34a' : 'transparent'}`,
              borderRadius: 8, fontSize: '0.88rem', fontFamily: 'inherit', outline: 'none',
              background: '#fff',
              color: useCurrentLocation && !isOriginActive ? '#16a34a' : '#0f172a',
              fontWeight: useCurrentLocation && !isOriginActive ? 600 : 400,
            }}
          />
          <input
            autoFocus
            value={destQuery}
            onChange={e => setDestQuery(e.target.value)}
            onFocus={() => setActiveField('destination')}
            onKeyDown={e => e.key === 'Enter' && handleFreeText()}
            placeholder={t('bookingFlow.destinationPlaceholder')}
            style={{
              padding: '10px 12px',
              border: `2px solid ${isDestActive ? '#16a34a' : 'transparent'}`,
              borderRadius: 8, fontSize: '0.88rem', fontFamily: 'inherit', outline: 'none',
              background: '#fff',
            }}
          />
        </div>
      </div>

      {/* "Use current location" quick button (when editing origin) */}
      {isOriginActive && !useCurrentLocation && (
        <button
          onClick={handleUseCurrentLocation}
          style={{
            display: 'flex', alignItems: 'center', gap: 8,
            padding: '10px 8px', borderRadius: 8, cursor: 'pointer',
            background: 'none', border: 'none', fontFamily: 'inherit',
            width: '100%', textAlign: 'left',
          }}
          onMouseEnter={e => e.currentTarget.style.background = '#f0fdf4'}
          onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
        >
          <div style={{
            width: 28, height: 28, borderRadius: 8, background: '#f0fdf4',
            display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
          }}>
            <svg width="14" height="14" fill="none" stroke="#16a34a" strokeWidth="2" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="3" /><path d="M12 2v4m0 12v4m10-10h-4M6 12H2" />
            </svg>
          </div>
          <span style={{ fontWeight: 600, fontSize: '0.85rem', color: '#16a34a' }}>
            {t('bookingFlow.currentLocation')}
          </span>
        </button>
      )}

      {/* Recent locations */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
        {filtered.map((loc, idx) => (
          <div
            key={idx}
            onClick={() => handleSelect(loc)}
            style={{
              display: 'flex', alignItems: 'center', gap: 10,
              padding: '10px 8px', borderRadius: 8, cursor: 'pointer',
              transition: 'background 100ms',
            }}
            onMouseEnter={e => e.currentTarget.style.background = '#f8fafc'}
            onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
          >
            <svg width="14" height="14" fill="none" stroke="#94a3b8" strokeWidth="2" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" /><path d="M12 6v6l4 2" />
            </svg>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontWeight: 600, fontSize: '0.85rem', color: '#0f172a', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {loc.name}
              </div>
              {loc.address !== loc.name && (
                <div style={{ fontSize: '0.72rem', color: '#94a3b8', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {loc.address}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ─── Pickup Confirmation Panel ─── */
function PickupPanel() {
  const { t } = useI18nStore();
  const {
    origin, destination, pickupDetails, setPickupDetails,
    setAvailableVehicles, setStep,
  } = useBookingStore();
  const [isLoading, setIsLoading] = useState(false);

  const handleConfirm = async () => {
    if (!origin?.lat || !origin?.lng) return;
    setIsLoading(true);
    try {
      const etas = await calculateETAs(origin.lat, origin.lng);
      setAvailableVehicles(etas.filter(v => v.is_available));
    } catch {
      setAvailableVehicles([]);
    }
    setIsLoading(false);
    setStep('vehicle-select');
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <button onClick={() => setStep('destination')} style={{
          width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: '#f1f5f9', border: 'none', borderRadius: 8, cursor: 'pointer',
        }}>
          <svg width="16" height="16" fill="none" stroke="#475569" strokeWidth="2" viewBox="0 0 24 24"><path d="M19 12H5m7-7l-7 7 7 7" /></svg>
        </button>
        <span style={{ fontWeight: 700, fontSize: '0.95rem', color: '#0f172a' }}>
          {t('bookingFlow.pickupTitle')}
        </span>
      </div>

      {!origin?.lat && (
        <div style={{ fontSize: '0.85rem', color: '#94a3b8', textAlign: 'center', padding: 12 }}>
          {t('booking.tapToSetPickup')}
        </div>
      )}

      {origin?.lat && (
        <>
          <div>
            <div style={{ fontSize: '0.72rem', color: '#94a3b8' }}>{t('bookingFlow.origin')}</div>
            <div style={{ fontSize: '0.88rem', fontWeight: 600, color: '#0f172a' }}>{origin.address}</div>
          </div>
          {destination && (
            <div>
              <div style={{ fontSize: '0.72rem', color: '#94a3b8' }}>{t('bookingFlow.destination')}</div>
              <div style={{ fontSize: '0.88rem', fontWeight: 600, color: '#0f172a' }}>{destination.name}</div>
            </div>
          )}
          <input
            value={pickupDetails}
            onChange={e => setPickupDetails(e.target.value)}
            placeholder={t('bookingFlow.pickupDetailsPlaceholder')}
            style={{
              padding: '10px 14px', border: '1px solid #e2e8f0', borderRadius: 10,
              fontSize: '0.88rem', outline: 'none', fontFamily: 'inherit',
            }}
            onFocus={e => e.target.style.borderColor = '#16a34a'}
            onBlur={e => e.target.style.borderColor = '#e2e8f0'}
          />
        </>
      )}

      <button
        onClick={handleConfirm}
        disabled={!origin?.lat || isLoading}
        style={{
          width: '100%', padding: '14px',
          background: origin?.lat ? '#16a34a' : '#94a3b8',
          color: '#fff', border: 'none', borderRadius: 12,
          cursor: origin?.lat && !isLoading ? 'pointer' : 'not-allowed',
          fontWeight: 700, fontSize: '0.95rem', fontFamily: 'inherit',
          boxShadow: origin?.lat ? '0 4px 16px rgba(22,163,74,0.3)' : 'none',
          opacity: isLoading ? 0.6 : 1,
        }}
      >
        {isLoading ? t('common.processing') : t('bookingFlow.confirmPickup')}
      </button>
    </div>
  );
}

/* ─── Vehicle Select Panel ─── */
function VehicleSelectPanel() {
  const { t } = useI18nStore();
  const {
    availableVehicles, selectedVehicleId, setSelectedVehicleId,
    isNow, setIsNow, scheduledStart, scheduledEnd, setScheduledTime,
    passengerName, setPassengerName, purpose, setPurpose,
    notes, setNotes, destinations, setDestinations,
    setStep, submit,
  } = useBookingStore();

  const [detailsOpen, setDetailsOpen] = useState(false);

  useEffect(() => {
    if (availableVehicles.length > 0 && !selectedVehicleId) {
      setSelectedVehicleId(availableVehicles[0].vehicle_id);
    }
  }, [availableVehicles, selectedVehicleId, setSelectedVehicleId]);

  const selectedVehicle = availableVehicles.find(v => v.vehicle_id === selectedVehicleId);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <button onClick={() => setStep('pickup-map')} style={{
          width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: '#f1f5f9', border: 'none', borderRadius: 8, cursor: 'pointer',
        }}>
          <svg width="16" height="16" fill="none" stroke="#475569" strokeWidth="2" viewBox="0 0 24 24"><path d="M19 12H5m7-7l-7 7 7 7" /></svg>
        </button>
        <span style={{ fontWeight: 700, fontSize: '0.95rem', color: '#0f172a' }}>
          {t('bookingFlow.vehicleSelectTitle')}
        </span>
      </div>

      {/* Now / Schedule toggle */}
      <div style={{ display: 'flex', gap: 6 }}>
        {[true, false].map(nowVal => (
          <button key={String(nowVal)} onClick={() => setIsNow(nowVal)} style={{
            flex: 1, padding: '8px', border: '2px solid',
            borderColor: isNow === nowVal ? '#16a34a' : '#e2e8f0',
            background: isNow === nowVal ? '#f0fdf4' : '#fff',
            color: isNow === nowVal ? '#16a34a' : '#64748b',
            borderRadius: 8, fontWeight: 600, fontSize: '0.82rem',
            cursor: 'pointer', fontFamily: 'inherit',
          }}>
            {nowVal ? t('bookingFlow.now') : t('bookingFlow.scheduled')}
          </button>
        ))}
      </div>

      {!isNow && (
        <div style={{ display: 'flex', gap: 6 }}>
          <div style={{ flex: 1 }}>
            <label style={{ fontSize: '0.68rem', color: '#94a3b8', fontWeight: 600 }}>{t('bookingFlow.startTime')}</label>
            <input type="datetime-local" value={scheduledStart || ''} onChange={e => setScheduledTime(e.target.value, scheduledEnd)}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid #e2e8f0', borderRadius: 6, fontSize: '0.8rem', fontFamily: 'inherit' }} />
          </div>
          <div style={{ flex: 1 }}>
            <label style={{ fontSize: '0.68rem', color: '#94a3b8', fontWeight: 600 }}>{t('bookingFlow.endTime')}</label>
            <input type="datetime-local" value={scheduledEnd || ''} onChange={e => setScheduledTime(scheduledStart, e.target.value)}
              style={{ width: '100%', padding: '6px 8px', border: '1px solid #e2e8f0', borderRadius: 6, fontSize: '0.8rem', fontFamily: 'inherit' }} />
          </div>
        </div>
      )}

      {/* Vehicle list */}
      {availableVehicles.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 16, color: '#94a3b8', fontSize: '0.85rem' }}>
          {t('bookingFlow.noVehiclesNearby')}
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {availableVehicles.map(v => {
            const etaMin = Math.max(1, Math.round(v.duration_sec / 60));
            const selected = v.vehicle_id === selectedVehicleId;
            return (
              <button key={v.vehicle_id} onClick={() => setSelectedVehicleId(v.vehicle_id)} style={{
                display: 'flex', alignItems: 'center', gap: 10,
                padding: '10px 12px', background: selected ? '#f0fdf4' : '#fff',
                border: `2px solid ${selected ? '#16a34a' : '#e2e8f0'}`,
                borderRadius: 12, cursor: 'pointer', textAlign: 'left',
                width: '100%', fontFamily: 'inherit',
              }}>
                <div style={{
                  width: 36, height: 36, borderRadius: 10,
                  background: selected ? '#dcfce7' : '#f1f5f9',
                  display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0,
                }}>
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={selected ? '#16a34a' : '#64748b'} strokeWidth="1.5">
                    <path d="M5 17h14M5 17a2 2 0 01-2-2V7a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2M5 17l-1 3m15-3l1 3" />
                  </svg>
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 700, fontSize: '0.85rem', color: '#0f172a' }}>{v.vehicle_name}</div>
                  <div style={{ fontSize: '0.72rem', color: '#64748b' }}>{v.driver_name} · {v.plate}</div>
                </div>
                <div style={{ textAlign: 'right', flexShrink: 0 }}>
                  <div style={{ fontWeight: 700, fontSize: '0.85rem', color: selected ? '#16a34a' : '#0f172a' }}>
                    {t('bookingFlow.etaMinutes').replace('{min}', String(etaMin))}
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      )}

      {/* Details (collapsible) */}
      <button onClick={() => setDetailsOpen(!detailsOpen)} style={{
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        padding: '8px 0', background: 'none', border: 'none', cursor: 'pointer', fontFamily: 'inherit', width: '100%',
      }}>
        <span style={{ fontWeight: 600, fontSize: '0.82rem', color: '#475569' }}>{t('bookingFlow.detailsSection')}</span>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" strokeWidth="2"
          style={{ transform: detailsOpen ? 'rotate(180deg)' : 'none', transition: 'transform 150ms' }}>
          <path d="M6 9l6 6 6-6" />
        </svg>
      </button>
      {detailsOpen && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <PanelInput label={t('bookingFlow.passengerLabel')} value={passengerName} onChange={setPassengerName} placeholder={t('bookingFlow.passengerPlaceholder')} />
          <PanelInput label={t('bookingFlow.purposeLabel')} value={purpose} onChange={setPurpose} placeholder={t('bookingFlow.purposePlaceholder')} />
          <PanelInput label={t('bookingFlow.notesLabel')} value={notes} onChange={setNotes} placeholder={t('bookingFlow.notesPlaceholder')} />
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <label style={{ fontSize: '0.72rem', color: '#94a3b8', fontWeight: 600 }}>{t('bookingFlow.destinationsLabel')}</label>
              <button onClick={() => setDestinations([...destinations, ''])} style={{
                background: 'none', border: 'none', color: '#16a34a', fontWeight: 600, fontSize: '0.72rem', cursor: 'pointer', fontFamily: 'inherit',
              }}>+ {t('bookingFlow.addDestination')}</button>
            </div>
            {destinations.map((d, i) => (
              <input key={i} value={d} onChange={e => { const n = [...destinations]; n[i] = e.target.value; setDestinations(n); }}
                placeholder={t('booking.destinationPlaceholder')}
                style={{ width: '100%', padding: '7px 10px', border: '1px solid #e2e8f0', borderRadius: 6, fontSize: '0.82rem', fontFamily: 'inherit', marginTop: 3, outline: 'none', boxSizing: 'border-box' }} />
            ))}
          </div>
        </div>
      )}

      {/* CTA */}
      <button onClick={() => submit()} disabled={!selectedVehicleId} style={{
        width: '100%', padding: '14px',
        background: selectedVehicleId ? '#16a34a' : '#94a3b8',
        color: '#fff', border: 'none', borderRadius: 12,
        cursor: selectedVehicleId ? 'pointer' : 'not-allowed',
        fontWeight: 700, fontSize: '0.95rem', fontFamily: 'inherit',
        boxShadow: selectedVehicleId ? '0 4px 16px rgba(22,163,74,0.3)' : 'none',
      }}>
        {selectedVehicle
          ? t('bookingFlow.bookVehicle').replace('{name}', selectedVehicle.vehicle_name)
          : t('bookingFlow.vehicleSelectTitle')}
      </button>
    </div>
  );
}

/* ─── Status Panel ─── */
function StatusPanel() {
  const { t } = useI18nStore();
  const {
    isSubmitting, submitError, result,
    origin, destination, purpose, purposeCategory,
    submit, reset,
  } = useBookingStore();

  if (isSubmitting) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '24px 0' }}>
        <div style={{
          width: 48, height: 48, borderRadius: '50%',
          border: '4px solid #e2e8f0', borderTopColor: '#16a34a',
          animation: 'spin 1s linear infinite', marginBottom: 16,
        }} />
        <div style={{ fontWeight: 700, fontSize: '1rem', color: '#0f172a', marginBottom: 4 }}>
          {t('bookingFlow.searching')}
        </div>
        <div style={{ fontSize: '0.82rem', color: '#94a3b8' }}>{t('bookingFlow.searchingSub')}</div>
        <button onClick={() => reset()} style={{
          marginTop: 20, padding: '10px 24px', background: '#fff', color: '#ef4444',
          border: '1px solid #fecaca', borderRadius: 10, fontWeight: 600, cursor: 'pointer', fontFamily: 'inherit',
        }}>{t('common.cancel')}</button>
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  if (submitError) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '24px 0' }}>
        <div style={{ fontWeight: 700, fontSize: '1rem', color: '#ef4444', marginBottom: 12 }}>{submitError}</div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button onClick={() => submit()} style={{
            padding: '10px 20px', background: '#16a34a', color: '#fff', border: 'none', borderRadius: 10, fontWeight: 600, cursor: 'pointer', fontFamily: 'inherit',
          }}>Retry</button>
          <button onClick={() => reset()} style={{
            padding: '10px 20px', background: '#fff', color: '#475569', border: '1px solid #e2e8f0', borderRadius: 10, fontWeight: 600, cursor: 'pointer', fontFamily: 'inherit',
          }}>{t('common.cancel')}</button>
        </div>
      </div>
    );
  }

  // Success
  const dispatch = result?.dispatch;
  const reservation = result?.reservation;
  const vehicleName = dispatch ? '' : reservation?.vehicle_name || '';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: '20px 0', gap: 12 }}>
      <div style={{
        width: 56, height: 56, borderRadius: '50%', background: '#f0fdf4',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        boxShadow: '0 4px 12px rgba(22,163,74,0.15)',
      }}>
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#16a34a" strokeWidth="2.5"><path d="M20 6L9 17l-5-5" /></svg>
      </div>
      <div style={{ fontWeight: 700, fontSize: '1.05rem', color: '#0f172a' }}>{t('bookingFlow.confirmed')}</div>

      <div style={{ width: '100%', background: '#f8fafc', borderRadius: 12, padding: 14, display: 'flex', flexDirection: 'column', gap: 8 }}>
        {vehicleName && <InfoRow label={t('bookingFlow.vehicleLabel')} value={vehicleName} />}
        {origin && <InfoRow label={t('bookingFlow.origin')} value={origin.name} />}
        {destination && <InfoRow label={t('bookingFlow.destination')} value={destination.name} />}
        {(purpose || purposeCategory) && <InfoRow label={t('bookingFlow.purposeLabel')} value={purpose || purposeCategory || ''} />}
      </div>

      <button onClick={() => reset()} style={{
        width: '100%', padding: '14px', background: '#16a34a', color: '#fff',
        border: 'none', borderRadius: 12, fontWeight: 700, fontSize: '0.95rem',
        cursor: 'pointer', fontFamily: 'inherit',
        boxShadow: '0 4px 16px rgba(22,163,74,0.3)',
      }}>
        {t('bookingFlow.newBooking')}
      </button>
    </div>
  );
}

/* ─── Helpers ─── */
function PanelInput({ label, value, onChange, placeholder }: {
  label: string; value: string; onChange: (v: string) => void; placeholder: string;
}) {
  return (
    <div>
      <label style={{ fontSize: '0.72rem', color: '#94a3b8', fontWeight: 600, marginBottom: 2, display: 'block' }}>{label}</label>
      <input value={value} onChange={e => onChange(e.target.value)} placeholder={placeholder}
        style={{ width: '100%', padding: '7px 10px', border: '1px solid #e2e8f0', borderRadius: 6, fontSize: '0.82rem', fontFamily: 'inherit', outline: 'none', boxSizing: 'border-box' }}
        onFocus={e => e.target.style.borderColor = '#16a34a'} onBlur={e => e.target.style.borderColor = '#e2e8f0'} />
    </div>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <span style={{ fontSize: '0.78rem', color: '#94a3b8' }}>{label}</span>
      <span style={{ fontSize: '0.85rem', fontWeight: 600, color: '#0f172a' }}>{value}</span>
    </div>
  );
}
