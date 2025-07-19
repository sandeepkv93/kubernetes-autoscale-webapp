const puppeteer = require('puppeteer');

async function testFrontend() {
  let browser;
  try {
    browser = await puppeteer.launch({ headless: true });
    const page = await browser.newPage();
    
    // Listen for console messages and errors
    page.on('console', msg => console.log('PAGE LOG:', msg.text()));
    page.on('pageerror', err => console.log('PAGE ERROR:', err.message));
    
    // Navigate to frontend
    await page.goto('http://localhost:3001', { waitUntil: 'networkidle0' });
    
    // Check if page loaded
    const title = await page.title();
    console.log('Page title:', title);
    
    // Check if users are loaded
    await page.waitForSelector('ul', { timeout: 5000 });
    const users = await page.$$eval('ul li', elements => elements.map(el => el.textContent));
    console.log('Existing users:', users.length);
    
    // Fill out the form
    await page.type('input[placeholder="Name"]', 'Test Frontend User');
    await page.type('input[placeholder="Email"]', 'test-frontend@example.com');
    
    // Submit form
    await page.click('button[type="submit"]');
    
    // Wait for the new user to appear
    await new Promise(resolve => setTimeout(resolve, 3000));
    
    // Check if user was added
    const newUsers = await page.$$eval('ul li', elements => elements.map(el => el.textContent));
    console.log('Users after creation:', newUsers.length);
    
    if (newUsers.length > users.length) {
      console.log('✅ SUCCESS: User was created successfully!');
      console.log('New user:', newUsers[0]);
    } else {
      console.log('❌ FAILURE: User was not created');
    }
    
  } catch (error) {
    console.error('❌ Test failed:', error.message);
  } finally {
    if (browser) {
      await browser.close();
    }
  }
}

testFrontend();