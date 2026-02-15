#!/usr/bin/env node
/**
 * Wrapper script to run Artillery with dynamic configuration
 * Reads VUS and DURATION from environment and generates config
 */

const fs = require('fs');
const { execSync } = require('child_process');
const path = require('path');

// Read environment variables
const TARGET_URL = process.env.TARGET_URL;
const VUS = parseInt(process.env.VUS || '10');
const DURATION = parseInt(process.env.DURATION || '60');

if (!TARGET_URL) {
  console.error('ERROR: TARGET_URL environment variable is required');
  process.exit(1);
}

console.log(`Configuring load test:`);
console.log(`  Target URL: ${TARGET_URL}`);
console.log(`  Virtual Users: ${VUS}`);
console.log(`  Duration: ${DURATION}s`);
console.log('');

// Read the base config
const configPath = path.join(__dirname, 'config.yml');
let config = fs.readFileSync(configPath, 'utf8');

// Replace placeholder values with actual numbers
config = config.replace(/duration: 60/, `duration: ${DURATION}`);
config = config.replace(/arrivalRate: 10/, `arrivalRate: ${VUS}`);

// Write temporary config
const tmpConfig = path.join(__dirname, '.config.tmp.yml');
fs.writeFileSync(tmpConfig, config);

try {
  // Run Artillery
  const args = process.argv.slice(2);
  const cmd = `npx artillery run ${tmpConfig} ${args.join(' ')}`;
  console.log(`Running: ${cmd}\n`);
  execSync(cmd, { stdio: 'inherit' });
} finally {
  // Cleanup
  if (fs.existsSync(tmpConfig)) {
    fs.unlinkSync(tmpConfig);
  }
}
