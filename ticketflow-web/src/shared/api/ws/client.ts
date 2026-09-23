import type { AvailabilityFrame } from '@/shared/api/types';

type Listener = (frame: AvailabilityFrame) => void;
type ConnectionState = 'connecting' | 'open' | 'closed';

const MAX_BACKOFF_MS = 30_000;

export class AvailabilitySocket {
  private socket: WebSocket | null = null;
  private reconnectAttempt = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private failedAttempts = 0;
  private pollTimer: ReturnType<typeof setInterval> | null = null;
  private closedByUser = false;
  private readonly listeners = new Set<Listener>();

  constructor(
    private readonly eventId: string,
    private readonly onState: (state: ConnectionState) => void,
    private readonly onFrame: Listener,
    private readonly startPolling: () => void,
    private readonly stopPolling: () => void,
  ) {}

  connect(): void {
    this.closedByUser = false;
    this.open();
    window.addEventListener('online', this.handleOnline);
    document.addEventListener('visibilitychange', this.handleVisibility);
  }

  disconnect(): void {
    this.closedByUser = true;
    this.teardown();
    window.removeEventListener('online', this.handleOnline);
    document.removeEventListener('visibilitychange', this.handleVisibility);
  }

  private teardown(): void {
    this.socket?.close();
    this.socket = null;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.reconnectTimer = null;
    this.failedAttempts = 0;
    this.stopPolling();
  }

  private open(): void {
    if (this.socket) return;
    const origin = process.env.NEXT_PUBLIC_WS_ORIGIN ?? 'ws://localhost:8080';
    this.onState('connecting');
    const socket = new WebSocket(`${origin}/ws/events/${this.eventId}`);

    socket.onopen = () => {
      this.reconnectAttempt = 0;
      this.failedAttempts = 0;
      this.stopPolling();
      this.onState('open');
    };
    socket.onmessage = (e) => {
      try {
        this.onFrame(JSON.parse(e.data as string) as AvailabilityFrame);
      } catch {
        return;
      }
    };
    socket.onclose = () => {
      this.socket = null;
      this.onState('closed');
      if (this.closedByUser) return;
      this.scheduleReconnect();
    };
    socket.onerror = () => socket.close();

    this.socket = socket;
  }

  private scheduleReconnect(): void {
    const delay = Math.min(1000 * 2 ** this.reconnectAttempt, MAX_BACKOFF_MS);
    const jitter = delay * (0.7 + Math.random() * 0.3);
    this.reconnectAttempt += 1;
    this.reconnectTimer = setTimeout(() => this.open(), jitter);
  }

  private handleOnline = (): void => {
    if (!this.socket && !this.closedByUser) {
      this.reconnectAttempt = 0;
      this.open();
    }
  };

  private handleVisibility = (): void => {
    if (document.visibilityState === 'visible' && !this.socket && !this.closedByUser) {
      this.reconnectAttempt = 0;
      this.open();
    }
  };

  notifyPollingFailure(): void {
    this.failedAttempts += 1;
    if (this.failedAttempts >= 3 && !this.pollTimer) this.startPolling();
  }

  notifyPollingSuccess(): void {
    this.failedAttempts = 0;
    this.stopPolling();
  }

  get state(): ConnectionState {
    return this.socket?.readyState === WebSocket.OPEN ? 'open' : 'closed';
  }
}
