import { readFile } from 'node:fs/promises';

export interface SetIR { name: string }
export interface ModelIR {
  schemaVersion: string;
  name: string;
  sets: SetIR[];
  parameters: unknown[];
  variables: unknown[];
  constraints: unknown[];
  objectives: unknown[];
}

export async function loadIr(path: string): Promise<ModelIR> {
  const raw = await readFile(path, 'utf8');
  return JSON.parse(raw) as ModelIR;
}
