#!/bin/bash
# 智能迁移修复脚本 - 跳过已存在的列

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

DB_PATH="${1:-data/v.db}"

if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}错误: 数据库文件不存在: $DB_PATH${NC}"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}     智能执行 SQL 迁移${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查 SQLite
if ! command -v sqlite3 &> /dev/null; then
    echo -e "${RED}错误: sqlite3 未安装${NC}"
    exit 1
fi

MIGRATIONS_DIR="internal/database/migrations"

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

# 处理单个 SQL 语句
execute_statement() {
    local stmt="$1"
    local result
    
    # 跳过空语句和注释
    if [ -z "$stmt" ] || [[ "$stmt" =~ ^[[:space:]]*-- ]]; then
        return 0
    fi
    
    # 尝试执行语句
    result=$(sqlite3 "$DB_PATH" "$stmt" 2>&1)
    local exit_code=$?
    
    # 检查是否是"列已存在"或"表已存在"错误
    if [ $exit_code -ne 0 ]; then
        if echo "$result" | grep -qi "duplicate column\|already exists"; then
            echo -e "  ${BLUE}⊙${NC} 跳过（已存在）: $(echo "$stmt" | head -c 60)..."
            return 0
        else
            echo -e "  ${RED}✗${NC} 失败: $result"
            return 1
        fi
    fi
    
    return 0
}

# 执行迁移文件
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
        
        # 读取文件并逐语句执行
        success=true
        while IFS= read -r line || [ -n "$line" ]; do
            # 累积语句直到遇到分号
            if [ -z "$current_stmt" ]; then
                current_stmt="$line"
            else
                current_stmt="$current_stmt
$line"
            fi
            
            # 如果行包含分号，执行语句
            if echo "$line" | grep -q ";"; then
                if ! execute_statement "$current_stmt"; then
                    success=false
                    break
                fi
                current_stmt=""
            fi
        done < "$migration_file"
        
        # 执行最后一条语句（如果有）
        if [ -n "$current_stmt" ] && [ "$success" = true ]; then
            if ! execute_statement "$current_stmt"; then
                success=false
            fi
            current_stmt=""
        fi
        
        if [ "$success" = true ]; then
            # 记录迁移
            sqlite3 "$DB_PATH" "INSERT INTO migrations (version, name, applied_at) VALUES ('$version', '$name', datetime('now'));" 2>/dev/null || true
            echo -e "${GREEN}✓${NC} $filename 应用成功"
        else
            echo -e "${RED}✗${NC} $filename 应用失败（部分语句可能已执行）"
            # 仍然记录迁移，避免重复尝试
            sqlite3 "$DB_PATH" "INSERT OR IGNORE INTO migrations (version, name, applied_at) VALUES ('$version', '$name', datetime('now'));" 2>/dev/null || true
        fi
    fi
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}     迁移处理完成${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 显示迁移状态
echo -e "${YELLOW}迁移状态:${NC}"
sqlite3 "$DB_PATH" "SELECT version, name, applied_at FROM migrations ORDER BY version;" -header -column
