// Exercise one real WASM program through booba's JS bridge in a fresh process.
// This checks startup/input/resize/exit, not browser canvas rendering.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';
import { setTimeout as delay } from 'node:timers/promises';
const demo = process.argv[2];
assert.ok(['catalog', 'dashboard', 'hud'].includes(demo));
vm.runInThisContext(await readFile('web/wasm_exec.js', 'utf8'));
const go = new Go();
let exitCode;
go.exit = code => { exitCode = code; };
const { instance } = await WebAssembly.instantiate(await readFile(`web/demos/${demo}.wasm`), go.importObject);
const timeout = setTimeout(() => { console.error(`${demo}: timed out`); process.exit(1); }, 15000);
const running = go.run(instance);
try {
  for (let n = 0; typeof globalThis.bubbletea_read !== 'function' && n < 100; n++) await delay(20);
  assert.equal(typeof globalThis.bubbletea_read, 'function', 'booba bridge installed');
  bubbletea_resize(92, 28);
  let output = '';
  for (let n = 0; n < 100 && !output.includes('\x1b]7501;?'); n++) {
    await delay(25);
    output += bubbletea_read();
  }
  assert.ok(output.length > 0, 'initial terminal view rendered');
  assert.ok(output.includes('\x1b]7501;?'), 'application queries action support');
  bubbletea_write('\x1b]7501;?\x1b\\');
  let report;
  for (let n = 0; n < 100 && !report; n++) {
    await delay(25);
    output += bubbletea_read();
    report = output.match(/\x1b\]7501;state=idle:[^\x1b]*:msg=([A-Za-z0-9+/=]+)\x1b\\/);
  }
  assert.ok(report, 'application publishes OSC action snapshot');
  const payload = Buffer.from(report[1], 'base64').toString('utf8');
  const snapshot = JSON.parse(payload.split(' | booba-actions/v1 ')[1]);
  const quit = snapshot.actions.find(action => action.label === 'Quit');
  assert.ok(quit?.enabled, 'quit action enabled');

  const checkAction = label => assert.ok(
    snapshot.actions.some(action => action.label === label && action.enabled),
    `${label} action`,
  );

  async function expectSummary(expected) {
    let received = '';
    for (let n = 0; n < 100; n++) {
      await delay(20);
      received += bubbletea_read();
      for (const match of received.matchAll(/\x1b\]7501;state=idle:[^\x1b]*:msg=([A-Za-z0-9+/=]+)\x1b\\/g)) {
        const message = Buffer.from(match[1], 'base64').toString('utf8');
        if (message.split(' | booba-actions/v1 ')[0] === expected) return;
      }
    }
    assert.fail(`Missing summary state: ${expected}; received ${JSON.stringify(received)}`);
  }

  async function send(bytes, summary) {
    bubbletea_read();
    bubbletea_write(bytes);
    await delay(25);
    bubbletea_write('\x1b]7501;?\x1b\\');
    await expectSummary(summary);
  }

  bubbletea_resize(88, 26);
  if (demo === 'catalog') {
    checkAction('Next tab');
    checkAction('Cycle theme');
    checkAction('Glyph / Raster');
    await send(']', 'Catalog: tab=Plots theme=0 border=0 backend=0');
    await send('t', 'Catalog: tab=Plots theme=1 border=0 backend=0');
    await send('b', 'Catalog: tab=Plots theme=1 border=1 backend=0');
    await send('g', 'Catalog: tab=Plots theme=1 border=1 backend=1');
    await send('g', 'Catalog: tab=Plots theme=1 border=1 backend=0');
    await send(']', 'Catalog: tab=Trees & Lists theme=1 border=1 backend=0');
  } else if (demo === 'dashboard') {
    checkAction('Pause / Resume');
    checkAction('Cycle theme');
    checkAction('Glyph / Raster');
    await send(' ', 'Dashboard: paused=true theme=0 backend=0');
    await send('t', 'Dashboard: paused=true theme=1 backend=0');
    await send('g', 'Dashboard: paused=true theme=1 backend=1');
    await send('g', 'Dashboard: paused=true theme=1 backend=0');
    await send(' ', 'Dashboard: paused=false theme=1 backend=0');
  } else if (demo === 'hud') {
    checkAction('Pause / Orbit');
    checkAction('Cycle mesh');
    checkAction('Zoom in');
    checkAction('Zoom out');
    checkAction('Reset camera');
    await send(' ', 'HUD: shape=0 rotate=false zoom=3.2 backend=0');
    await send('m', 'HUD: shape=1 rotate=false zoom=3.2 backend=0');
    await send('\x1b[<64;50;15M', 'HUD: shape=1 rotate=false zoom=3.0 backend=0');
    await send('r', 'HUD: shape=1 rotate=false zoom=3.2 backend=0');
    await send('g', 'HUD: shape=1 rotate=false zoom=3.2 backend=1');
    await send('g', 'HUD: shape=1 rotate=false zoom=3.2 backend=0');
  }

  bubbletea_write(quit.keys[quit.preferred].key);
  await running;
  assert.equal(exitCode, 0);
  assert.equal(globalThis.bubbletea_exited, true, 'exit published to adapter');
  console.log(`${demo}: startup, actions, resize, exact demo state transitions, and clean exit passed`);
} finally {
  clearTimeout(timeout);
}
process.exit(0);
