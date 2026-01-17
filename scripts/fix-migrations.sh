#!/bin/bash
# 手动执行 SQL 迁移脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 数据库路径
DB_PATH="${1:-data/v.db}"

if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}错误: 数据库文件不存在: $DB_PATH${NC}"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}     执行 SQL 迁移${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查 SQLite 是否安装
if ! command -v sqlite3 &> /dev/null; then
    echo -e "${RED}错误: sqlite3 未安装${NC}"
    exit 1
fi

# 迁移文件目录
MIGRATIONS_DIR="internal/database/migrations"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${RED}错误: 迁移目录不存在: $MIGRATIONS_DIR${NC}"
    exit 1
fi

# 创建迁移表
echo -e "${YELLOW}创建迁移表...${NC}"
sqlite3 "$DB_PATH" <<EOF
CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255),
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
EOF
echo -e "${GREEN}✓ 迁移表已创建${NC}"
echo ""

# 获取已应用的迁移
APPLIED_MIGRATIONS=$(sqlite3 "$DB_PATH" "SELECT version FROM migrations;" 2>/dev/null || echo "")

# 执行所有 SQL 迁移文件
echo -e "${YELLOW}执行迁移文件...${NC}"
for migration_file in "$MIGRATIONS_DIR"/*.sql; do
    if [ -f "$migration_file" ]; then
        filename=$(basename "$migration_file")
        version=$(echo "$filename" | cut -d'_' -f1)
        name=$(echo "$filename" | sed 's/^[0-9]*_//' | sed 's/.sql$//')
        
        # 检查是否已应用
        if echo "$APPLIED_MIGRATIONS" | grep -q "^${version}$"; then
            echo -e "${BLUE}⊙${NC} $filename (已应用)"
            continue
        fi
        
        echo -e "${YELLOW}→${NC} 应用 $filename..."
        
        # 执行迁移
        if sqlite3 "$DB_PATH" < "$migration_file" 2>&1; then
            # 记录迁移
            sqlite3 "$DB_PATH" "INSERT INTO migrations (version, name, applied_at) VALUES ('$version', '$name', datetime('now'));"
            echo -e "${GREEN}✓${NC} $filename 应用成功"
        else
            echo -e "${RED}✗${NC} $filename 应用失败"
            exit 1
        fi
    fi
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}     所有迁移已完成${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 显示迁移状态
echo -e "${YELLOW}迁移状态:${NC}"
sqlite3 "$DB_PATH" "SELECT version, name, applied_at FROM migrations ORDER BY version;" -header -column
