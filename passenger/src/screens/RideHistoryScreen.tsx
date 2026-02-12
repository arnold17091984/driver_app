import React, {useEffect} from 'react';
import {
  View,
  Text,
  FlatList,
  StyleSheet,
  ActivityIndicator,
  RefreshControl,
} from 'react-native';
import {useRideStore} from '../stores/rideStore';
import type {Dispatch, DispatchStatus} from '../types';

const STATUS_CONFIG: Record<
  DispatchStatus,
  {label: string; color: string; bg: string}
> = {
  pending: {label: 'Pending', color: '#d97706', bg: '#fffbeb'},
  assigned: {label: 'Assigned', color: '#2563eb', bg: '#eff6ff'},
  accepted: {label: 'Accepted', color: '#16a34a', bg: '#f0fdf4'},
  en_route: {label: 'En Route', color: '#2563eb', bg: '#eff6ff'},
  arrived: {label: 'Arrived', color: '#16a34a', bg: '#f0fdf4'},
  completed: {label: 'Completed', color: '#64748b', bg: '#f8fafc'},
  cancelled: {label: 'Cancelled', color: '#dc2626', bg: '#fef2f2'},
};

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function RideItem({ride}: {ride: Dispatch}) {
  const config = STATUS_CONFIG[ride.status] || STATUS_CONFIG.pending;

  return (
    <View style={styles.card}>
      {/* Header: date + status badge */}
      <View style={styles.cardHeader}>
        <Text style={styles.date}>{formatDate(ride.created_at)}</Text>
        <View style={[styles.statusBadge, {backgroundColor: config.bg}]}>
          <View style={[styles.statusDot, {backgroundColor: config.color}]} />
          <Text style={[styles.statusLabel, {color: config.color}]}>
            {config.label}
          </Text>
        </View>
      </View>

      {/* Route visualization (web-style dots + line) */}
      <View style={styles.routeContainer}>
        <View style={styles.routeDotsColumn}>
          <View style={[styles.routeDot, {backgroundColor: '#16a34a'}]} />
          {ride.dropoff_address && (
            <>
              <View style={styles.routeLine} />
              <View style={[styles.routeDot, {backgroundColor: '#ef4444'}]} />
            </>
          )}
        </View>
        <View style={styles.routeTextsColumn}>
          <Text style={styles.routeAddress} numberOfLines={1}>
            {ride.pickup_address}
          </Text>
          {ride.dropoff_address && (
            <Text
              style={[styles.routeAddress, {marginTop: 14}]}
              numberOfLines={1}>
              {ride.dropoff_address}
            </Text>
          )}
        </View>
      </View>

      {/* Duration/distance info */}
      {ride.estimated_duration_sec != null && (
        <View style={styles.metaRow}>
          <Text style={styles.metaText}>
            {Math.round(ride.estimated_duration_sec / 60)} min
            {ride.estimated_distance_m
              ? ` · ${(ride.estimated_distance_m / 1000).toFixed(1)} km`
              : ''}
          </Text>
        </View>
      )}
    </View>
  );
}

export default function RideHistoryScreen() {
  const rideHistory = useRideStore(s => s.rideHistory);
  const isLoadingHistory = useRideStore(s => s.isLoadingHistory);
  const fetchHistory = useRideStore(s => s.fetchHistory);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  return (
    <View style={styles.container}>
      <FlatList
        data={rideHistory}
        keyExtractor={item => item.id}
        renderItem={({item}) => <RideItem ride={item} />}
        contentContainerStyle={styles.list}
        refreshControl={
          <RefreshControl
            refreshing={isLoadingHistory}
            onRefresh={fetchHistory}
            tintColor="#2563eb"
          />
        }
        ListEmptyComponent={
          isLoadingHistory ? (
            <View style={styles.emptyContainer}>
              <ActivityIndicator size="large" color="#2563eb" />
            </View>
          ) : (
            <View style={styles.emptyContainer}>
              <View style={styles.emptyIcon}>
                <Text style={styles.emptyIconText}>🚕</Text>
              </View>
              <Text style={styles.emptyTitle}>No ride history yet</Text>
              <Text style={styles.emptySubtitle}>
                Your completed rides will appear here
              </Text>
            </View>
          )
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f1f5f9',
  },
  list: {
    padding: 16,
    paddingBottom: 32,
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 10,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 3,
    elevation: 1,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 14,
  },
  date: {
    fontSize: 12,
    color: '#94a3b8',
    fontWeight: '500',
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 6,
    gap: 6,
  },
  statusDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
  },
  statusLabel: {
    fontSize: 11,
    fontWeight: '600',
  },
  routeContainer: {
    flexDirection: 'row',
    gap: 10,
  },
  routeDotsColumn: {
    alignItems: 'center',
    paddingTop: 4,
  },
  routeDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
  },
  routeLine: {
    width: 2,
    height: 14,
    backgroundColor: '#cbd5e1',
    marginVertical: 2,
  },
  routeTextsColumn: {
    flex: 1,
  },
  routeAddress: {
    fontSize: 14,
    fontWeight: '600',
    color: '#0f172a',
  },
  metaRow: {
    marginTop: 10,
    borderTopWidth: 1,
    borderTopColor: '#f1f5f9',
    paddingTop: 10,
  },
  metaText: {
    fontSize: 12,
    color: '#94a3b8',
    fontWeight: '500',
  },
  emptyContainer: {
    alignItems: 'center',
    paddingTop: 80,
  },
  emptyIcon: {
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: '#f0fdf4',
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 16,
    shadowColor: '#16a34a',
    shadowOffset: {width: 0, height: 4},
    shadowOpacity: 0.15,
    shadowRadius: 12,
    elevation: 3,
  },
  emptyIconText: {
    fontSize: 28,
  },
  emptyTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#0f172a',
  },
  emptySubtitle: {
    fontSize: 13,
    color: '#94a3b8',
    marginTop: 4,
  },
});
