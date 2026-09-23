import { RequireAuth } from '@/shared/ui/require-auth';
import { SessionInfo } from '@/features/auth/ui/session-info';

export const metadata = { title: 'Protected sandbox' };

export default function ProtectedSandboxPage() {
  return (
    <RequireAuth>
      <SessionInfo />
    </RequireAuth>
  );
}
