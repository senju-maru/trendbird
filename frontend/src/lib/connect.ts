import type { GenService, GenServiceMethods } from '@bufbuild/protobuf/codegenv2';
import type { Client } from '@connectrpc/connect';
import { createClient } from '@connectrpc/connect';
import { getTransport } from './transport';

export async function getClient<S extends GenService<GenServiceMethods>>(
  service: S,
): Promise<Client<S>> {
  const transport = await getTransport();
  return createClient(service, transport);
}
