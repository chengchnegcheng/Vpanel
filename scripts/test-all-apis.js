#!/usr/bin/env node

/**
 * 全面的API测试脚本
 * 测试用户门户和后台管理的所有关键API端点
 */

const puppeteer = require('puppeteer-core');
const fs = require('fs');
const path = require('path');

// 配置
const CONFIG = {
  baseURL: process.env.BASE_URL || 'http://localhost:8080',
  adminUsername: process.env.ADMIN_USER || 'admin',
  adminPassword: process.env.ADMIN_PASS || 'admin123',
  userUsername: process.env.USER_NAME || 'testuser',
  userPassword: process.env.USER_PASS || 'test123',
  headless: process.env.HEADLESS !== 'false',
  slowMo: parseInt(process.env.SLOW_MO || '0'),
  timeout: parseInt(process.env.TIMEOUT || '30000'),
};

// 测试结果
const results = {
  total: 0,
  passed: 0,
  failed: 0,
  errors: [],
  apiCalls: [],
};

// 日志函数
function log(message, type = 'info') {
  const timestamp = new Date().toISOString();
  const prefix = {
    info: '📝',
    success: '✅',
    error: '❌',
    warning: '⚠️',
    api: '🔌',
  }[type] || '📝';
  
  console.log(`${prefix} [${timestamp}] ${message}`);
}

// 记录API调用
function recordAPICall(method, url, status, error = null) {
  const call = {
    method,
    url,
    status,
    error,
    timestamp: new Date().toISOString(),
  };
  results.apiCalls.push(call);
  
  if (error) {
    log(`API ${method} ${url} - Status: ${status} - Error: ${error}`, 'error');
  } else {
    log(`API ${method} ${url} - Status: ${status}`, 'api');
  }
}

// 测试函数
async function runTest(name, testFn) {
  results.total++;
  log(`Running test: ${name}`, 'info');
  
  try {
    await testFn();
    results.passed++;
    log(`Test passed: ${name}`, 'success');
    return true;
  } catch (error) {
    results.failed++;
    results.errors.push({ test: name, error: error.message, stack: error.stack });
    log(`Test failed: ${name} - ${error.message}`, 'error');
    return false;
  }
}

// 启动浏览器
async function launchBrowser() {
  log('Launching browser...', 'info');
  
  const executablePath = process.env.CHROME_PATH || 
    '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
  
  if (!fs.existsSync(executablePath)) {
    throw new Error(`Chrome not found at: ${executablePath}`);
  }
  
  const browser = await puppeteer.launch({
    executablePath,
    headless: CONFIG.headless,
    slowMo: CONFIG.slowMo,
    args: [
      '--no-sandbox',
      '--disable-setuid-sandbox',
      '--disable-dev-shm-usage',
      '--disable-web-security',
    ],
  });
  
  log('Browser launched successfully', 'success');
  return browser;
}

// 创建页面并设置监听
async function createPage(browser) {
  const page = await browser.newPage();
  await page.setViewport({ width: 1920, height: 1080 });
  
  // 监听所有网络请求
  page.on('response', async (response) => {
    const url = response.url();
    const status = response.status();
    const method = response.request().method();
    
    // 只记录API调用
    if (url.includes('/api/')) {
      let error = null;
      if (status >= 400) {
        try {
          const text = await response.text();
          error = text.substring(0, 200);
        } catch (e) {
          error = 'Failed to read response';
        }
      }
      recordAPICall(method, url, status, error);
    }
  });
  
  // 监听控制台错误
  page.on('console', msg => {
    if (msg.type() === 'error') {
      log(`Console error: ${msg.text()}`, 'warning');
    }
  });
  
  // 监听页面错误
  page.on('pageerror', error => {
    log(`Page error: ${error.message}`, 'error');
  });
  
  return page;
}

// 等待并点击
async function clickAndWait(page, selector, waitTime = 1000) {
  await page.waitForSelector(selector, { timeout: CONFIG.timeout });
  await page.click(selector);
  await page.waitForTimeout(waitTime);
}

// 输入文本
async function typeText(page, selector, text) {
  await page.waitForSelector(selector, { timeout: CONFIG.timeout });
  await page.type(selector, text);
}

// ==================== 后台管理测试 ====================

