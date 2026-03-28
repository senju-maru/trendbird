export const TRACE_ID_HEADER = 'X-Trace-Id';

export function generateTraceId(): string {
  return crypto.randomUUID();
}
