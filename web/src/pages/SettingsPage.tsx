import { useState, useEffect, useRef } from 'react';
import { listUsers, updateUserRole, updateUserPriority } from '../api/admin';
import { listVehicles, createVehicle, updateVehicle, deleteVehicle, uploadVehiclePhoto } from '../api/vehicles';
import type { User, Vehicle } from '../types/api';
import { useI18nStore } from '../stores/i18nStore';
import { useIsMobile } from '../hooks/useIsMobile';
import { ResponsiveTable } from '../components/common/ResponsiveTable';
import { vehicleStatusLabel, vehicleStatusColor } from '../utils/formatters';
import type { TranslationKey } from '../i18n';

const API_BASE = import.meta.env.VITE_API_BASE || '';

const roleBadge: Record<string, { bg: string; color: string }> = {
  admin: { bg: '#ede9fe', color: '#6d28d9' },
  dispatcher: { bg: '#dbeafe', color: '#1e40af' },
  viewer: { bg: '#f1f5f9', color: '#475569' },
  driver: { bg: '#d1fae5', color: '#065f46' },
};

type Tab = 'users' | 'vehicles';

export function SettingsPage() {
  const { t } = useI18nStore();
  const isMobile = useIsMobile();
  const [tab, setTab] = useState<Tab>('vehicles');

  return (
    <div style={{ padding: isMobile ? 16 : 28 }}>
      <div style={{ marginBottom: isMobile ? 16 : 24 }}>
        <h1 style={{ margin: 0, fontSize: '1.25rem', fontWeight: 700, color: '#0f172a' }}>{t('settings.title')}</h1>
        <p style={{ margin: '2px 0 0', fontSize: '0.8rem', color: '#94a3b8' }}>
          {t('settings.subtitle')}
        </p>
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 4, marginBottom: 20, background: '#f1f5f9', borderRadius: 10, padding: 4, width: 'fit-content' }}>
        {([['vehicles', t('settings.vehiclesTab')], ['users', t('settings.usersTab')]] as [Tab, string][]).map(([key, label]) => (
          <button key={key} onClick={() => setTab(key)} style={{
            padding: '8px 20px', border: 'none', borderRadius: 8, cursor: 'pointer',
            fontWeight: 600, fontSize: '0.85rem', fontFamily: 'inherit',
            background: tab === key ? '#fff' : 'transparent',
            color: tab === key ? '#0f172a' : '#64748b',
            boxShadow: tab === key ? '0 1px 3px rgba(0,0,0,0.1)' : 'none',
          }}>{label}</button>
        ))}
      </div>

      {tab === 'users' && <UserManagement />}
      {tab === 'vehicles' && <VehicleManagement />}
    </div>
  );
}

// --- Vehicle Management ---

