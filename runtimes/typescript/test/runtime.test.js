import test from 'node:test';
import assert from 'node:assert/strict';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { loadIr } from '../src/index.ts';

test('loads shared IR', async () => {
  const here = path.dirname(fileURLToPath(import.meta.url));
  const model = await loadIr(path.resolve(here, '../../../testdata/transport.ir.json'));
  assert.equal(model.schemaVersion, 'miplang.ir/v1alpha2');
  assert.equal(model.name, 'anonymous');
  assert.deepEqual(model.sets.map((s) => s.name), ['TRIPS']);
});
