/**
 * Load Test Helpers
 * Utilities for generating realistic test data
 */

const { faker } = require('@faker-js/faker');

/**
 * Generate a unique email for a virtual user
 * Format: loadtest+{timestamp}+{vuId}@example.com
 */
function generateEmail(vuId) {
  const timestamp = Date.now();
  return `loadtest+${timestamp}+${vuId}@example.com`;
}

/**
 * Generate a realistic name
 */
function generateName() {
  return faker.person.fullName();
}

/**
 * Generate a valid UK mobile phone number in E.164 format
 * Format: +447XXXXXXXXX (UK mobile numbers)
 * Uses fixed test number to avoid validation issues
 */
function generateUKMobile() {
  // Use fixed UK test number range (07700 9xxxxx)
  // This is a reserved range for testing and will pass mobile validation
  return '+447700900123';
}

/**
 * Randomly select 'a' or 'b' for question answers
 */
function selectRandomChoice() {
  return Math.random() < 0.5 ? 'a' : 'b';
}

/**
 * Generate a complete set of test data for a user journey
 */
function generateUserData(vuId) {
  return {
    name: generateName(),
    email: generateEmail(vuId),
    phone: generateUKMobile(),
    vuId: vuId
  };
}

/**
 * Add a random delay to simulate realistic user behavior
 * Returns a promise that resolves after the delay
 */
async function humanDelay(minMs = 500, maxMs = 2000) {
  const delay = Math.floor(Math.random() * (maxMs - minMs + 1)) + minMs;
  return new Promise(resolve => setTimeout(resolve, delay));
}

/**
 * Extract slug from target URL
 * Example: https://pick6.fly.dev/tk03 -> tk03
 */
function extractSlugFromUrl(url) {
  try {
    const urlObj = new URL(url);
    const pathParts = urlObj.pathname.split('/').filter(part => part.length > 0);
    if (pathParts.length > 0) {
      return pathParts[0];
    }
    throw new Error('No slug found in URL path');
  } catch (error) {
    throw new Error(`Failed to extract slug from URL: ${url}. Error: ${error.message}`);
  }
}

/**
 * Get base URL without slug
 * Example: https://pick6.fly.dev/tk03 -> https://pick6.fly.dev
 */
function getBaseUrl(url) {
  try {
    const urlObj = new URL(url);
    return `${urlObj.protocol}//${urlObj.host}`;
  } catch (error) {
    throw new Error(`Invalid URL: ${url}`);
  }
}

module.exports = {
  generateEmail,
  generateName,
  generateUKMobile,
  selectRandomChoice,
  generateUserData,
  humanDelay,
  extractSlugFromUrl,
  getBaseUrl
};
