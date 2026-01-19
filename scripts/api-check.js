#!/usr/bin/env node

/**
 * API端点检查脚本
 * 检查前端调用的API是否在后端路由中存在
 */

const fs = require('fs');
const path = require('path');

// 从routes.go提取的后端API路由
const backendRoutes = [
  // Auth routes
  'POST /api/auth/login',
  'POST /api/auth/refresh',
  'POST /api/auth/logout',
  'GET /api/auth/me',
  'PUT /api/auth/password',
  
  // Proxy routes
  'GET /api/proxies',
  'POST /api/proxies',
  'POST /api/proxies/batch',
  'GET /api/proxies/:id',
  'PUT /api/proxies/:id',
  'DELETE /api/proxies/:id',
  'GET /api/proxies/:id/link',
  'POST /api/proxies/:id/toggle',
  'POST /api/proxies/:id/start',
  'POST /api/proxies/:id/stop',
  'GET /api/proxies/:id/stats',
  
  // System routes
  'GET /api/system/info',
  'GET /api/system/status',
  'GET /api/system/stats',
  
  // Stats routes
  'GET /api/stats/dashboard',
  'GET /api/stats/protocol',
  'GET /api/stats/traffic',
  'GET /api/stats/user',
  'GET /api/stats/detailed',
  
  // User routes (admin)
  'GET /api/users',
  'POST /api/users',
  'GET /api/users/:id',
  'PUT /api/users/:id',
  'DELETE /api/users/:id',
  'POST /api/users/:id/enable',
  'POST /api/users/:id/disable',
  'POST /api/users/:id/reset-password',
  'GET /api/users/:id/login-history',
  'DELETE /api/users/:id/login-history',
  
  // Settings routes (admin)
  'GET /api/settings',
  'PUT /api/settings',
  'POST /api/settings/backup',
  'POST /api/settings/restore',
  'GET /api/settings/xray',
  'POST /api/settings/xray',
  'GET /api/settings/protocols',
  'POST /api/settings/protocols',
  'GET /api/settings/ip-restriction',
  'PUT /api/settings/ip-restriction',
  
  // Xray routes (admin)
  'GET /api/xray/status',
  'POST /api/xray/start',
  'POST /api/xray/stop',
  'POST /api/xray/restart',
  'GET /api/xray/config',
  'PUT /api/xray/config',
  'POST /api/xray/validate',
  'POST /api/xray/test-config',
  'GET /api/xray/version',
  'GET /api/xray/versions',
  'POST /api/xray/sync-versions',
  'GET /api/xray/check-updates',
  'POST /api/xray/download',
  'POST /api/xray/install',
  'POST /api/xray/update',
  'POST /api/xray/switch-version',
  
  // Subscription routes
  'GET /api/subscription/:token',
  'GET /s/:code',
  'GET /api/subscription/link',
  'GET /api/subscription/info',
  'POST /api/subscription/regenerate',
  
  // Admin subscription routes
  'GET /api/admin/subscriptions',
  'DELETE /api/admin/subscriptions/:user_id',
  'POST /api/admin/subscriptions/:user_id/reset-stats',
  
  // Plan routes
  'GET /api/plans',
  'GET /api/plans/:id',
  'GET /api/plans/:id/prices',
  'GET /api/plans-with-prices',
  
  // Admin plan routes
  'GET /api/admin/plans',
  'POST /api/admin/plans',
  'PUT /api/admin/plans/:id',
  'DELETE /api/admin/plans/:id',
  'PUT /api/admin/plans/:id/status',
  'PUT /api/admin/plans/:id/prices',
  'DELETE /api/admin/plans/:id/prices/:currency',
  
  // Currency routes
  'GET /api/currencies',
  'GET /api/currencies/detect',
  'GET /api/currencies/rate',
  'POST /api/currencies/convert',
  
  // Admin currency routes
  'POST /api/admin/currencies/update-rates',
  
  // Order routes
  'POST /api/orders',
  'GET /api/orders',
  'GET /api/orders/:id',
  'POST /api/orders/:id/cancel',
  
  // Admin order routes
  'GET /api/admin/orders',
  'PUT /api/admin/orders/:id/status',
  
  // Payment routes
  'POST /api/payments/create',
  'GET /api/payments/status/:orderNo',
  'GET /api/payments/methods',
  'POST /api/payments/switch-method',
  'POST /api/payments/retry',
  'GET /api/payments/retry/:orderID',
  'POST /api/payments/callback/:method',
  
  // Balance routes
  'GET /api/balance',
  'GET /api/balance/transactions',
  
  // Admin balance routes
  'POST /api/admin/balance/adjust',
  
  // Coupon routes
  'POST /api/coupons/validate',
  
  // Admin coupon routes
  'GET /api/admin/coupons',
  'POST /api/admin/coupons',
  'DELETE /api/admin/coupons/:id',
  'POST /api/admin/coupons/batch',
  
  // Invite routes
  'GET /api/invite/code',
  'GET /api/invite/referrals',
  'GET /api/invite/stats',
  'GET /api/invite/commissions',
  'GET /api/invite/earnings',
  
  // Invoice routes
  'GET /api/invoices',
  'GET /api/invoices/:id/download',
  
  // Admin invoice routes
  'POST /api/admin/invoices/generate',
  
  // Report routes (admin)
  'GET /api/admin/reports/revenue',
  'GET /api/admin/reports/orders',
  'GET /api/admin/reports/failed-payments',
  'GET /api/admin/reports/pause-stats',
  
  // Trial routes
  'GET /api/trial',
  'POST /api/trial/activate',
  
  // Admin trial routes
  'GET /api/admin/trials',
  'GET /api/admin/trials/stats',
  'POST /api/admin/trials/grant',
  'GET /api/admin/trials/user/:user_id',
  'POST /api/admin/trials/expire',
  
  // Plan change routes
  'POST /api/plan-change/calculate',
  'POST /api/plan-change/upgrade',
  'POST /api/plan-change/downgrade',
  'GET /api/plan-change/downgrade',
  'DELETE /api/plan-change/downgrade',
  
  // Subscription pause routes
  'GET /api/subscription/pause',
  'POST /api/subscription/pause',
  'GET /api/subscription/pause/history',
  'POST /api/subscription/resume',
  
  // Admin pause routes
  'GET /api/admin/subscription/pause/stats',
  'POST /api/admin/subscription/pause/auto-resume',
  
  // Gift card routes
  'POST /api/gift-cards/redeem',
  'GET /api/gift-cards',
  'POST /api/gift-cards/validate',
  
  // Admin gift card routes
  'GET /api/admin/gift-cards',
  'POST /api/admin/gift-cards/batch',
  'GET /api/admin/gift-cards/stats',
  'GET /api/admin/gift-cards/:id',
  'PUT /api/admin/gift-cards/:id/status',
  'DELETE /api/admin/gift-cards/:id',
  'GET /api/admin/gift-cards/batch/:batch_id/stats',
  
  // Node routes (admin)
  'GET /api/admin/nodes',
  'POST /api/admin/nodes',
  'GET /api/admin/nodes/statistics',
  'GET /api/admin/nodes/:id',
  'PUT /api/admin/nodes/:id',
  'DELETE /api/admin/nodes/:id',
  'PUT /api/admin/nodes/:id/status',
  'POST /api/admin/nodes/:id/token',
  'POST /api/admin/nodes/:id/token/rotate',
  'POST /api/admin/nodes/:id/token/revoke',
  'POST /api/admin/nodes/:id/health-check',
  'GET /api/admin/nodes/:id/health-history',
  'GET /api/admin/nodes/:id/health-latest',
  'GET /api/admin/nodes/:id/health-stats',
  'POST /api/admin/nodes/health-check',
  'GET /api/admin/nodes/cluster-health',
  'GET /api/admin/nodes/traffic/total',
  'GET /api/admin/nodes/traffic/by-node',
  'GET /api/admin/nodes/traffic/by-group',
  'GET /api/admin/nodes/traffic/aggregated',
  'GET /api/admin/nodes/traffic/realtime',
  'POST /api/admin/nodes/traffic',
  'POST /api/admin/nodes/traffic/batch',
  'POST /api/admin/nodes/traffic/cleanup',
  'GET /api/admin/nodes/:id/traffic',
  'GET /api/admin/nodes/:id/traffic/top-users',
  
  // Node group routes (admin)
  'GET /api/admin/node-groups',
  'POST /api/admin/node-groups',
  'GET /api/admin/node-groups/with-stats',
  'GET /api/admin/node-groups/stats',
  'GET /api/admin/node-groups/:id',
  'PUT /api/admin/node-groups/:id',
  'DELETE /api/admin/node-groups/:id',
  'GET /api/admin/node-groups/:id/stats',
  'GET /api/admin/node-groups/:id/nodes',
  'PUT /api/admin/node-groups/:id/nodes',
  'POST /api/admin/node-groups/:id/nodes/:node_id',
  'DELETE /api/admin/node-groups/:id/nodes/:node_id',
  'GET /api/admin/node-groups/:id/traffic',
  
  // Health checker routes (admin)
  'GET /api/admin/health-checker/status',
  'POST /api/admin/health-checker/start',
  'POST /api/admin/health-checker/stop',
  'PUT /api/admin/health-checker/config',
  
  // IP restriction routes (admin)
  'GET /api/admin/ip-restrictions/stats',
  'GET /api/admin/ip-restrictions/online',
  'GET /api/admin/ip-restrictions/history',
  'GET /api/admin/ip-whitelist',
  'POST /api/admin/ip-whitelist',
  'DELETE /api/admin/ip-whitelist/:id',
  'POST /api/admin/ip-whitelist/import',
  'GET /api/admin/ip-blacklist',
  'POST /api/admin/ip-blacklist',
  'DELETE /api/admin/ip-blacklist/:id',
  'GET /api/admin/users/:id/online-ips',
  'POST /api/admin/users/:id/kick-ip',
  'GET /api/admin/users/:id/node-traffic',
  'GET /api/admin/users/:id/node-traffic/breakdown',
  
  // User device routes
  'GET /api/user/devices',
  'POST /api/user/devices/:ip/kick',
  'GET /api/user/ip-stats',
  'GET /api/user/ip-history',
  
  // Certificates routes (admin)
  'GET /api/certificates',
  'POST /api/certificates/apply',
  'POST /api/certificates/upload',
  'POST /api/certificates/:id/renew',
  'GET /api/certificates/:id/validate',
  'DELETE /api/certificates/:id',
  'PUT /api/certificates/:id/auto-renew',
  
  // Logs routes (admin)
  'GET /api/logs',
  'GET /api/logs/export',
  'GET /api/logs/:id',
  'DELETE /api/logs',
  'POST /api/logs/cleanup',
  
  // Role routes
  'GET /api/roles',
  'POST /api/roles',
  'GET /api/roles/:id',
  'PUT /api/roles/:id',
  'DELETE /api/roles/:id',
  'GET /api/permissions',
  
  // Portal routes
  'POST /api/portal/auth/register',
  'POST /api/portal/auth/login',
  'POST /api/portal/auth/forgot-password',
  'POST /api/portal/auth/reset-password',
  'GET /api/portal/auth/verify-email',
  'POST /api/portal/auth/2fa/login',
  'POST /api/portal/auth/logout',
  'GET /api/portal/auth/profile',
  'PUT /api/portal/auth/profile',
  'PUT /api/portal/auth/password',
  'POST /api/portal/auth/2fa/enable',
  'POST /api/portal/auth/2fa/verify',
  'POST /api/portal/auth/2fa/disable',
  'GET /api/portal/dashboard',
  'GET /api/portal/dashboard/traffic',
  'GET /api/portal/dashboard/announcements',
  'GET /api/portal/nodes',
  'GET /api/portal/nodes/:id',
  'POST /api/portal/nodes/:id/ping',
  'GET /api/portal/tickets',
  'POST /api/portal/tickets',
  'GET /api/portal/tickets/:id',
  'POST /api/portal/tickets/:id/reply',
  'POST /api/portal/tickets/:id/close',
  'POST /api/portal/tickets/:id/reopen',
  'GET /api/portal/announcements',
  'GET /api/portal/announcements/:id',
  'POST /api/portal/announcements/:id/read',
  'GET /api/portal/announcements/unread-count',
  'GET /api/portal/stats/traffic',
  'GET /api/portal/stats/usage',
  'GET /api/portal/stats/daily',
  'GET /api/portal/stats/export',
  'GET /api/portal/help/articles',
  'GET /api/portal/help/articles/:slug',
  'GET /api/portal/help/search',
  'GET /api/portal/help/featured',
  'GET /api/portal/help/categories',
  'POST /api/portal/help/articles/:slug/helpful',
  
  // Error reporting
  'POST /api/errors/report',
  
  // Health check
  'GET /health',
  'GET /ready',
];

