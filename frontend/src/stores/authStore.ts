import { create } from 'zustand';
import type { User } from '@/types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isNewUser: boolean;
}

interface AuthActions {
  setUser: (user: User | null) => void;
  logout: () => void;
  updateEmail: (email: string) => void;
}

export const useAuthStore = create<AuthState & AuthActions>((set, get) => ({
  user: null,
  isAuthenticated: false,
  isLoading: false,
  isNewUser: false,

  setUser: (user) =>
    set({
      user,
      isAuthenticated: user !== null,
    }),

  logout: () =>
    set({
      user: null,
      isAuthenticated: false,
      isNewUser: false,
    }),

  updateEmail: (email) => {
    const { user } = get();
    if (!user) return;
    set({
      user: { ...user, email },
    });
  },
}));