function VehicleManagement() {
  const { t } = useI18nStore();
  const isMobile = useIsMobile();
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState<string | null>(null);
  const [form, setForm] = useState({ name: '', license_plate: '', driver_id: '' });
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploadingId, setUploadingId] = useState<string | null>(null);

  const fetchData = async () => {
    const [v, u] = await Promise.all([listVehicles(), listUsers()]);
    setVehicles(v || []);
    setUsers(u || []);
  };

  useEffect(() => { fetchData(); }, []);

  const drivers = users.filter(u => u.role === 'driver');
  const assignedDriverIds = new Set(vehicles.map(v => v.driver_id));

  const openCreate = () => {
    setEditId(null);
    setForm({ name: '', license_plate: '', driver_id: '' });
    setShowForm(true);
  };

  const openEdit = (v: Vehicle) => {
    setEditId(v.id);
    setForm({ name: v.name, license_plate: v.license_plate, driver_id: v.driver_id });
    setShowForm(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (editId) {
      await updateVehicle(editId, form);
    } else {
      await createVehicle(form);
    }
    setShowForm(false);
    fetchData();
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t('settings.confirmDeleteVehicle'))) return;
    await deleteVehicle(id);
    fetchData();
  };

  const handlePhotoClick = (vehicleId: string) => {
    setUploadingId(vehicleId);
    fileRef.current?.click();
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !uploadingId) return;
    await uploadVehiclePhoto(uploadingId, file);
    setUploadingId(null);
    e.target.value = '';
    fetchData();
  };

  return (
    <>
      <input ref={fileRef} type="file" accept="image/*" style={{ display: 'none' }} onChange={handleFileChange} />

      {/* Form Modal */}
      {showForm && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.4)', zIndex: 1000,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }} onClick={() => setShowForm(false)}>
          <form onSubmit={handleSubmit} onClick={e => e.stopPropagation()} style={{
            background: '#fff', borderRadius: 16, padding: isMobile ? 20 : 28, width: isMobile ? '90vw' : 420,
            boxShadow: '0 20px 60px rgba(0,0,0,0.2)',
          }}>
            <h2 style={{ margin: '0 0 20px', fontSize: '1.1rem', fontWeight: 700 }}>
              {editId ? t('settings.editVehicle') : t('settings.addVehicle')}
            </h2>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <div>
                <label style={labelStyle}>{t('settings.vehicleNameLabel')} *</label>
                <input required value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder={t('settings.vehicleNamePlaceholder')} />
              </div>
              <div>
                <label style={labelStyle}>{t('settings.licensePlateLabel')} *</label>
                <input required value={form.license_plate} onChange={e => setForm(f => ({ ...f, license_plate: e.target.value }))}
                  placeholder={t('settings.licensePlatePlaceholder')} />
              </div>
              <div>
                <label style={labelStyle}>{t('settings.assignedDriverLabel')} *</label>
                <select required value={form.driver_id} onChange={e => setForm(f => ({ ...f, driver_id: e.target.value }))}>
                  <option value="">{t('settings.selectDriver')}</option>
                  {drivers.map(d => (
                    <option key={d.id} value={d.id} disabled={assignedDriverIds.has(d.id) && d.id !== form.driver_id}>
                      {d.name} ({d.employee_id}){assignedDriverIds.has(d.id) && d.id !== form.driver_id ? ` ${t('settings.driverAssigned')}` : ''}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <div style={{ display: 'flex', gap: 10, marginTop: 24, justifyContent: 'flex-end' }}>
              <button type="button" onClick={() => setShowForm(false)} style={{
                padding: '10px 20px', background: '#f1f5f9', color: '#475569',
                border: 'none', borderRadius: 8, cursor: 'pointer', fontWeight: 600, fontFamily: 'inherit',
              }}>{t('common.cancel')}</button>
              <button type="submit" style={{
                padding: '10px 24px', background: '#2563eb', color: '#fff',
                border: 'none', borderRadius: 8, cursor: 'pointer', fontWeight: 600, fontFamily: 'inherit',
              }}>{editId ? t('settings.saveChanges') : t('settings.addVehicle')}</button>
            </div>
          </form>
        </div>
      )}

      <div style={{
        background: '#fff', borderRadius: 12,
        border: '1px solid #e2e8f0', overflow: 'hidden',
        boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
      }}>
        <div style={{
          padding: '14px 20px', borderBottom: '1px solid #e2e8f0',
          background: '#f8fafc', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        }}>
          <h2 style={{ margin: 0, fontSize: '0.9rem', fontWeight: 600, color: '#475569' }}>
            {t('settings.vehicleManagement')}
            <span style={{ marginLeft: 8, fontSize: '0.8rem', fontWeight: 400, color: '#94a3b8' }}>
              {t('settings.vehicleCount', { count: vehicles.length })}
            </span>
          </h2>
          <button onClick={openCreate} style={{
            padding: '7px 16px', background: '#2563eb', color: '#fff',
            border: 'none', borderRadius: 8, cursor: 'pointer',
            fontWeight: 600, fontSize: '0.82rem', fontFamily: 'inherit',
            display: 'flex', alignItems: 'center', gap: 6,
          }}>
            <svg width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14" /></svg>
            {t('settings.addVehicle')}
          </button>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'repeat(auto-fill, minmax(300px, 1fr))', gap: 16, padding: isMobile ? 12 : 20 }}>
          {vehicles.map((v) => (
            <div key={v.id} style={{
              border: '1px solid #e2e8f0', borderRadius: 12, overflow: 'hidden',
              background: '#fff', transition: 'box-shadow 150ms',
            }}>
              {/* Photo area */}
              <div
                onClick={() => handlePhotoClick(v.id)}
                style={{
                  height: 160, background: '#f1f5f9', cursor: 'pointer',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  position: 'relative', overflow: 'hidden',
                }}
              >
                {v.photo_url ? (
                  <img src={`${API_BASE}${v.photo_url}`} alt={v.name}
                    style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                ) : (
                  <div style={{ textAlign: 'center', color: '#94a3b8' }}>
                    <svg width="40" height="40" fill="none" stroke="currentColor" strokeWidth="1.5" viewBox="0 0 24 24" style={{ opacity: 0.4 }}>
                      <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z" />
                      <circle cx="12" cy="13" r="4" />
                    </svg>
                    <div style={{ fontSize: '0.75rem', marginTop: 6 }}>{t('settings.uploadPhoto')}</div>
                  </div>
                )}
                {v.photo_url && (
                  <div style={{
                    position: 'absolute', bottom: 8, right: 8,
                    background: 'rgba(0,0,0,0.6)', color: '#fff', borderRadius: 6,
                    padding: '4px 8px', fontSize: '0.7rem',
                  }}>{t('settings.changePhoto')}</div>
                )}
              </div>

              {/* Info */}
              <div style={{ padding: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div>
                    <div style={{ fontWeight: 700, fontSize: '1rem', color: '#0f172a' }}>{v.name}</div>
                    <div style={{ fontSize: '0.82rem', color: '#64748b', marginTop: 2 }}>
                      {v.license_plate}
                    </div>
                  </div>
                  <StatusDot status={v.status} />
                </div>
                <div style={{ fontSize: '0.82rem', color: '#475569', marginTop: 10 }}>
                  {t('settings.driverLabel')} <strong>{v.driver_name}</strong>
                </div>
                <div style={{ display: 'flex', gap: 8, marginTop: 14 }}>
                  <button onClick={() => openEdit(v)} style={{
                    flex: 1, padding: '8px', background: '#f1f5f9', color: '#475569',
                    border: '1px solid #e2e8f0', borderRadius: 8, cursor: 'pointer',
                    fontWeight: 600, fontSize: '0.78rem', fontFamily: 'inherit',
                  }}>{t('common.edit')}</button>
                  <button onClick={() => handleDelete(v.id)} style={{
                    padding: '8px 12px', background: '#fef2f2', color: '#dc2626',
                    border: '1px solid #fecaca', borderRadius: 8, cursor: 'pointer',
                    fontWeight: 600, fontSize: '0.78rem', fontFamily: 'inherit',
                  }}>{t('common.delete')}</button>
                </div>
              </div>
            </div>
          ))}

          {vehicles.length === 0 && (
            <div style={{ gridColumn: '1 / -1', padding: 48, textAlign: 'center', color: '#94a3b8' }}>
              {t('settings.noVehicles')}
            </div>
          )}
        </div>
      </div>
    </>
  );
}