// 前端调用的API（从前面的搜索结果中提取）
const frontendAPICalls = [
  'GET /api/gift-cards/stats', // 已修复为 /api/admin/gift-cards/stats
  'GET /api/admin/reports/pause-stats',
  'GET /api/admin/reports/failed-payments',
  'GET /api/admin/reports/revenue',
  'GET /api/admin/reports/orders',
  'GET /api/admin/trials/stats',
  'GET /api/trial',
  'GET /api/admin/trials/user/:user_id',
  'GET /api/proxies',
  'GET /api/monitor/stats',
  'GET /api/users',
  'GET /api/protocols',
  'GET /api/traffic',
  'GET /api/user/devices',
  'GET /api/user/ip-history',
  'GET /api/user/subscription-ips',
  'GET /api/admin/ip-restrictions/stats',
  'GET /api/admin/settings/ip-restriction',
  'GET /api/admin/ip-whitelist',
  'GET /api/admin/ip-blacklist',
  'GET /api/admin/ip-restrictions/online',
  'GET /api/admin/ip-restrictions/history',
  'GET /api/traffic/monitor',
  'GET /api/stats/dashboard',
  'GET /api/stats/traffic',
  'GET /api/proxies/:id/link',
  'GET /api/xray/versions',
  'GET /api/system/info',
  'GET /api/xray/status',
  'GET /api/settings/xray',
  'GET /api/settings/protocols',
  'GET /api/xray/check-updates',
  'GET /api/portal/stats/traffic',
  'GET /api/portal/stats/usage',
];

