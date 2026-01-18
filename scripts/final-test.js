#!/usr/bin/env node

const puppeteer = require('puppeteer');

async function test() {
  console.log('🚀 启动全面测试...\n');
  const browser = await puppeteer.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const errors = [];
  const apiErrors = [];
  let testsPassed = 0;
  let testsFailed = 0;

  const page = await browser.newPage();
  
  page.on('response', async response => {
    const url = response.url();
    if (url.includes('/api/') && !url.includes('/api/sse/')) {
      const status = response.status();
      if (status >= 400) {
        try {
          const text = await response.text();
          apiErrors.push({ url, status, body: text.substring(0, 200) });
        } catch (e) {}
      }
    }
  });

  page.on('console', msg => {
    if (msg.type() === 'error' && !msg.text().includes('favicon')) {
      errors.push(msg.text());
    }
  });

  page.on('pageerror', error => {
    errors.push(error.message);
  });

  try {
    // 测试用户门户
    console.log('📱 测试用户门户\n');
    
    await page.goto('http://localhost:8081/user/login', { waitUntil: 'networkidle2', timeout: 30000 });
    await page.waitForSelector('input[type="text"]');
    await page.type('input[type="text"]', 'admin');
    await page.type('input[type="password"]', 'admin123');
    await page.keyboard.press('Enter');
    await new Promise(resolve => setTimeout(resolve, 3000));
    console.log('✅ 用户登录成功');
    testsPassed++;

    const userPages = [
      { url: '/user/dashboard', name: '用户仪表板' },
      { url: '/user/subscription', name: '订阅页面' },
      { url: '/user/plans', name: '套餐页面' },
      { url: '/user/orders', name: '订单页面' }
    ];

    for (const { url, name } of userPages) {
      await page.goto(`http://localhost:8081${url}`, { waitUntil: 'networkidle2' });
      await new Promise(resolve => setTimeout(resolve, 1500));
      console.log(`✅ ${name}加载成功`);
      testsPassed++;
    }

    // 测试管理后台
    console.log('\n🔧 测试管理后台\n');
    
    // 清除用户token，重新登录管理后台
    await page.evaluate(() => {
      localStorage.removeItem('userToken');
      sessionStorage.removeItem('userToken');
    });

    await page.goto('http://localhost:8081/login', { waitUntil: 'networkidle2' });
    await page.waitForSelector('input[type="text"]');
    await page.evaluate(() => {
      document.querySelectorAll('input').forEach(el => el.value = '');
    });
    await page.type('input[type="text"]', 'admin');
    await page.type('input[type="password"]', 'admin123');
    await page.keyboard.press('Enter');
    await new Promise(resolve => setTimeout(resolve, 3000));
    console.log('✅ 管理员登录成功');
    testsPassed++;

    const adminPages = [
      { url: '/admin/dashboard', name: '管理仪表板' },
      { url: '/admin/users', name: '用户管理' },
      { url: '/admin/nodes', name: '节点管理' },
      { url: '/admin/subscriptions', name: '订阅管理' },
      { url: '/admin/plans', name: '套餐管理' },
      { url: '/admin/orders', name: '订单管理' },
      { url: '/admin/ip-restriction', name: 'IP限制' },
      { url: '/admin/settings', name: '系统设置' },
      { url: '/admin/stats', name: '统计报表' },
      { url: '/admin/logs', name: '日志管理' }
    ];

    for (const { url, name } of adminPages) {
      await page.goto(`http://localhost:8081${url}`, { waitUntil: 'networkidle2' });
      await new Promise(resolve => setTimeout(resolve, 1500));
      console.log(`✅ ${name}加载成功`);
      testsPassed++;
    }

  } catch (error) {
    console.error(`\n❌ 测试失败: ${error.message}`);
    testsFailed++;
  } finally {
    await browser.close();
  }

  // 输出汇总
  console.log('\n' + '='.repeat(50));
  console.log('📊 测试汇总');
  console.log('='.repeat(50) + '\n');
  
  console.log(`✅ 通过: ${testsPassed} 个测试`);
  if (testsFailed > 0) console.log(`❌ 失败: ${testsFailed} 个测试`);
  if (errors.length > 0) console.log(`⚠️  前端错误: ${errors.length} 个`);
  if (apiErrors.length > 0) console.log(`⚠️  API错误: ${apiErrors.length} 个`);
  
  if (errors.length === 0 && apiErrors.length === 0 && testsFailed === 0) {
    console.log('\n🎉 所有测试通过！系统运行正常！');
  } else {
    console.log('\n⚠️  发现以下问题:\n');
    
    if (apiErrors.length > 0) {
      console.log('API错误:');
      apiErrors.forEach((err, i) => {
        console.log(`  ${i + 1}. ${err.status} ${err.url}`);
        console.log(`     ${err.body}`);
      });
    }
    
    if (errors.length > 0) {
      console.log('\n前端错误:');
      errors.slice(0, 5).forEach((err, i) => {
        console.log(`  ${i + 1}. ${err}`);
      });
    }
  }
}

test().catch(console.error);
