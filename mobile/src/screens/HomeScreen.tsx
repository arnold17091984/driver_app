import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Alert,
  RefreshControl,
  ScrollView,
} from 'react-native';
import { useAuthStore } from '../stores/authStore';
import { useTripStore } from '../stores/tripStore';
import client from '../services/apiClient';
import { startTracking, stopTracking } from '../services/locationService';
import type { AttendanceStatus } from '../types';

export function HomeScreen({ navigation }: any) {
  const { user, logout } = useAuthStore();
  const { currentTrip, fetchCurrentTrip } = useTripStore();
  const [isClockedIn, setIsClockedIn] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [attendanceError, setAttendanceError] = useState<string | null>(null);

  const checkAttendance = async () => {
    try {
      const { data } = await client.get<AttendanceStatus>('/attendance/status');
      setIsClockedIn(data.clocked_in);
      setAttendanceError(null);
    } catch {
      setAttendanceError('勤怠情報の取得に失敗しました');
    }
  };

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await Promise.all([checkAttendance(), fetchCurrentTrip()]);
    setRefreshing(false);
  }, [fetchCurrentTrip]);

  useEffect(() => {
    checkAttendance();
    fetchCurrentTrip();
    const interval = setInterval(() => {
      checkAttendance();
      fetchCurrentTrip();
    }, 15000);
    return () => clearInterval(interval);
  }, [fetchCurrentTrip]);

  const handleClockIn = async () => {
    try {
      await client.post('/attendance/clock-in');
      setIsClockedIn(true);
      setAttendanceError(null);
      startTracking();
    } catch (e: any) {
      const msg = e.response?.data?.error?.message || '出勤処理に失敗しました';
      setAttendanceError(msg);
      Alert.alert('エラー', msg);
    }
  };

  const handleClockOut = async () => {
    Alert.alert('退勤確認', '退勤しますか？位置情報の送信が停止します。', [
      { text: 'キャンセル', style: 'cancel' },
      {
        text: '退勤',
        style: 'destructive',
        onPress: async () => {
          try {
            await client.post('/attendance/clock-out');
            setIsClockedIn(false);
            setAttendanceError(null);
            stopTracking();
          } catch (e: any) {
            const msg = e.response?.data?.error?.message || '退勤処理に失敗しました';
            setAttendanceError(msg);
            Alert.alert('エラー', msg);
          }
        },
      },
    ]);
  };

  const handleLogout = () => {
    Alert.alert('ログアウト', 'ログアウトしますか？', [
      { text: 'キャンセル', style: 'cancel' },
      {
        text: 'ログアウト',
        style: 'destructive',
        onPress: () => {
          stopTracking();
          logout();
        },
      },
    ]);
  };

  return (
    <ScrollView
      style={styles.container}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
    >
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.greeting}>おはようございます</Text>
          <Text style={styles.name}>{user?.name}</Text>
        </View>
        <TouchableOpacity onPress={handleLogout} style={styles.logoutBtn}>
          <Text style={styles.logoutText}>ログアウト</Text>
        </TouchableOpacity>
      </View>

      {/* Attendance Error Banner */}
      {attendanceError && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorBannerText}>{attendanceError}</Text>
        </View>
      )}

      {/* Clock In/Out */}
      <TouchableOpacity
        style={[styles.clockButton, isClockedIn ? styles.clockOutButton : styles.clockInButton]}
        onPress={isClockedIn ? handleClockOut : handleClockIn}
      >
        <Text style={styles.clockButtonIcon}>{isClockedIn ? '🔴' : '🟢'}</Text>
        <Text style={styles.clockButtonText}>{isClockedIn ? '退勤' : '出勤'}</Text>
        <Text style={styles.clockButtonSub}>
          {isClockedIn ? '位置情報送信中' : 'タップして出勤'}
        </Text>
      </TouchableOpacity>

      {/* Current Trip */}
      {currentTrip && (
        <TouchableOpacity
          style={styles.tripCard}
          onPress={() => navigation.navigate('TripDetail', { tripId: currentTrip.id })}
        >
          <View style={styles.tripHeader}>
            <Text style={styles.tripLabel}>現在の配車</Text>
            <View style={[styles.statusBadge, { backgroundColor: statusColor(currentTrip.status) }]}>
              <Text style={styles.statusText}>{statusLabel(currentTrip.status)}</Text>
            </View>
          </View>
          <Text style={styles.tripPurpose}>{currentTrip.purpose}</Text>
          <Text style={styles.tripAddress}>📍 {currentTrip.pickup_address}</Text>
          {currentTrip.dropoff_address && (
            <Text style={styles.tripAddress}>🏁 {currentTrip.dropoff_address}</Text>
          )}
          {(currentTrip.estimated_distance_m != null || currentTrip.estimated_duration_sec != null) && (
            <View style={styles.etaRow}>
              {currentTrip.estimated_distance_m != null && (
                <Text style={styles.etaText}>
                  🚗 {currentTrip.estimated_distance_m >= 1000
                    ? `${(currentTrip.estimated_distance_m / 1000).toFixed(1)} km`
                    : `${currentTrip.estimated_distance_m} m`}
                </Text>
              )}
              {currentTrip.estimated_duration_sec != null && (
                <Text style={styles.etaText}>
                  ⏱ {currentTrip.estimated_duration_sec >= 3600
                    ? `${Math.floor(currentTrip.estimated_duration_sec / 3600)}h${Math.floor((currentTrip.estimated_duration_sec % 3600) / 60)}m`
                    : `${Math.ceil(currentTrip.estimated_duration_sec / 60)} min`}
                </Text>
              )}
            </View>
          )}
          <Text style={styles.tapHint}>タップして詳細を表示 →</Text>
        </TouchableOpacity>
      )}

      {!currentTrip && isClockedIn && (
        <View style={styles.noTrip}>
          <Text style={styles.noTripText}>現在の配車はありません</Text>
          <Text style={styles.noTripSub}>新しい配車が割り当てられると通知されます</Text>
        </View>
      )}
    </ScrollView>
  );
}

