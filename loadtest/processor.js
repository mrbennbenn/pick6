/**
 * Artillery Playwright Processor
 * Simulates complete user journey through Pick6 application
 */

const { chromium } = require('playwright');
const helpers = require('./helpers');

// Track active contexts for cleanup
const activeContexts = new Map();

/**
 * Main user journey function called by Artillery
 * Each virtual user executes this complete flow
 */
async function userJourney(page, vuContext, events, test) {
  const startTime = Date.now();
  
  try {
    // Get target URL from environment
    const targetUrl = process.env.TARGET_URL;
    if (!targetUrl) {
      throw new Error('TARGET_URL environment variable is required');
    }

    // Extract base URL and slug
    const baseUrl = helpers.getBaseUrl(targetUrl);
    const slug = helpers.extractSlugFromUrl(targetUrl);
    
    // Generate unique test data for this virtual user
    const vuId = vuContext.vars.$uuid || Date.now();
    const userData = helpers.generateUserData(vuId);

    console.log(`[VU-${vuId}] Starting journey for slug: ${slug}`);

    // Step 1: Initial visit - triggers session creation
    console.log(`[VU-${vuId}] Step 1: Visiting home page`);
    const step1Start = Date.now();
    await page.goto(`${baseUrl}/${slug}/`, { waitUntil: 'networkidle' });
    events.emit('histogram', 'load.step1.initial_visit', Date.now() - step1Start);
    
    // Should redirect to question 1
    await page.waitForURL(`**/${slug}/question/1`, { timeout: 10000 });
    console.log(`[VU-${vuId}] Redirected to question 1`);

    // Step 2: Navigate through all 6 questions
    for (let questionNum = 1; questionNum <= 6; questionNum++) {
      console.log(`[VU-${vuId}] Step 2.${questionNum}: Answering question ${questionNum}`);
      
      // Wait for page to load
      const questionStepStart = Date.now();
      await page.waitForSelector('form', { timeout: 10000 });
      
      // Simulate user reading the question
      await helpers.humanDelay(300, 800);
      
      // Randomly select an answer (a or b)
      const choice = helpers.selectRandomChoice();
      console.log(`[VU-${vuId}] Selecting choice: ${choice}`);
      
      // Click the radio button for the choice
      await page.click(`input[type="radio"][value="${choice}"]`);
      
      // Small delay before submitting
      await helpers.humanDelay(200, 500);
      
      // Submit the form
      const submitStart = Date.now();
      await Promise.all([
        page.click('button[type="submit"]'),
        page.waitForNavigation({ timeout: 10000 })
      ]);
      
      const submitTime = Date.now() - submitStart;
      events.emit('histogram', `load.step2.question${questionNum}_submit`, submitTime);
      events.emit('histogram', 'load.step2.question_submit_all', submitTime);
      
      const totalQuestionTime = Date.now() - questionStepStart;
      console.log(`[VU-${vuId}] Question ${questionNum} completed in ${totalQuestionTime}ms`);
    }

    // Step 3: Fill out info form
    console.log(`[VU-${vuId}] Step 3: Filling info form`);
    const step3Start = Date.now();
    
    // Should now be on submit-info page
    await page.waitForURL(`**/${slug}/submit-info`, { timeout: 10000 });
    await page.waitForSelector('form', { timeout: 10000 });
    
    // Fill in the form fields
    await page.fill('input[name="name"]', userData.name);
    await page.fill('input[name="email"]', userData.email);
    await page.fill('input[name="phone"]', userData.phone);
    
    console.log(`[VU-${vuId}] Submitting info: ${userData.email}`);
    
    // Small delay before submitting
    await helpers.humanDelay(200, 500);
    
    // Submit info form
    const infoSubmitStart = Date.now();
    await Promise.all([
      page.click('button[type="submit"]'),
      page.waitForNavigation({ timeout: 10000 })
    ]);
    
    const infoSubmitTime = Date.now() - infoSubmitStart;
    events.emit('histogram', 'load.step3.info_submit', infoSubmitTime);
    
    const totalStep3Time = Date.now() - step3Start;
    console.log(`[VU-${vuId}] Info form completed in ${totalStep3Time}ms`);

    // Step 4: Verify end page
    console.log(`[VU-${vuId}] Step 4: Verifying end page`);
    const step4Start = Date.now();
    
    await page.waitForURL(`**/${slug}/end`, { timeout: 10000 });
    
    // Wait for the page to fully load
    await page.waitForLoadState('networkidle');
    
    events.emit('histogram', 'load.step4.end_page', Date.now() - step4Start);

    // Calculate total journey time
    const totalTime = Date.now() - startTime;
    events.emit('histogram', 'load.journey.total_time', totalTime);
    events.emit('counter', 'load.journey.completed', 1);
    
    console.log(`[VU-${vuId}] ✓ Journey completed successfully in ${totalTime}ms`);

  } catch (error) {
    const totalTime = Date.now() - startTime;
    console.error(`[VU-${vuContext.vars.$uuid}] ✗ Journey failed after ${totalTime}ms:`, error.message);
    events.emit('counter', 'load.journey.failed', 1);
    events.emit('counter', `load.errors.${error.name || 'UnknownError'}`, 1);
    throw error;
  }
}

