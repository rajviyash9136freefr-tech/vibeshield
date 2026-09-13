#!/usr/bin/env node
// vibeshield npm wrapper — entry point.
//
// Deliberately has NO install-time script: the Go binary is fetched on first
// run and cached (~/.cache/vibeshield or %LOCALAPPDATA%\vibeshield), so
// `npm i -g vibeshield` never touches the network and works on locked-down
// machines. Network use at runtime stays opt-in exactly like the scanner's
// own promise: this download is the only request the tool ever makes, it
// goes to GitHub Releases over HTTPS, and the sha256sums.txt from the same
// release is verified before the binary is ever executed.
import { main } from '../lib/run.js';

main(process.argv.slice(2));
