import React, { useEffect } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert, Linking, Platform } from 'react-native';
import { useTripStore } from '../stores/tripStore';

export function TripDetailScreen({ route, navigation }: any) {
  const { currentTrip, fetchCurrentTrip, acceptTrip } = useTripStore();

  useEffect(() => {
    fetchCurrentTrip();
  }, [fetchCurrentTrip]);

  if (!currentTrip) {
    return (
      <View style={styles.center}>
        <Text style={styles.emptyText}>配車データが見つかりません</Text>
      </View>
    );
  }

  const handleAccept = async () => {
    try {
      await acceptTrip(currentTrip.id);
      navigation.navigate('TripActive', { tripId: currentTrip.id });
    } catch {
      Alert.alert('エラー', '受領処理に失敗しました');
    }
  };

  const openMaps = () => {
    if (!currentTrip.pickup_lat || !currentTrip.pickup_lng) {
      Alert.alert('エラー', '位置情報がありません');
      return;
    }
    const lat = currentTrip.pickup_lat;
    const lng = currentTrip.pickup_lng;
    const url = Platform.select({
      ios: `maps:0,0?q=${lat},${lng}`,
      android: `geo:${lat},${lng}?q=${lat},${lng}`,
    });
    if (url) Linking.openURL(url);
  };

  return (
    <View style={styles.container}>
      <View style={styles.card}>
        <Text style={styles.label}>目的</Text>
        <Text style={styles.value}>{currentTrip.purpose}</Text>

        <Text style={styles.label}>ピックアップ</Text>
        <Text style={styles.value}>{currentTrip.pickup_address}</Text>

        {currentTrip.dropoff_address && (
          <>
            <Text style={styles.label}>目的地</Text>
            <Text style={styles.value}>{currentTrip.dropoff_address}</Text>
          </>
        )}

        {currentTrip.passenger_name && (
          <>
            <Text style={styles.label}>乗客</Text>
            <Text style={styles.value}>
              {currentTrip.passenger_name} ({currentTrip.passenger_count}名)
            </Text>
          </>
        )}

        {currentTrip.notes && (
          <>
            <Text style={styles.label}>メモ</Text>
            <Text style={styles.value}>{currentTrip.notes}</Text>
          </>
        )}

        {currentTrip.estimated_duration_sec && (
          <>
            <Text style={styles.label}>予想到着時間</Text>
            <Text style={styles.value}>
              {Math.round(currentTrip.estimated_duration_sec / 60)}分
            </Text>
          </>
        )}
      </View>

      <View style={styles.actions}>
        <TouchableOpacity style={styles.mapsButton} onPress={openMaps}>
          <Text style={styles.mapsButtonText}>地図で開く</Text>
        </TouchableOpacity>

        {currentTrip.status === 'assigned' && (
          <TouchableOpacity style={styles.acceptButton} onPress={handleAccept}>
            <Text style={styles.acceptButtonText}>受領する</Text>
          </TouchableOpacity>
        )}

        {['accepted', 'en_route', 'arrived'].includes(currentTrip.status) && (
          <TouchableOpacity
            style={styles.activeButton}
            onPress={() => navigation.navigate('TripActive', { tripId: currentTrip.id })}
          >
            <Text style={styles.activeButtonText}>進行状況</Text>
          </TouchableOpacity>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f1f5f9', padding: 20 },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  emptyText: { color: '#94a3b8', fontSize: 16 },
  card: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 20,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
    elevation: 3,
  },
  label: { fontSize: 12, fontWeight: '600', color: '#94a3b8', marginTop: 16, textTransform: 'uppercase' },
  value: { fontSize: 16, color: '#1e293b', marginTop: 4 },
  actions: { marginTop: 20, gap: 12 },
  mapsButton: {
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#3b82f6',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
  },
  mapsButtonText: { color: '#3b82f6', fontSize: 16, fontWeight: '600' },
  acceptButton: {
    backgroundColor: '#22c55e',
    borderRadius: 12,
    padding: 18,
    alignItems: 'center',
  },
  acceptButtonText: { color: '#fff', fontSize: 18, fontWeight: '700' },
  activeButton: {
    backgroundColor: '#3b82f6',
    borderRadius: 12,
    padding: 18,
    alignItems: 'center',
  },
  activeButtonText: { color: '#fff', fontSize: 18, fontWeight: '700' },
});
