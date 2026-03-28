export type TwitterConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error';

export interface TwitterConnectionInfo {
  status: TwitterConnectionStatus;
  connectedAt: string | null;
  lastTestedAt: string | null;
  errorMessage: string | null;
}
