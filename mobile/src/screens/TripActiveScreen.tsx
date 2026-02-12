import React, { useEffect } from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Alert } from 'react-native';
import { useTripStore } from '../stores/tripStore';
import type { DispatchStatus } from '../types';

const STEPS: { status: DispatchStatus; label: string }[] = [
  { status: 'accepted', label: '受領' },
  { status: 'en_route', label: '移動中' },
  { status: 'arrived', label: '到着' },
  { status: 'completed', label: '完了' },
];

export function TripActiveScreen({ navigation }: any) {
  const { currentTrip, fetchCurrentTrip, startEnRoute, markArrived, completeTrip } = useTripStore();

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

  const currentStepIndex = STEPS.findIndex((s) => s.status === currentTrip.status);

  const handleNext = async () => {
    try {
      switch (currentTrip.status) {
        case 'accepted':
          await startEnRoute(currentTrip.id);
          break;
        case 'en_route':
          await markArrived(currentTrip.id);
          break;
        case 'arrived':
          Alert.alert('完了確認', '配車を完了しますか？', [
            { text: 'キャンセル', style: 'cancel' },
            {
              text: '完了',
              onPress: async () => {
                await completeTrip(currentTrip.id);
                navigation.navigate('Home');
              },
            },
          ]);
          return;
      }
    } catch {
      Alert.alert('エラー', 'ステータス更新に失敗しました');
    }
  };

  const getNextAction = (): { label: string; color: string } => {
    switch (currentTrip.status) {
      case 'accepted':
        return { label: '出発する', color: '#8b5cf6' };
      case 'en_route':
        return { label: '到着した', color: '#22c55e' };
      case 'arrived':
        return { label: '完了する', color: '#3b82f6' };
      default:
        return { label: '', color: '#6b7280' };
    }
  };

  const nextAction = getNextAction();

  return (
    <View style={styles.container}>
      {/* Progress Stepper */}
      <View style={styles.stepper}>
        {STEPS.map((step, index) => (
          <View key={step.status} style={styles.stepItem}>
            <View
              style={[
                styles.stepCircle,
                index <= currentStepIndex ? styles.stepActive : styles.stepInactive,
              ]}
            >
              <Text style={styles.stepNumber}>{index + 1}</Text>
            </View>
            <Text
              style={[
                styles.stepLabel,
                index <= currentStepIndex ? styles.stepLabelActive : styles.stepLabelInactive,
              ]}
            >
              {step.label}
            </Text>
            {index < STEPS.length - 1 && (
              <View
                style={[
                  styles.stepLine,
                  index < currentStepIndex ? styles.stepLineActive : styles.stepLineInactive,
                ]}
              />
            )}
          </View>
        ))}
      </View>

      {/* Trip Info */}
      <View style={styles.infoCard}>
        <Text style={styles.purpose}>{currentTrip.purpose}</Text>
        <View style={styles.addressRow}>
          <Text style={styles.addressIcon}>📍</Text>
          <Text style={styles.addressText}>{currentTrip.pickup_address}</Text>
        </View>
        {currentTrip.dropoff_address && (
          <View style={styles.addressRow}>
            <Text style={styles.addressIcon}>🏁</Text>
            <Text style={styles.addressText}>{currentTrip.dropoff_address}</Text>
          </View>
        )}
        {currentTrip.passenger_name && (
          <Text style={styles.passenger}>
            乗客: {currentTrip.passenger_name}
          </Text>
        )}
      </View>

      {/* Action Button */}
      {nextAction.label && (
        <TouchableOpacity
          style={[styles.actionButton, { backgroundColor: nextAction.color }]}
          onPress={handleNext}
        >
          <Text style={styles.actionButtonText}>{nextAction.label}</Text>
        </TouchableOpacity>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f1f5f9', padding: 20 },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  emptyText: { color: '#94a3b8', fontSize: 16 },
  stepper: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 20,
    marginBottom: 20,
  },
  stepItem: { alignItems: 'center', flex: 1 },
  stepCircle: {
    width: 36,
    height: 36,
    borderRadius: 18,
    justifyContent: 'center',
    alignItems: 'center',
  },
  stepActive: { backgroundColor: '#3b82f6' },
  stepInactive: { backgroundColor: '#e2e8f0' },
  stepNumber: { color: '#fff', fontWeight: '700', fontSize: 14 },
  stepLabel: { fontSize: 11, marginTop: 4, fontWeight: '500' },
  stepLabelActive: { color: '#3b82f6' },
  stepLabelInactive: { color: '#94a3b8' },
  stepLine: {
    position: 'absolute',
    top: 18,
    left: '60%',
    right: '-40%',
    height: 2,
  },
  stepLineActive: { backgroundColor: '#3b82f6' },
  stepLineInactive: { backgroundColor: '#e2e8f0' },
  infoCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 20,
    marginBottom: 20,
  },
  purpose: { fontSize: 20, fontWeight: '700', color: '#1e293b', marginBottom: 16 },
  addressRow: { flexDirection: 'row', alignItems: 'center', marginBottom: 8 },
  addressIcon: { fontSize: 16, marginRight: 8 },
  addressText: { fontSize: 15, color: '#475569', flex: 1 },
  passenger: { fontSize: 14, color: '#64748b', marginTop: 8 },
  actionButton: {
    borderRadius: 16,
    padding: 20,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 6,
  },
  actionButtonText: { color: '#fff', fontSize: 20, fontWeight: '700' },
});