async function testAdminLogin(page) {
  await runTest('Admin Login', async () => {
    log('Testing admin login...', 'info');
    
    await page.goto(`${CONFIG.baseURL}/login`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    // 输入用户名和密码
    await typeText(page, 'input[type="text"], input[placeholder*="用户名"]', CONFIG.adminUsername);
    await typeText(page, 'input[type="password"]', CONFIG.adminPassword);
    
    // 点击登录
    await clickAndWait(page, 'button[type="submit"], button:has-text("登录")', 3000);
    
    // 验证登录成功
    const url = page.url();
    if (!url.includes('/dashboard') && !url.includes('/admin')) {
      throw new Error(`Login failed, current URL: ${url}`);
    }
    
    log('Admin login successful', 'success');
  });
}

async function testAdminAPIs(page) {
  // 测试用户列表API
  await runTest('Admin - Get Users List', async () => {
    await page.goto(`${CONFIG.baseURL}/admin/users`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    // 检查是否有API错误
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/users') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Users API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
  
  // 测试代理列表API
  await runTest('Admin - Get Proxies List', async () => {
    await page.goto(`${CONFIG.baseURL}/admin/proxies`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/proxies') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Proxies API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
  
  // 测试订单列表API
  await runTest('Admin - Get Orders List', async () => {
    await page.goto(`${CONFIG.baseURL}/admin/orders`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/admin/orders') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Orders API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
  
  // 测试套餐列表API
  await runTest('Admin - Get Plans List', async () => {
    await page.goto(`${CONFIG.baseURL}/admin/plans`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/admin/plans') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Plans API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
  
  // 测试系统信息API
  await runTest('Admin - Get System Info', async () => {
    await page.goto(`${CONFIG.baseURL}/admin/system`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/system') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`System API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
  
  // 测试统计数据API
  await runTest('Admin - Get Stats', async () => {
    await page.goto(`${CONFIG.baseURL}/admin/dashboard`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/stats') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Stats API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
}

// ==================== 用户门户测试 ====================

async function testUserLogin(page) {
  await runTest('User Login', async () => {
    log('Testing user login...', 'info');
    
    // 先登出管理员
    await page.goto(`${CONFIG.baseURL}/logout`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(1000);
    
    await page.goto(`${CONFIG.baseURL}/user/login`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    // 输入用户名和密码
    await typeText(page, 'input[type="text"], input[placeholder*="用户名"]', CONFIG.userUsername);
    await typeText(page, 'input[type="password"]', CONFIG.userPassword);
    
    // 点击登录
    await clickAndWait(page, 'button[type="submit"], button:has-text("登录")', 3000);
    
    // 验证登录成功
    const url = page.url();
    if (!url.includes('/user/dashboard') && !url.includes('/user')) {
      throw new Error(`User login failed, current URL: ${url}`);
    }
    
    log('User login successful', 'success');
  });
}

async function testUserAPIs(page) {
  // 测试用户信息API
  await runTest('User - Get Profile', async () => {
    await page.goto(`${CONFIG.baseURL}/user/profile`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/auth/me') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Profile API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
  
  // 测试订阅信息API
  await runTest('User - Get Subscription', async () => {
    await page.goto(`${CONFIG.baseURL}/user/subscription`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/subscription') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Subscription API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
  
  // 测试套餐列表API
  await runTest('User - Get Plans', async () => {
    await page.goto(`${CONFIG.baseURL}/user/plans`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/plans') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Plans API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
  
  // 测试订单列表API
  await runTest('User - Get Orders', async () => {
    await page.goto(`${CONFIG.baseURL}/user/orders`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/orders') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Orders API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
  
  // 测试余额API
  await runTest('User - Get Balance', async () => {
    await page.goto(`${CONFIG.baseURL}/user/balance`, { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    const errorAPIs = results.apiCalls.filter(call => 
      call.url.includes('/api/balance') && call.status >= 400
    );
    
    if (errorAPIs.length > 0) {
      throw new Error(`Balance API failed: ${JSON.stringify(errorAPIs)}`);
    }
  });
}

// ==================== 主测试流程 ====================

async function main() {
  log('Starting comprehensive API tests...', 'info');
  log(`Base URL: ${CONFIG.baseURL}`, 'info');
  
  let browser;
  let page;
  
  try {
    browser = await launchBrowser();
    page = await createPage(browser);
    
    // 测试后台管理
    log('=== Testing Admin Portal ===', 'info');
    await testAdminLogin(page);
    await testAdminAPIs(page);
    
    // 测试用户门户
    log('=== Testing User Portal ===', 'info');
    await testUserLogin(page);
    await testUserAPIs(page);
    
  } catch (error) {
    log(`Fatal error: ${error.message}`, 'error');
    results.errors.push({ test: 'Main', error: error.message, stack: error.stack });
  } finally {
    if (browser) {
      await browser.close();
    }
  }
  
  // 生成报告
  generateReport();
}

function generateReport() {
  log('\n=== Test Results ===', 'info');
  log(`Total tests: ${results.total}`, 'info');
  log(`Passed: ${results.passed}`, 'success');
  log(`Failed: ${results.failed}`, results.failed > 0 ? 'error' : 'info');
  
  // 统计API调用
  const apiStats = {
    total: results.apiCalls.length,
    success: results.apiCalls.filter(c => c.status < 400).length,
    clientError: results.apiCalls.filter(c => c.status >= 400 && c.status < 500).length,
    serverError: results.apiCalls.filter(c => c.status >= 500).length,
  };
  
  log('\n=== API Call Statistics ===', 'info');
  log(`Total API calls: ${apiStats.total}`, 'info');
  log(`Successful (2xx-3xx): ${apiStats.success}`, 'success');
  log(`Client errors (4xx): ${apiStats.clientError}`, apiStats.clientError > 0 ? 'error' : 'info');
  log(`Server errors (5xx): ${apiStats.serverError}`, apiStats.serverError > 0 ? 'error' : 'info');
  
  // 显示失败的API调用
  const failedAPIs = results.apiCalls.filter(c => c.status >= 400);
  if (failedAPIs.length > 0) {
    log('\n=== Failed API Calls ===', 'error');
    failedAPIs.forEach(call => {
      log(`${call.method} ${call.url} - Status: ${call.status}`, 'error');
      if (call.error) {
        log(`  Error: ${call.error}`, 'error');
      }
    });
  }
  
  // 显示测试错误
  if (results.errors.length > 0) {
    log('\n=== Test Errors ===', 'error');
    results.errors.forEach(err => {
      log(`Test: ${err.test}`, 'error');
      log(`Error: ${err.error}`, 'error');
    });
  }
  
  // 保存详细报告
  const reportPath = path.join(__dirname, 'api-test-report.json');
  fs.writeFileSync(reportPath, JSON.stringify(results, null, 2));
  log(`\nDetailed report saved to: ${reportPath}`, 'info');
  
  // 退出码
  process.exit(results.failed > 0 ? 1 : 0);
}

// 运行测试
main().catch(error => {
  log(`Unhandled error: ${error.message}`, 'error');
  console.error(error);
  process.exit(1);
});
