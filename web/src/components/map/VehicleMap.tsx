import { useEffect, useRef } from 'react';
import { useGoogleMap } from '../../hooks/useGoogleMap';
import { computeRoute } from '../../api/routes';
import type { RouteResult } from '../../api/routes';
import type { Vehicle, Dispatch } from '../../types/api';
import { vehicleStatusColor, vehicleStatusLabel, dispatchStatusLabel } from '../../utils/formatters';
import { useI18nStore } from '../../stores/i18nStore';

interface Props {
  vehicles: Vehicle[];
  dispatches?: Dispatch[];
  selectedVehicleId: string | null;
  onSelectVehicle: (id: string | null) => void;
  onMapClick?: (lat: number, lng: number) => void;
  pickupMarker?: { lat: number; lng: number } | null;
  bookingRoute?: { origin: { lat: number; lng: number }; destination: { lat: number; lng: number } } | null;
  hideLegend?: boolean;
}

const MANILA_CENTER = { lat: 14.5547, lng: 121.0244 };

const DEMO_OFFSETS: [number, number][] = [
  [0.005, 0.003],
  [-0.003, 0.008],
  [0.008, -0.004],
  [-0.006, -0.006],
  [0.002, 0.012],
];

function vehiclePosition(v: Vehicle, i: number) {
  const offset = DEMO_OFFSETS[i % DEMO_OFFSETS.length];
  return {
    lat: v.latitude ?? MANILA_CENTER.lat + offset[0],
    lng: v.longitude ?? MANILA_CENTER.lng + offset[1],
  };
}

// Distinct colors per vehicle (up to 10)
const VEHICLE_PALETTE = [
  '#2563eb', // blue
  '#dc2626', // red
  '#16a34a', // green
  '#f59e0b', // amber
  '#8b5cf6', // purple
  '#ec4899', // pink
  '#0891b2', // cyan
  '#ea580c', // orange
  '#4f46e5', // indigo
  '#65a30d', // lime
];

// Stable mapping: vehicle ID -> color (persists across re-renders)
const vehicleColorMap = new Map<string, string>();

function getVehicleColor(vehicleId: string, vehicles: Vehicle[]): string {
  if (vehicleColorMap.has(vehicleId)) return vehicleColorMap.get(vehicleId)!;
  const idx = vehicles.findIndex(v => v.id === vehicleId);
  const color = VEHICLE_PALETTE[(idx >= 0 ? idx : vehicleColorMap.size) % VEHICLE_PALETTE.length];
  vehicleColorMap.set(vehicleId, color);
  return color;
}

// Cache route results to avoid redundant API calls on each polling cycle
const routeCache = new Map<string, RouteResult>();

function routeCacheKey(d: Dispatch, vPos: { lat: number; lng: number }): string {
  const r = (n: number) => n.toFixed(4);
  return `${d.id}_${r(vPos.lat)}_${r(vPos.lng)}_${d.status}`;
}

// Decode Google Encoded Polyline into LatLng array
function decodePolyline(encoded: string): google.maps.LatLngLiteral[] {
  const points: google.maps.LatLngLiteral[] = [];
  let index = 0;
  let lat = 0;
  let lng = 0;

  while (index < encoded.length) {
    let shift = 0;
    let result = 0;
    let byte: number;
    do {
      byte = encoded.charCodeAt(index++) - 63;
      result |= (byte & 0x1f) << shift;
      shift += 5;
    } while (byte >= 0x20);
    lat += result & 1 ? ~(result >> 1) : result >> 1;

    shift = 0;
    result = 0;
    do {
      byte = encoded.charCodeAt(index++) - 63;
      result |= (byte & 0x1f) << shift;
      shift += 5;
    } while (byte >= 0x20);
    lng += result & 1 ? ~(result >> 1) : result >> 1;

    points.push({ lat: lat / 1e5, lng: lng / 1e5 });
  }

  return points;
}

