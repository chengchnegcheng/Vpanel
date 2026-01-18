#!/usr/bin/env node

const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({
    headless: false,
    args: ['--no-sandbox']
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1920, height: 1080 });

  try {
    console.log('登录...');
    await page.goto('http://localhost:8081/login', { waitUntil: 'networkidle2' });
    await new Promise(r => setTimeout(r, 2000));

    await page.waitForSelector('input[type="text"]');
    await page.type('input[type="text"]', 'admin');
    await page.type('input[type="password"]', 'admin123');
    await page.keyboard.press('Enter');
    await new Promise(r => setTimeout(r, 3000));

    console.log('浅色模式截图...');
    await page.screenshot({ path: 'light.png', fullPage: true });

    console.log('切换深色模式...');
    const btn = await page.$('button[title="切换主题"]');
    if (btn) {
      await btn.click();
      await new Promise(r => setTimeout(r, 2000));
    }

    console.log('深色模式截图...');
    await page.screenshot({ path: 'dark.png', fullPage: true });

    const isDark = await page.evaluate(() => document.documentElement.classList.contains('dark'));
    console.log('深色模式:', isDark ? '启用' : '未启用');

    await new Promise(r => setTimeout(r, 5000));

  } catch (e) {
    console.error('错误:', e.message);
  } finally {
    await browser.close();
  }
})();
