const puppeteer = require('puppeteer');

(async () => {
  console.log('========================================');
  console.log('IP限制错误检查');
  console.log('错误ID: ERR-MKJ8RI3F-GOCPYN');
  console.log('========================================\n');

  const browser = await puppeteer.launch({
    headless: false,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();
  
  // 监听控制台消息
  const consoleMessages = [];
  page.on('console', msg => {
    const text = msg.text();
    consoleMessages.push(text);
    console.log('浏览器控制台:', text);
  });

  // 监听网络请求
  const failedRequests = [];
  page.on('requestfailed', request => {
    failedRequests.push({
      url: request.url(),
      failure: request.failure()
    });
    console.log('❌ 请求失败:', request.url(), request.failure());
  });

  // 监听响应
  page.on('response', async response => {
    const url = response.url();
    const status = response.status();
    
    if (url.includes('/api/')) {
      console.log(`API响应: ${status} ${url}`);
      
      if (status >= 400) {
        try {
          const body = await response.text();
          console.log('错误响应体:', body);
        } catch (e) {
          console.log('无法读取响应体');
        }
      }
    }
  });

  try {
    // 1. 访问登录页面
    console.log('\n1. 访问登录页面...');
    await page.goto('http://localhost:8081/admin/login', { 
      waitUntil: 'networkidle2',
      timeout: 30000 
    });
    await new Promise(resolve => setTimeout(resolve, 2000));

    // 2. 登录
    console.log('\n2. 执行登录...');
    
    // 等待输入框出现
    await page.waitForSelector('input', { timeout: 5000 });
    
    console.log('等待登录完成...');
    await new Promise(resolve => setTimeout(resolve, 3000));

    // 检查是否登录成功
    const currentUrl = page.url();
    console.log('当前URL:', currentUrl);

    // 检查 localStorage
    const token = await page.evaluate(() => localStorage.getItem('token'));
    console.log('Token存在:', !!token);
    
    if (!token) {
      console.log('❌ 登录失败，没有获取到token');
      await browser.close();
      return;
    }

    // 3. 访问IP限制页面
    console.log('\n3. 访问IP限制页面...');
    await page.goto('http://localhost:8081/admin/ip-restriction', { 
      waitUntil: 'networkidle2',
      timeout: 30000 
    });
    await new Promise(resolve => setTimeout(resolve, 3000));

    // 4. 检查页面内容
    console.log('\n4. 检查页面内容...');
    
    // 检查是否有错误提示
    const errorElements = await page.$$('.el-message--error, .error-message, [class*="error"]');
    if (errorElements.length > 0) {
      console.log('❌ 发现错误元素:', errorElements.length);
      for (const el of errorElements) {
        const text = await page.evaluate(e => e.textContent, el);
        console.log('  错误内容:', text);
      }
    }

    // 检查是否有应用错误
    const appError = await page.$('.app-error, [class*="app-error"]');
    if (appError) {
      const errorText = await page.evaluate(e => e.textContent, appError);
      console.log('❌ 应用错误:', errorText);
    }

    // 5. 手动测试API
    console.log('\n5. 手动测试API...');
    const apiTest = await page.evaluate(async (token) => {
      try {
        const response = await fetch('http://localhost:8081/api/admin/ip-restrictions/stats', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });
        
        const data = await response.json();
        return {
          status: response.status,
          ok: response.ok,
          data: data
        };
      } catch (error) {
        return {
          error: error.message
        };
      }
    }, token);

    console.log('API测试结果:');
    console.log(JSON.stringify(apiTest, null, 2));

    // 6. 检查网络错误
    console.log('\n6. 网络错误汇总:');
    if (failedRequests.length > 0) {
      console.log('失败的请求:', failedRequests.length);
      failedRequests.forEach(req => {
        console.log('  -', req.url, req.failure);
      });
    } else {
      console.log('✓ 没有失败的请求');
    }

    // 7. 检查控制台错误
    console.log('\n7. 控制台消息汇总:');
    const errors = consoleMessages.filter(msg => 
      msg.includes('error') || 
      msg.includes('Error') || 
      msg.includes('ERR-') ||
      msg.includes('failed') ||
      msg.includes('Failed')
    );
    
    if (errors.length > 0) {
      console.log('发现错误消息:', errors.length);
      errors.forEach(err => console.log('  -', err));
    } else {
      console.log('✓ 没有错误消息');
    }

    // 8. 截图
    console.log('\n8. 保存截图...');
    await page.screenshot({ path: 'ip-restriction-error.png', fullPage: true });
    console.log('✓ 截图已保存: ip-restriction-error.png');

    console.log('\n========================================');
    console.log('检查完成');
    console.log('========================================');

  } catch (error) {
    console.error('\n❌ 测试过程中出错:', error);
  } finally {
    await browser.close();
  }
})();