function StatusDot({ status }: { status: string }) {
  const { t } = useI18nStore();
  const c = vehicleStatusColor(status);
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 5,
      padding: '3px 10px', borderRadius: 9999,
      fontSize: '0.7rem', fontWeight: 600, color: c, background: `${c}14`,
    }}>
      <span style={{ width: 6, height: 6, borderRadius: '50%', background: c }} />
      {vehicleStatusLabel(status, t)}
    </span>
  );
}

const labelStyle: React.CSSProperties = {
  display: 'block', fontSize: '0.8rem', fontWeight: 600, color: '#475569', marginBottom: 5,
};

// --- User Management ---

function UserManagement() {
  const { t } = useI18nStore();
  const isMobile = useIsMobile();
  const [users, setUsers] = useState<User[]>([]);

  const fetchUsers = async () => {
    const data = await listUsers();
    setUsers(data || []);
  };

  useEffect(() => { fetchUsers(); }, []);

  const handleRoleChange = async (userId: string, role: string) => {
    await updateUserRole(userId, role);
    fetchUsers();
  };

  const handlePriorityChange = async (userId: string, priority: string) => {
    const num = parseInt(priority);
    if (!isNaN(num)) {
      await updateUserPriority(userId, num);
      fetchUsers();
    }
  };

  const roleOptions: { value: string; key: TranslationKey }[] = [
    { value: 'admin', key: 'role.admin' },
    { value: 'dispatcher', key: 'role.dispatcher' },
    { value: 'viewer', key: 'role.viewer' },
    { value: 'driver', key: 'role.driver' },
  ];

  const tableHeaders = [
    t('settings.tableHeaderName'),
    t('settings.tableHeaderEmployeeId'),
    t('settings.tableHeaderRole'),
    t('settings.tableHeaderPriority'),
  ];

  return (
    <div style={{
      background: '#fff', borderRadius: 12,
      border: '1px solid #e2e8f0', overflow: 'hidden',
      boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
    }}>
      <div style={{
        padding: '14px 20px', borderBottom: '1px solid #e2e8f0',
        background: '#f8fafc',
      }}>
        <h2 style={{ margin: 0, fontSize: '0.9rem', fontWeight: 600, color: '#475569' }}>
          {t('settings.userManagement')}
          <span style={{ marginLeft: 8, fontSize: '0.8rem', fontWeight: 400, color: '#94a3b8' }}>
            {t('settings.userCount', { count: users.length })}
          </span>
        </h2>
      </div>
      <ResponsiveTable>
      <table style={{ width: '100%', borderCollapse: 'collapse', minWidth: isMobile ? 600 : undefined }}>
        <thead>
          <tr>
            {tableHeaders.map((h) => (
              <th key={h} style={{ padding: '12px 16px', textAlign: 'left', fontSize: '0.75rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.04em', color: '#64748b', background: '#f8fafc', borderBottom: '1px solid #e2e8f0' }}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {users.map((u) => {
            const badge = roleBadge[u.role] || { bg: '#f1f5f9', color: '#475569' };
            return (
              <tr key={u.id} style={{ borderBottom: '1px solid #f1f5f9' }}>
                <td style={{ padding: '12px 16px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <div style={{
                      width: 32, height: 32, borderRadius: 8,
                      background: badge.bg, color: badge.color,
                      display: 'flex', alignItems: 'center', justifyContent: 'center',
                      fontWeight: 700, fontSize: '0.82rem',
                    }}>{u.name.charAt(0)}</div>
                    <span style={{ fontWeight: 500 }}>{u.name}</span>
                  </div>
                </td>
                <td style={{ padding: '12px 16px', fontSize: '0.82rem', color: '#64748b', fontFamily: 'monospace' }}>{u.employee_id}</td>
                <td style={{ padding: '12px 16px' }}>
                  <select value={u.role} onChange={e => handleRoleChange(u.id, e.target.value)}
                    style={{ width: 'auto', padding: '5px 10px', borderRadius: 6, fontSize: '0.82rem' }}>
                    {roleOptions.map(({ value, key }) => (
                      <option key={value} value={value}>{t(key)}</option>
                    ))}
                  </select>
                </td>
                <td style={{ padding: '12px 16px' }}>
                  <input type="number" value={u.priority_level} onChange={e => handlePriorityChange(u.id, e.target.value)}
                    style={{ width: 70, padding: '5px 10px', borderRadius: 6, fontSize: '0.82rem', textAlign: 'center' }} />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
      </ResponsiveTable>
    </div>
  );
}
