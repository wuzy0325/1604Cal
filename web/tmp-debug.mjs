import { chromium } from 'playwright';
(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));
  page.on('console', msg => console.log('CONSOLE:', msg.type(), msg.text()));
  await page.goto('http://localhost:4173/', { waitUntil: 'load', timeout: 15000 });
  await new Promise(r => setTimeout(r, 2000));
  await page.locator('a[href="#/device-management"]').click();
  await new Promise(r => setTimeout(r, 5000));
  const html = await page.content();
  console.log(html.slice(0, 5000));
  await browser.close();
})();
