import { format, formatDistanceToNow } from 'date-fns';
import { ja, ko, zhCN } from 'date-fns/locale';
import type { Locale } from 'date-fns';
import type { TranslationKey } from '../i18n';

type TFn = (key: TranslationKey, params?: Record<string, string | number>) => string;

const dateFnsLocales: Record<string, Locale> = { ja, ko, zh: zhCN };

export function formatDateTime(iso: string, locale?: string): string {
  return format(new Date(iso), 'MMM d, yyyy HH:mm', {
    locale: locale ? dateFnsLocales[locale] : undefined,
  });
}

export function formatTime(iso: string): string {
  return format(new Date(iso), 'HH:mm');
}

export function formatRelative(iso: string, locale?: string): string {
  return formatDistanceToNow(new Date(iso), {
    addSuffix: true,
    locale: locale ? dateFnsLocales[locale] : undefined,
  });
}

export function formatDuration(seconds: number, t?: TFn): string {
  const min = Math.round(seconds / 60);
  if (min < 60) return t ? t('format.minutesShort', { min }) : `${min} min`;
  const h = Math.floor(min / 60);
  const m = min % 60;
  if (m > 0) return t ? t('format.hoursMinutes', { h, m }) : `${h}h ${m}m`;
  return t ? t('format.hoursOnly', { h }) : `${h}h`;
}

export function formatDistance(meters: number, t?: TFn): string {
  if (meters < 1000) return t ? t('format.meters', { m: meters }) : `${meters}m`;
  const km = (meters / 1000).toFixed(1);
  return t ? t('format.kilometers', { km }) : `${km} km`;
}

export function vehicleStatusLabel(status: string, t?: TFn): string {
  if (t) return t(`vehicleStatus.${status}` as TranslationKey);
  const labels: Record<string, string> = {
    available: 'Available',
    waiting: 'Waiting',
    driver_absent: 'Driver Absent',
    reserved: 'Reserved',
    in_trip: 'In Trip',
    maintenance: 'Maintenance',
    stale_location: 'No Signal',
  };
  return labels[status] || status;
}

export function vehicleStatusColor(status: string): string {
  const colors: Record<string, string> = {
    available: '#16a34a',
    waiting: '#0891b2',
    driver_absent: '#94a3b8',
    reserved: '#d97706',
    in_trip: '#2563eb',
    maintenance: '#dc2626',
    stale_location: '#eab308',
  };
  return colors[status] || '#6b7280';
}

export function dispatchStatusLabel(status: string, t?: TFn): string {
  if (t) return t(`dispatchStatus.${status}` as TranslationKey);
  const labels: Record<string, string> = {
    pending: 'Pending',
    assigned: 'Assigned',
    accepted: 'Accepted',
    en_route: 'En Route',
    arrived: 'Arrived',
    completed: 'Completed',
    cancelled: 'Cancelled',
  };
  return labels[status] || status;
}

export function reservationStatusLabel(status: string, t?: TFn): string {
  if (t) return t(`reservationStatus.${status}` as TranslationKey);
  const labels: Record<string, string> = {
    confirmed: 'Confirmed',
    pending_conflict: 'Conflict',
    pending_driver: 'Awaiting Driver',
    driver_declined: 'Declined',
    cancelled: 'Cancelled',
    completed: 'Completed',
  };
  return labels[status] || status;
}

export function reservationStatusColor(status: string): string {
  const colors: Record<string, string> = {
    confirmed: '#16a34a',
    pending_conflict: '#ef4444',
    pending_driver: '#f59e0b',
    driver_declined: '#dc2626',
    cancelled: '#6b7280',
    completed: '#3b82f6',
  };
  return colors[status] || '#6b7280';
}
