import { useState, useCallback } from 'react';
import { useVehicleStore } from '../stores/vehicleStore';
import { useBookingStore } from '../stores/bookingStore';
import { useI18nStore } from '../stores/i18nStore';
import { usePolling } from '../hooks/usePolling';
import { useIsMobile } from '../hooks/useIsMobile';
import { usePermission } from '../hooks/usePermission';
import { VehicleList } from '../components/vehicle/VehicleList';
import { VehicleMap } from '../components/map/VehicleMap';
import { BookingOverlay } from '../components/dashboard/BookingOverlay';
import { BottomSheet } from '../components/common/BottomSheet';
import { listDispatches } from '../api/dispatches';
import { useNotificationSound } from '../hooks/useNotificationSound';
import type { Dispatch } from '../types/api';

const ACTIVE_STATUSES = ['assigned', 'accepted', 'en_route', 'arrived'];

export function DashboardPage() {
  const { vehicles, selectedVehicleId, fetchVehicles, selectVehicle } = useVehicleStore();
  const { t } = useI18nStore();
  const isMobile = useIsMobile();
  const canBook = usePermission('admin', 'dispatcher');
  const [dispatches, setDispatches] = useState<Dispatch[]>([]);
  const { checkForNew: checkNewDispatches } = useNotificationSound<Dispatch>('dispatch');

  // Booking state
  const bookingStep = useBookingStore(s => s.step);
  const bookingOrigin = useBookingStore(s => s.origin);
  const bookingDest = useBookingStore(s => s.destination);
  const setOrigin = useBookingStore(s => s.setOrigin);
  const setStep = useBookingStore(s => s.setStep);
  const isBooking = bookingStep !== 'idle';

  const fetchAll = useCallback(async () => {
    const [, allDispatches] = await Promise.all([
      fetchVehicles(),
      listDispatches(),
    ]);
    const active = (allDispatches || []).filter(d => ACTIVE_STATUSES.includes(d.status));
    setDispatches(active);
    checkNewDispatches(active);
  }, [fetchVehicles, checkNewDispatches]);

  usePolling(fetchAll, 10000);

  const availCount = vehicles.filter((v) => v.status === 'available').length;

  // Map click handler — only active during pickup-map step
  const handleMapClick = bookingStep === 'pickup-map'
    ? (lat: number, lng: number) => {
        setOrigin({
          name: t('bookingFlow.currentLocation'),
          address: `${lat.toFixed(6)}, ${lng.toFixed(6)}`,
          lat,
          lng,
        });
      }
    : undefined;

  const pickupMarker = bookingOrigin ? { lat: bookingOrigin.lat, lng: bookingOrigin.lng } : null;
  const bookingRoute = bookingOrigin && bookingDest && bookingDest.lat !== 0
    ? { origin: { lat: bookingOrigin.lat, lng: bookingOrigin.lng }, destination: { lat: bookingDest.lat, lng: bookingDest.lng } }
    : null;

  const statsOverlay = (
    <div style={{
      position: 'absolute', top: isMobile ? 12 : 16, left: isMobile ? 12 : 16, zIndex: 10,
      display: 'flex', gap: isMobile ? 6 : 8, flexWrap: 'wrap',
    }}>
      <div style={{
        background: '#fff', borderRadius: 10, padding: isMobile ? '8px 12px' : '10px 16px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
        display: 'flex', alignItems: 'center', gap: 8,
      }}>
        <span style={{ fontSize: '0.75rem', color: '#64748b', fontWeight: 500 }}>{t('dashboard.total')}</span>
        <span style={{ fontSize: '1.1rem', fontWeight: 700, color: '#0f172a' }}>{vehicles.length}</span>
      </div>
      <div style={{
        background: '#fff', borderRadius: 10, padding: isMobile ? '8px 12px' : '10px 16px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
        display: 'flex', alignItems: 'center', gap: 8,
      }}>
        <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#16a34a' }} />
        <span style={{ fontSize: '0.75rem', color: '#64748b', fontWeight: 500 }}>{t('dashboard.available')}</span>
        <span style={{ fontSize: '1.1rem', fontWeight: 700, color: '#16a34a' }}>{availCount}</span>
      </div>
      {dispatches.length > 0 && (
        <div style={{
          background: '#fff', borderRadius: 10, padding: isMobile ? '8px 12px' : '10px 16px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
          display: 'flex', alignItems: 'center', gap: 8,
        }}>
          <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#2563eb' }} />
          <span style={{ fontSize: '0.75rem', color: '#64748b', fontWeight: 500 }}>{t('dashboard.activeTrips')}</span>
          <span style={{ fontSize: '1.1rem', fontWeight: 700, color: '#2563eb' }}>{dispatches.length}</span>
        </div>
      )}
    </div>
  );

  // Booking FAB button
  const bookingFab = canBook && !isBooking ? (
    <button
      onClick={() => setStep('destination')}
      style={{
        width: '100%', padding: '14px',
        background: '#16a34a', color: '#fff',
        border: 'none', borderRadius: 12,
        cursor: 'pointer', fontWeight: 700, fontSize: '0.95rem',
        fontFamily: 'inherit',
        boxShadow: '0 4px 16px rgba(22,163,74,0.3)',
        display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
        marginBottom: 12,
      }}
    >
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z" />
        <circle cx="12" cy="10" r="3" />
      </svg>
      {t('bookingFlow.title')}
    </button>
  ) : null;

  // Panel content: booking overlay or vehicle list
  const panelContent = isBooking ? (
    <BookingOverlay />
  ) : (
    <>
      {bookingFab}
      <VehicleList
        vehicles={vehicles}
        dispatches={dispatches}
        selectedId={selectedVehicleId}
        onSelect={selectVehicle}
      />
    </>
  );

  if (isMobile) {
    return (
      <div style={{ height: '100%', position: 'relative' }}>
        <VehicleMap
          vehicles={vehicles}
          dispatches={dispatches}
          selectedVehicleId={isBooking ? null : selectedVehicleId}
          onSelectVehicle={selectVehicle}
          onMapClick={handleMapClick}
          pickupMarker={isBooking ? pickupMarker : null}
          bookingRoute={isBooking ? bookingRoute : null}
          hideLegend
        />
        {!isBooking && statsOverlay}
        <BottomSheet
          peekHeight={isBooking ? 240 : 100}
          maxHeight={isBooking ? '75vh' : '60vh'}
          open={isBooking}
          fitContent={isBooking}
          header={
            isBooking ? undefined : (
              <div>
                <div style={{ fontWeight: 700, fontSize: '0.95rem' }}>{t('dashboard.vehicles')}</div>
                <div style={{ fontSize: '0.72rem', color: '#94a3b8' }}>
                  {t('dashboard.vehiclesTracked', { count: vehicles.length })}
                </div>
              </div>
            )
          }
        >
          {panelContent}
        </BottomSheet>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', height: '100%' }}>
      {/* Map area */}
      <div style={{ flex: 1, position: 'relative' }}>
        <VehicleMap
          vehicles={vehicles}
          dispatches={dispatches}
          selectedVehicleId={isBooking ? null : selectedVehicleId}
          onSelectVehicle={selectVehicle}
          onMapClick={handleMapClick}
          pickupMarker={isBooking ? pickupMarker : null}
          bookingRoute={isBooking ? bookingRoute : null}
        />
        {!isBooking && statsOverlay}
      </div>

      {/* Side panel */}
      <div style={{
        width: 380, borderLeft: '1px solid #e2e8f0',
        background: '#fff', display: 'flex', flexDirection: 'column',
      }}>
        <div style={{
          padding: '18px 20px', borderBottom: '1px solid #f1f5f9',
        }}>
          <h2 style={{ margin: 0, fontSize: '1rem', fontWeight: 700, color: '#0f172a' }}>
            {isBooking ? t('bookingFlow.title') : t('dashboard.vehicles')}
          </h2>
          {!isBooking && (
            <p style={{ margin: '2px 0 0', fontSize: '0.75rem', color: '#94a3b8' }}>
              {t('dashboard.vehiclesTracked', { count: vehicles.length })}
            </p>
          )}
        </div>
        <div style={{ flex: 1, overflow: 'auto', padding: '12px 16px' }}>
          {panelContent}
        </div>
      </div>
    </div>
  );
}