console.log('=== API端点检查 ===\n');

// 检查前端调用的API是否在后端存在
const missingAPIs = [];
const foundAPIs = [];

frontendAPICalls.forEach(apiCall => {
  const found = backendRoutes.some(route => {
    // 简单匹配，忽略参数
    const apiPattern = apiCall.replace(/:\w+/g, ':id');
    const routePattern = route.replace(/:\w+/g, ':id');
    return apiPattern === routePattern;
  });
  
  if (found) {
    foundAPIs.push(apiCall);
  } else {
    missingAPIs.push(apiCall);
  }
});

console.log(`✓ 找到 ${foundAPIs.length} 个匹配的API端点`);
console.log(`✗ 缺失 ${missingAPIs.length} 个API端点\n`);

if (missingAPIs.length > 0) {
  console.log('缺失的API端点：');
  missingAPIs.forEach(api => {
    console.log(`  ✗ ${api}`);
  });
  console.log('');
}

// 检查特定的问题API
console.log('=== 特定问题检查 ===\n');

const problemAPIs = [
  { api: 'GET /api/admin/gift-cards/stats', description: '礼品卡统计' },
  { api: 'GET /api/admin/reports/pause-stats', description: '暂停统计' },
  { api: 'GET /api/portal/stats/usage', description: '用户门户使用统计' },
  { api: 'GET /api/portal/stats/traffic', description: '用户门户流量统计' },
];

problemAPIs.forEach(({ api, description }) => {
  const exists = backendRoutes.includes(api);
  console.log(`${exists ? '✓' : '✗'} ${description}: ${api}`);
});

console.log('\n=== 检查完成 ===');

if (missingAPIs.length > 0) {
  console.log('\n需要在后端添加以下路由：');
  missingAPIs.forEach(api => {
    console.log(`  - ${api}`);
  });
  process.exit(1);
} else {
  console.log('\n所有API端点都已正确配置！');
  process.exit(0);
}
