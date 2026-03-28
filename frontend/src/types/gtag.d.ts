interface GtagUserProperties {}

interface Gtag {
  (command: 'set', targetId: string, config: Record<string, unknown>): void;
  (command: 'set', config: { user_id?: string }): void;
  (command: 'set', field: 'user_properties', properties: GtagUserProperties): void;
  (command: 'config', targetId: string, config?: Record<string, unknown>): void;
  (command: 'event', eventName: string, eventParams?: Record<string, unknown>): void;
}

interface Window {
  gtag?: Gtag;
}