function statusLabel(status: string): string {
  const labels: Record<string, string> = {
    assigned: '割当済',
    accepted: '受領済',
    en_route: '移動中',
    arrived: '到着',
  };
  return labels[status] || status;
}

function statusColor(status: string): string {
  const colors: Record<string, string> = {
    assigned: '#f59e0b',
    accepted: '#3b82f6',
    en_route: '#8b5cf6',
    arrived: '#22c55e',
  };
  return colors[status] || '#6b7280';
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f1f5f9' },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 20,
    paddingTop: 60,
    backgroundColor: '#1e293b',
  },
  greeting: { color: '#94a3b8', fontSize: 14 },
  name: { color: '#fff', fontSize: 20, fontWeight: '700' },
  logoutBtn: { padding: 8 },
  logoutText: { color: '#94a3b8', fontSize: 14 },
  clockButton: {
    margin: 20,
    padding: 32,
    borderRadius: 16,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.15,
    shadowRadius: 12,
    elevation: 6,
  },
  clockInButton: { backgroundColor: '#22c55e' },
  clockOutButton: { backgroundColor: '#ef4444' },
  clockButtonIcon: { fontSize: 40, marginBottom: 8 },
  clockButtonText: { color: '#fff', fontSize: 24, fontWeight: '700' },
  clockButtonSub: { color: 'rgba(255,255,255,0.8)', fontSize: 14, marginTop: 4 },
  tripCard: {
    margin: 20,
    marginTop: 0,
    padding: 20,
    backgroundColor: '#fff',
    borderRadius: 12,
    borderLeftWidth: 4,
    borderLeftColor: '#3b82f6',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
    elevation: 3,
  },
  tripHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 },
  tripLabel: { fontSize: 13, fontWeight: '600', color: '#64748b', textTransform: 'uppercase' },
  statusBadge: { paddingHorizontal: 8, paddingVertical: 2, borderRadius: 12 },
  statusText: { color: '#fff', fontSize: 12, fontWeight: '600' },
  tripPurpose: { fontSize: 18, fontWeight: '600', color: '#1e293b', marginBottom: 8 },
  tripAddress: { fontSize: 14, color: '#475569', marginBottom: 4 },
  etaRow: {
    flexDirection: 'row',
    gap: 16,
    marginTop: 8,
    paddingTop: 8,
    borderTopWidth: 1,
    borderTopColor: '#e2e8f0',
  },
  etaText: {
    fontSize: 14,
    color: '#475569',
    fontWeight: '500',
  },
  tapHint: { fontSize: 13, color: '#3b82f6', marginTop: 8, fontWeight: '500' },
  noTrip: {
    margin: 20,
    marginTop: 0,
    padding: 32,
    backgroundColor: '#fff',
    borderRadius: 12,
    alignItems: 'center',
  },
  noTripText: { fontSize: 16, color: '#64748b', fontWeight: '500' },
  noTripSub: { fontSize: 13, color: '#94a3b8', marginTop: 4 },
  errorBanner: {
    marginHorizontal: 20,
    marginTop: 12,
    padding: 12,
    backgroundColor: '#fef3c7',
    borderRadius: 8,
    borderLeftWidth: 4,
    borderLeftColor: '#f59e0b',
  },
  errorBannerText: { color: '#92400e', fontSize: 14, fontWeight: '500' },
});