/**
 * Initialize browser for the test scenario
 * Called once per virtual user
 */
async function createBrowserContext(context, events, done) {
  try {
    // Launch browser with performance optimizations
    const browser = await chromium.launch({
      headless: true,
      args: [
        '--disable-dev-shm-usage',
        '--no-sandbox',
        '--disable-setuid-sandbox',
        '--disable-gpu'
      ]
    });

    // Create a new browser context (isolated session)
    const browserContext = await browser.newContext({
      viewport: { width: 1280, height: 720 },
      userAgent: 'LoadTest-Artillery/Playwright',
      // Ensure cookies are enabled
      acceptDownloads: false,
      javaScriptEnabled: true
    });

    // Create a new page
    const page = await browserContext.newPage();

    // Store for cleanup
    const contextId = context.vars.$uuid || Date.now();
    activeContexts.set(contextId, { browser, browserContext, page });

    // Make page available to the scenario
    context.vars.page = page;

    console.log(`[Context-${contextId}] Browser context created`);
    done();
  } catch (error) {
    console.error('Failed to create browser context:', error);
    done(error);
  }
}

/**
 * Cleanup browser context after scenario completes
 */
async function closeBrowserContext(context, events, done) {
  try {
    const contextId = context.vars.$uuid || Date.now();
    const resources = activeContexts.get(contextId);

    if (resources) {
      const { page, browserContext, browser } = resources;
      
      if (page) await page.close();
      if (browserContext) await browserContext.close();
      if (browser) await browser.close();
      
      activeContexts.delete(contextId);
      console.log(`[Context-${contextId}] Browser context closed`);
    }

    done();
  } catch (error) {
    console.error('Error closing browser context:', error);
    done(error);
  }
}

/**
 * Override Artillery config at runtime with environment variables
 * This allows numeric values which don't work in YAML templates
 * 
 * Uses a two-phase approach:
 * - Phase 1: Spawn all VUs immediately (1 second)
 * - Phase 2: Let them run concurrently for the remaining duration
 */
function config(scriptConfig) {
  const duration = parseInt(process.env.DURATION || '60');
  const vus = parseInt(process.env.VUS || '10');
  
  // Phase 1: Initialize all users immediately
  // Phase 2: Run for remaining duration with no new arrivals
  scriptConfig.phases = [
    {
      duration: 1,
      arrivalCount: vus,
      name: 'Initialize Users'
    },
    {
      duration: Math.max(duration - 1, 1),
      arrivalRate: 0,
      name: 'Sustained Load'
    }
  ];
  
  console.log(`[Config] Using DURATION=${duration}, VUS=${vus} (concurrent)`);
  console.log(`[Config] Phase 1: Spawn ${vus} users in 1s`);
  console.log(`[Config] Phase 2: Run for ${Math.max(duration - 1, 1)}s with 0 new arrivals`);
  
  return scriptConfig;
}

module.exports = {
  config,
  userJourney,
  createBrowserContext,
  closeBrowserContext
};