export function VehicleMap({ vehicles, dispatches = [], selectedVehicleId, onSelectVehicle, onMapClick, pickupMarker, bookingRoute, hideLegend }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { map, ready, error } = useGoogleMap(containerRef, {
    center: MANILA_CENTER,
    zoom: 14,
    gestureHandling: 'greedy',
    mapTypeControl: false,
    streetViewControl: false,
    fullscreenControl: false,
  });

  const markersRef = useRef<Map<string, google.maps.Marker>>(new Map());
  const infoRef = useRef<google.maps.InfoWindow | null>(null);
  const onSelectRef = useRef(onSelectVehicle);
  useEffect(() => {
    onSelectRef.current = onSelectVehicle;
  }, [onSelectVehicle]);

  // Route overlay refs
  const polylinesRef = useRef<google.maps.Polyline[]>([]);
  const routeMarkersRef = useRef<google.maps.Marker[]>([]);

  // Booking overlay refs
  const pickupMarkerRef = useRef<google.maps.Marker | null>(null);
  const bookingPolyRef = useRef<google.maps.Polyline | null>(null);
  const bookingDestMarkerRef = useRef<google.maps.Marker | null>(null);
  const onMapClickRef = useRef(onMapClick);
  useEffect(() => {
    onMapClickRef.current = onMapClick;
  }, [onMapClick]);

  // Sync vehicle markers
  useEffect(() => {
    if (!map || !ready) return;

    const currentIds = new Set<string>();

    vehicles.forEach((v, i) => {
      currentIds.add(v.id);
      const pos = vehiclePosition(v, i);
      const isSelected = v.id === selectedVehicleId;
      const color = getVehicleColor(v.id, vehicles);
      const label = v.name.charAt(v.name.length - 1);

      let marker = markersRef.current.get(v.id);
      const icon: google.maps.Symbol = {
        path: google.maps.SymbolPath.CIRCLE,
        scale: isSelected ? 18 : 14,
        fillColor: color,
        fillOpacity: 1,
        strokeColor: isSelected ? '#1e293b' : '#fff',
        strokeWeight: isSelected ? 3 : 2,
      };

      if (marker) {
        marker.setPosition(pos);
        marker.setIcon(icon);
        marker.setLabel({ text: label, color: '#fff', fontWeight: '700', fontSize: isSelected ? '14px' : '11px' });
        marker.setZIndex(isSelected ? 100 : 1);
      } else {
        marker = new google.maps.Marker({
          position: pos,
          map,
          icon,
          label: { text: label, color: '#fff', fontWeight: '700', fontSize: isSelected ? '14px' : '11px' },
          zIndex: isSelected ? 100 : 1,
        });
        marker.addListener('click', () => {
          onSelectRef.current(v.id);
          if (!infoRef.current) {
            infoRef.current = new google.maps.InfoWindow();
          }
          infoRef.current.setContent(popupContent(v, vehicles));
          infoRef.current.open(map, marker);
        });
        markersRef.current.set(v.id, marker);
      }
    });

    for (const [id, marker] of markersRef.current) {
      if (!currentIds.has(id)) {
        marker.setMap(null);
        markersRef.current.delete(id);
      }
    }
  }, [map, ready, vehicles, selectedVehicleId]);

  // Draw routes via backend Routes API
  useEffect(() => {
    if (!map || !ready) return;

    // Clear previous
    polylinesRef.current.forEach(p => p.setMap(null));
    polylinesRef.current = [];
    routeMarkersRef.current.forEach(m => m.setMap(null));
    routeMarkersRef.current = [];

    if (dispatches.length === 0) return;

    // Vehicle position lookup
    const vehiclePos = new Map<string, { lat: number; lng: number }>();
    vehicles.forEach((v, i) => {
      vehiclePos.set(v.id, vehiclePosition(v, i));
    });

    for (const d of dispatches) {
      if (!d.vehicle_id || !d.pickup_lat || !d.pickup_lng) continue;
      const vPos = vehiclePos.get(d.vehicle_id);
      if (!vPos) continue;

      const color = getVehicleColor(d.vehicle_id, vehicles);
      const pickup = { lat: d.pickup_lat, lng: d.pickup_lng };
      const hasDropoff = d.dropoff_lat && d.dropoff_lng;
      const dropoff = hasDropoff ? { lat: d.dropoff_lat!, lng: d.dropoff_lng! } : null;

      const origin = vPos;
      const destination = dropoff || pickup;
      const intermediates = dropoff ? [pickup] : [];

      const cacheKey = routeCacheKey(d, vPos);
      const cached = routeCache.get(cacheKey);

      const renderRoute = (result: RouteResult) => {
        // Decode polyline and draw on map
        const path = decodePolyline(result.polyline);
        const polyline = new google.maps.Polyline({
          path,
          map,
          strokeColor: color,
          strokeOpacity: 0.8,
          strokeWeight: 5,
          zIndex: 5,
        });
        polylinesRef.current.push(polyline);

        // Pickup marker
        const pickupMarker = new google.maps.Marker({
          position: pickup,
          map,
          zIndex: 10,
          icon: {
            path: google.maps.SymbolPath.CIRCLE,
            scale: 8,
            fillColor: color,
            fillOpacity: 1,
            strokeColor: '#fff',
            strokeWeight: 2.5,
          },
        });
        pickupMarker.addListener('click', () => {
          if (!infoRef.current) infoRef.current = new google.maps.InfoWindow();
          const leg = result.legs[0];
          infoRef.current.setContent(routePopup(d, 'pickup', leg, color));
          infoRef.current.open(map, pickupMarker);
        });
        routeMarkersRef.current.push(pickupMarker);

        // Dropoff marker
        if (dropoff) {
          const dropoffMarker = new google.maps.Marker({
            position: dropoff,
            map,
            zIndex: 10,
            icon: {
              path: google.maps.SymbolPath.CIRCLE,
              scale: 8,
              fillColor: '#ef4444',
              fillOpacity: 1,
              strokeColor: '#fff',
              strokeWeight: 2.5,
            },
          });
          dropoffMarker.addListener('click', () => {
            if (!infoRef.current) infoRef.current = new google.maps.InfoWindow();
            const leg = result.legs[1] || result.legs[0];
            infoRef.current.setContent(routePopup(d, 'dropoff', leg, color));
            infoRef.current.open(map, dropoffMarker);
          });
          routeMarkersRef.current.push(dropoffMarker);
        }
      };

      if (cached) {
        renderRoute(cached);
      } else {
        computeRoute(origin, destination, intermediates).then(result => {
          routeCache.set(cacheKey, result);
          renderRoute(result);
        }).catch(err => {
          console.error('[Route] Failed to compute route:', err);
        });
      }
    }

    // Prune old cache entries (keep max 20)
    if (routeCache.size > 20) {
      const keys = Array.from(routeCache.keys());
      for (let i = 0; i < keys.length - 20; i++) {
        routeCache.delete(keys[i]);
      }
    }
  }, [map, ready, dispatches, vehicles]);

  // Pan to selected
  useEffect(() => {
    if (!map || !selectedVehicleId) return;
    const marker = markersRef.current.get(selectedVehicleId);
    if (marker) {
      const pos = marker.getPosition();
      if (pos) map.panTo(pos);
    }
  }, [map, selectedVehicleId]);

  // Map click for booking pickup
  useEffect(() => {
    if (!map || !ready) return;
    if (!onMapClickRef.current) return;
    const listener = map.addListener('click', (e: google.maps.MapMouseEvent) => {
      if (e.latLng && onMapClickRef.current) {
        onMapClickRef.current(e.latLng.lat(), e.latLng.lng());
      }
    });
    return () => google.maps.event.removeListener(listener);
  }, [map, ready, onMapClick]);

  // Booking pickup marker
  useEffect(() => {
    if (!map || !ready) return;
    if (pickupMarker) {
      const pos = { lat: pickupMarker.lat, lng: pickupMarker.lng };
      if (pickupMarkerRef.current) {
        pickupMarkerRef.current.setPosition(pos);
      } else {
        pickupMarkerRef.current = new google.maps.Marker({
          position: pos,
          map,
          zIndex: 200,
          icon: {
            path: google.maps.SymbolPath.CIRCLE,
            scale: 12,
            fillColor: '#16a34a',
            fillOpacity: 1,
            strokeColor: '#fff',
            strokeWeight: 3,
          },
        });
      }
      map.panTo(pos);
    } else if (pickupMarkerRef.current) {
      pickupMarkerRef.current.setMap(null);
      pickupMarkerRef.current = null;
    }
  }, [map, ready, pickupMarker]);

  // Booking route polyline
  useEffect(() => {
    if (!map || !ready) return;
    // Clear previous
    if (bookingPolyRef.current) { bookingPolyRef.current.setMap(null); bookingPolyRef.current = null; }
    if (bookingDestMarkerRef.current) { bookingDestMarkerRef.current.setMap(null); bookingDestMarkerRef.current = null; }

    if (!bookingRoute) return;
    const { origin, destination } = bookingRoute;

    // Destination marker
    bookingDestMarkerRef.current = new google.maps.Marker({
      position: destination,
      map,
      zIndex: 200,
      icon: {
        path: google.maps.SymbolPath.CIRCLE,
        scale: 10,
        fillColor: '#ef4444',
        fillOpacity: 1,
        strokeColor: '#fff',
        strokeWeight: 2.5,
      },
    });

    // Route polyline
    computeRoute(origin, destination).then(result => {
      const path = decodePolyline(result.polyline);
      bookingPolyRef.current = new google.maps.Polyline({
        path,
        map,
        strokeColor: '#16a34a',
        strokeOpacity: 0.9,
        strokeWeight: 5,
        zIndex: 100,
      });
      const bounds = new google.maps.LatLngBounds();
      bounds.extend(origin);
      bounds.extend(destination);
      map.fitBounds(bounds, 60);
    }).catch(() => {
      // Fallback: just fit bounds
      const bounds = new google.maps.LatLngBounds();
      bounds.extend(origin);
      bounds.extend(destination);
      map.fitBounds(bounds, 60);
    });
  }, [map, ready, bookingRoute]);

  const { t } = useI18nStore();

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <div ref={containerRef} style={{ width: '100%', height: '100%' }} />
      {error && (
        <div style={{
          position: 'absolute', top: 16, left: 16, right: 16,
          background: '#fef2f2', border: '1px solid #fecaca', borderRadius: 8,
          padding: 16, color: '#991b1b', fontSize: '0.85rem', zIndex: 999,
        }}>
          <strong>Map Error:</strong> {error}
        </div>
      )}
      {vehicles.length > 0 && !hideLegend && (
        <div style={{
          position: 'absolute', bottom: 16, left: 16, zIndex: 1000,
          background: '#fff', borderRadius: 10, padding: '10px 14px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.12)', fontSize: '0.72rem',
        }}>
          <div style={{ fontWeight: 700, marginBottom: 6, color: '#475569' }}>{t('dashboard.vehicles')}</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            {vehicles.map(v => {
              const c = getVehicleColor(v.id, vehicles);
              const activeDispatch = dispatches.find(d => d.vehicle_id === v.id);
              return (
                <div key={v.id} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <div style={{ width: 10, height: 10, borderRadius: '50%', background: c, flexShrink: 0 }} />
                  <span style={{ color: '#1e293b', fontWeight: 600 }}>{v.name}</span>
                  {activeDispatch && (
                    <span style={{
                      padding: '1px 6px', borderRadius: 9999, fontSize: '0.62rem',
                      background: `${c}18`, color: c, fontWeight: 600,
                    }}>
                      {dispatchStatusLabel(activeDispatch.status, t)}
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

function routePopup(d: Dispatch, type: 'pickup' | 'dropoff', leg?: { duration_text: string; distance_text: string }, vehicleColor?: string): string {
  const { t } = useI18nStore.getState();
  const color = vehicleColor || '#2563eb';
  const label = type === 'pickup' ? t('dispatch.pickup') : t('dispatch.destination');
  const address = type === 'pickup' ? d.pickup_address : (d.dropoff_address || '');
  const duration = leg?.duration_text || '';
  const distance = leg?.distance_text || '';

  return `
    <div style="min-width:200px;font-family:'Inter',sans-serif">
      <div style="display:flex;align-items:center;gap:6px;margin-bottom:6px">
        <div style="width:8px;height:8px;border-radius:50%;background:${type === 'pickup' ? color : '#ef4444'}"></div>
        <strong style="font-size:13px;color:#0f172a">${label}</strong>
        <span style="padding:2px 8px;border-radius:9999px;background:${color}18;color:${color};font-weight:600;font-size:10px;margin-left:auto">
          ${dispatchStatusLabel(d.status, t)}
        </span>
      </div>
      <div style="font-size:12px;color:#475569;display:flex;flex-direction:column;gap:3px">
        ${address ? `<div>${address}</div>` : ''}
        <div>${t('dispatch.purpose')}: ${d.purpose}</div>
        ${d.passenger_name ? `<div>${t('dispatch.passengerNameLabel')}: ${d.passenger_name}</div>` : ''}
        ${duration ? `
          <div style="margin-top:4px;padding:6px 10px;background:#f1f5f9;border-radius:6px;display:flex;gap:12px">
            <div><strong style="color:#0f172a">${duration}</strong></div>
            <div><strong style="color:#0f172a">${distance}</strong></div>
          </div>
        ` : ''}
      </div>
    </div>`;
}

function popupContent(v: Vehicle, vehicles: Vehicle[]): string {
  const { t } = useI18nStore.getState();
  const color = getVehicleColor(v.id, vehicles);
  const statusColor = vehicleStatusColor(v.status);
  return `
    <div style="min-width:180px;font-family:'Inter',sans-serif">
      <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
        <div style="width:8px;height:8px;border-radius:50%;background:${color}"></div>
        <strong style="font-size:14px;color:#0f172a">${v.name}</strong>
      </div>
      <div style="font-size:12px;color:#475569;display:flex;flex-direction:column;gap:3px">
        <div>${v.license_plate}</div>
        <div>${v.driver_name}</div>
        <div style="margin-top:4px">
          <span style="padding:2px 10px;border-radius:9999px;background:${statusColor}18;color:${statusColor};font-weight:600;font-size:11px;">
            ${vehicleStatusLabel(v.status, t)}
          </span>
        </div>
      </div>
    </div>`;
}
