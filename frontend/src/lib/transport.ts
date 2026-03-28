import type { Transport, Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { ConnectError, Code } from '@connectrpc/connect';
import { TRACE_ID_HEADER, generateTraceId } from './trace';

let cachedTransport: Transport | null = null;

const traceIdInterceptor: Interceptor = (next) => async (req) => {
  req.header.set(TRACE_ID_HEADER, generateTraceId());
  return next(req);
};

const authRedirectInterceptor: Interceptor = (next) => async (req) => {
  try {
    return await next(req);
  } catch (error) {
    if (
      typeof window !== 'undefined' &&
      error instanceof ConnectError &&
      error.code === Code.Unauthenticated
    ) {
      window.location.href = '/api/clear-session';
      await new Promise(() => {});
    }
    throw error;
  }
};

export function getTransport(): Transport {
  if (cachedTransport) return cachedTransport;

  cachedTransport = createConnectTransport({
    baseUrl: process.env.NEXT_PUBLIC_API_URL ?? '',
    fetch: (input, init) =>
      globalThis.fetch(input, { ...init, credentials: 'include' }),
    interceptors: [traceIdInterceptor, authRedirectInterceptor],
  });

  return cachedTransport;
}
