import { create } from 'zustand';
import type { TwitterConnectionInfo, TwitterConnectionStatus } from '@/types/twitter';

interface TwitterState {
  connectionInfo: TwitterConnectionInfo;
  isTestingConnection: boolean;
  isConnectionLoaded: boolean;
}

interface TwitterActions {
  setConnectionInfo: (info: TwitterConnectionInfo) => void;
  setIsTestingConnection: (v: boolean) => void;
  setIsConnectionLoaded: (v: boolean) => void;
  updateConnectionStatus: (status: TwitterConnectionStatus, errorMessage?: string) => void;
  resetConnection: () => void;
}

const DISCONNECTED_INFO: TwitterConnectionInfo = {
  status: 'disconnected',
  connectedAt: null,
  lastTestedAt: null,
  errorMessage: null,
};

export const useTwitterStore = create<TwitterState & TwitterActions>((set) => ({
  connectionInfo: { ...DISCONNECTED_INFO },
  isTestingConnection: false,
  isConnectionLoaded: false,

  setConnectionInfo: (info) => set({ connectionInfo: info }),
  setIsTestingConnection: (v) => set({ isTestingConnection: v }),
  setIsConnectionLoaded: (v) => set({ isConnectionLoaded: v }),

  updateConnectionStatus: (status, errorMessage) =>
    set((state) => ({
      connectionInfo: {
        ...state.connectionInfo,
        status,
        errorMessage: errorMessage ?? null,
      },
    })),

  resetConnection: () =>
    set({ connectionInfo: { ...DISCONNECTED_INFO } }),
}));
