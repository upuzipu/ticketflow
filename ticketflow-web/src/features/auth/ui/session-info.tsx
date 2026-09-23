'use client';

import { useAuthStore } from '../model/auth-store';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';

export function SessionInfo() {
  const userId = useAuthStore((s) => s.userId);
  const userRole = useAuthStore((s) => s.userRole);
  const hasAccess = useAuthStore((s) => s.accessToken !== null);
  const hasRefresh = useAuthStore((s) => s.refreshToken !== null);

  return (
    <main className="mx-auto max-w-2xl space-y-3 p-6">
      <h1 className="text-2xl font-semibold">Protected content</h1>
      <p className="text-sm text-muted">The guard let you through: you are authenticated.</p>
      <Card className="space-y-2 p-4 text-sm">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="success">authenticated</Badge>
          {userRole && <Badge variant="accent">{userRole}</Badge>}
          <Badge variant={hasAccess ? 'success' : 'warning'}>
            {hasAccess ? 'access token in memory' : 'no access token'}
          </Badge>
          <Badge variant={hasRefresh ? 'success' : 'muted'}>
            {hasRefresh ? 'refresh token persisted' : 'no refresh token'}
          </Badge>
        </div>
        {userId && (
          <div className="text-muted">
            User ID: <code className="text-foreground">{userId}</code>
          </div>
        )}
      </Card>
    </main>
  );
}
