const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({
    headless: false, // 显示浏览器
    args: ['--no-sandbox', '--disable-setuid-sandbox'],
    devtools: true // 打开开发者工具
  });
  
  const page = await browser.newPage();
  
  // 监听所有请求和响应
  page.on('request', request => {
    const url = request.url();
    if (url.includes('/api/admin/ip')) {
      console.log(`→ ${request.method()} ${url}`);
    }
  });
  
  page.on('response', async response => {
    const url = response.url();
    if (url.includes('/api/admin/ip')) {
      const status = response.status();
      console.log(`← ${status} ${url}`);
      
      if (status >= 400) {
        try {
          const body = await response.text();
          console.log(`ERROR BODY:`, body);
        } catch (e) {}
      }
    }
  });
  
  // 监听控制台
  page.on('console', msg => {
    const text = msg.text();
    if (text.includes('Failed') || text.includes('error') || text.includes('ERR-')) {
      console.log(`[CONSOLE ${msg.type()}]`, text);
    }
  });
  
  try {
    console.log('访问登录页面...');
    await page.goto('http://localhost:8081/login', { waitUntil: 'networkidle2' });
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    console.log('输入凭据...');
    await page.type('input[type="text"], input[placeholder*="用户"]', 'admin');
    await page.type('input[type="password"], input[placeholder*="密码"]', 'admin123');
    
    console.log('点击登录...');
    await page.click('button[type="submit"], button:has-text("登录")');
    await page.waitForNavigation({ waitUntil: 'networkidle2', timeout: 10000 }).catch(() => {});
    await new Promise(resolve => setTimeout(resolve, 3000));
    
    console.log('\n访问IP限制页面...');
    await page.goto('http://localhost:8081/admin/ip-restriction', { waitUntil: 'networkidle2' });
    
    console.log('\n等待10秒观察错误...');
    await new Promise(resolve => setTimeout(resolve, 10000));
    
    console.log('\n按任意键关闭浏览器...');
    
  } catch (error) {
    console.error('错误:', error.message);
  }
})();
