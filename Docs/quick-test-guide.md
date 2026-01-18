# macOS 显示问题修复 - 快速测试指南

## 已完成的修复

✅ 已更新以下文件：
1. `web/src/styles/theme.css` - 添加 Element Plus 组件深色模式和高分辨率显示修复
2. `web/src/assets/styles/base.scss` - 添加全局 SVG 渲染优化
3. `web/src/views/Dashboard.vue` - 优化仪表板圆形图表显示
4. `web/dist/` - 已重新构建前端资源

## 测试步骤

### 1. 启动服务

```bash
# 方式一：使用脚本启动
./vpanel.sh start

# 方式二：直接运行
./v
```

### 2. 访问管理面板

打开浏览器访问：
```
http://localhost:8080/admin/dashboard
```

### 3. 验证修复效果

检查以下内容是否正常显示：

#### ✓ 系统概览部分
- [ ] CPU 使用率圆形图表显示完整
- [ ] 内存使用率圆形图表显示完整
- [ ] 磁盘使用率圆形图表显示完整
- [ ] 百分比数字清晰可见
- [ ] 图表颜色正确（绿色/黄色/红色）

#### ✓ 不同显示模式测试
- [ ] 浅色模式下显示正常
- [ ] 深色模式下显示正常（系统偏好设置 → 外观 → 深色）
- [ ] 缩放 50% 显示正常（Cmd + -）
- [ ] 缩放 100% 显示正常（Cmd + 0）
- [ ] 缩放 150% 显示正常（Cmd + +）
- [ ] 缩放 200% 显示正常（Cmd + +）

#### ✓ 浏览器兼容性测试
- [ ] Safari 显示正常
- [ ] Chrome 显示正常
- [ ] Firefox 显示正常（如已安装）

### 4. 强制刷新浏览器缓存

如果修复后仍有问题，请强制刷新：

```bash
# macOS 快捷键
Cmd + Shift + R  # Safari/Chrome/Firefox
```

或者清除浏览器缓存：
- Safari: 开发 → 清空缓存
- Chrome: 设置 → 隐私和安全 → 清除浏览数据

### 5. 检查浏览器控制台

如果仍有问题，打开开发者工具检查错误：

```bash
# 打开开发者工具
Cmd + Option + I  # Chrome/Firefox
Cmd + Option + C  # Safari（需先启用开发菜单）
```

查看 Console 标签页是否有错误信息。

## 修复原理

### 问题原因
1. **SVG 渲染问题**：macOS 在某些显示模式下，SVG 元素可能不使用硬件加速
2. **高分辨率显示**：Retina 显示器的像素密度导致 SVG 渲染精度问题
3. **深色模式兼容**：Element Plus 组件在深色模式下的颜色对比度不足

### 解决方案
1. **强制硬件加速**：使用 `transform: translateZ(0)` 触发 GPU 加速
2. **几何精度优化**：设置 `shape-rendering: geometricPrecision` 提高渲染质量
3. **颜色对比度增强**：为深色模式单独设置文字和背景颜色

## 技术细节

### CSS 修复关键代码

```css
/* 强制硬件加速 */
.el-progress {
  -webkit-transform: translateZ(0);
  transform: translateZ(0);
}

/* 高分辨率优化 */
@media (-webkit-min-device-pixel-ratio: 2) {
  svg {
    shape-rendering: geometricPrecision;
  }
}

/* 深色模式文字颜色 */
.dark .el-progress__text {
  color: var(--color-text-primary) !important;
}
```

## 如果问题仍然存在

### 方案 A：检查浏览器设置

1. **Chrome/Edge**
   - 设置 → 系统 → 使用硬件加速（确保已开启）

2. **Safari**
   - 开发 → 实验性功能 → WebGL 2.0（确保已开启）

### 方案 B：调整 macOS 显示设置

1. 打开"系统偏好设置"
2. 选择"显示器"
3. 尝试不同的缩放选项：
   - 默认
   - 更大文字
   - 更多空间

### 方案 C：使用不同浏览器

如果某个浏览器有问题，尝试使用其他浏览器：
- Safari（推荐，macOS 原生）
- Chrome
- Firefox
- Edge

### 方案 D：查看详细日志

```bash
# 查看浏览器控制台
# 按 F12 或 Cmd + Option + I

# 查看 Elements 标签
# 检查 .el-progress 元素的 computed styles
# 确认 transform 和 shape-rendering 属性是否生效
```

## 联系支持

如果以上方法都无法解决问题，请提供以下信息：

1. macOS 版本：系统偏好设置 → 关于本机
2. 浏览器版本：浏览器 → 关于
3. 显示器信息：系统偏好设置 → 显示器
4. 截图：显示问题的截图
5. 控制台错误：开发者工具中的错误信息

## 更新记录

- 2026-01-17: 初始修复版本
  - 添加 SVG 硬件加速
  - 优化高分辨率显示
  - 增强深色模式兼容性
