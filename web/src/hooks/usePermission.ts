import { useAuthStore } from '../stores/authStore';
import type { Role } from '../types/api';

export function usePermission(...roles: Role[]): boolean {
  const user = useAuthStore((s) => s.user);
  return user ? roles.includes(user.role) : false;
}
