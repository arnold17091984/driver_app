import { vehicleStatusLabel, vehicleStatusColor } from '../../utils/formatters';
import { useI18nStore } from '../../stores/i18nStore';

interface Props {
  status: string;
}

export function VehicleStatusBadge({ status }: Props) {
  const { t } = useI18nStore();
  const color = vehicleStatusColor(status);
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 5,
        padding: '3px 10px',
        borderRadius: 9999,
        fontSize: '0.7rem',
        fontWeight: 600,
        color: color,
        background: `${color}14`,
        letterSpacing: '0.02em',
      }}
    >
      <span style={{
        width: 6, height: 6, borderRadius: '50%',
        background: color,
      }} />
      {vehicleStatusLabel(status, t)}
    </span>
  );
}
