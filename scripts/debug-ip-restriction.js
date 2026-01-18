const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });
  
  const page = await browser.newPage();
  
  const allRequests = [];
  const allResponses = [];
  const consoleMessages = [];
  
  // 监听所有请求
  page.on('request', request => {
    if (request.url().includes('/api/')) {
      allRequests.push({
        method: request.method(),
        url: request.url()
      });
    }
  });
  
  // 监听所有响应
  page.on('response', async response => {
    const url = response.url();
    if (url.includes('/api/')) {
      const status = response.status();
      const record = {
        status,
        url,
        statusText: response.statusText()
      };
      
      try {
        const contentType = response.headers()['content-type'];
        if (contentType && contentType.includes('application/json')) {
          const body = await response.json();
          record.body = body;
        } else {
          const text = await response.text();
          record.body = text.substring(0, 200);
        }
      } catch (e) {
        record.body = '[无法读取]';
      }
      
      allResponses.push(record);
      
      if (status >= 400) {
        console.log(`\n❌ ${status} ${url}`);
        console.log(`   响应:`, JSON.stringify(record.body, null, 2));
      }
    }
  });
  
  // 监听控制台
  page.on('console', msg => {
    const text = msg.text();
    consoleMessages.push({ type: msg.type(), text });
    
    if (msg.type() === 'error') {
      console.log(`\n[浏览器错误] ${text}`);
    }
  });
  
  // 监听页面错误
  page.on('pageerror', error => {
    console.log(`\n[页面错误] ${error.message}`);
  });
  
  try {
    console.log('========================================');
    console.log('开始测试IP限制页面');
    console.log('========================================\n');
    
    // 登录
    console.log('1. 登录...');
    await page.goto('http://localhost:8081/login', { waitUntil: 'networkidle2' });
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const inputs = await page.$$('input');
    if (inputs.length >= 2) {
      await inputs[0].type('admin');
      await inputs[1].type('admin123');
    }
    
    const buttons = await page.$$('button');
    for (const button of buttons) {
      const text = await page.evaluate(el => el.textContent, button);
      if (text.includes('登录')) {
        await button.click();
        break;
      }
    }
    
    await page.waitForNavigation({ waitUntil: 'networkidle2', timeout: 10000 }).catch(() => {});
    await new Promise(resolve => setTimeout(resolve, 2000));
    console.log('   ✓ 登录成功\n');
    
    // 清空之前的记录
    allRequests.length = 0;
    allResponses.length = 0;
    consoleMessages.length = 0;
    
    // 访问IP限制页面
    console.log('2. 访问IP限制页面...');
    await page.goto('http://localhost:8081/admin/ip-restriction', { waitUntil: 'networkidle2' });
    await new Promise(resolve => setTimeout(resolve, 3000));
    
    // 点击所有标签页
    console.log('3. 测试所有标签页...\n');
    
    const tabs = await page.$$('.el-tabs__item');
    console.log(`   发现 ${tabs.length} 个标签页`);
    
    for (let i = 0; i < tabs.length; i++) {
      const tabText = await page.evaluate(el => el.textContent, tabs[i]);
      console.log(`   点击标签: ${tabText.trim()}`);
      await tabs[i].click();
      await new Promise(resolve => setTimeout(resolve, 1500));
    }
    
    // 检查错误消息
    const errorMessages = await page.evaluate(() => {
      const errors = [];
      document.querySelectorAll('.el-message').forEach(el => {
        if (el.classList.contains('el-message--error')) {
          errors.push(el.textContent);
        }
      });
      return errors;
    });
    
    console.log('\n========================================');
    console.log('测试结果');
    console.log('========================================\n');
    
    // 显示所有API调用
    console.log('API调用汇总:');
    const uniqueAPIs = [...new Set(allResponses.map(r => `${r.status} ${r.url.replace('http://localhost:8081', '')}`))];
    uniqueAPIs.forEach(api => {
      const status = parseInt(api.split(' ')[0]);
      const icon = status >= 400 ? '❌' : '✅';
      console.log(`  ${icon} ${api}`);
    });
    
    // 显示失败的API详情
    const failedAPIs = allResponses.filter(r => r.status >= 400);
    if (failedAPIs.length > 0) {
      console.log('\n失败的API详情:');
      failedAPIs.forEach(api => {
        console.log(`\n  ${api.status} ${api.url}`);
        console.log(`  响应:`, JSON.stringify(api.body, null, 2));
      });
    }
    
    // 显示错误消息
    if (errorMessages.length > 0) {
      console.log('\n❌ 页面错误消息:');
      errorMessages.forEach(msg => console.log(`  ${msg}`));
    } else {
      console.log('\n✅ 没有页面错误消息');
    }
    
    // 显示控制台错误
    const errors = consoleMessages.filter(m => m.type === 'error');
    if (errors.length > 0) {
      console.log('\n控制台错误:');
      errors.forEach(err => console.log(`  ${err.text}`));
    }
    
    // 截图
    await page.screenshot({ path: 'ip-restriction-debug.png', fullPage: true });
    console.log('\n[截图已保存]: ip-restriction-debug.png');
    
  } catch (error) {
    console.error('\n[测试失败]:', error.message);
    console.error(error.stack);
    await page.screenshot({ path: 'error-debug.png', fullPage: true });
  } finally {
    await browser.close();
  }
})();
