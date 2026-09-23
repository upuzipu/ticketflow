export function computeSecondsLeft(
  expiresAt: string,
  serverTime: string | null,
  now: number,
): number {
  const offset = serverTime ? Date.parse(serverTime) - now : 0;
  const deadline = Date.parse(expiresAt);
  return Math.max(0, Math.floor((deadline - (now + offset)) / 1000));
}
