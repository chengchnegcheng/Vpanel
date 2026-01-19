#!/bin/bash

# 数据库迁移验证脚本

set -e

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 数据库配置（从环境变量或默认值）
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_NAME=${DB_NAME:-"vpanel"}
DB_USER=${DB_USER:-"vpanel"}

echo -e "${GREEN}验证数据库迁移${NC}"
echo "数据库: $DB_HOST:$DB_PORT/$DB_NAME"
echo ""

# 检查 psql 是否可用
if ! command -v psql &> /dev/null; then
    echo -e "${RED}错误: psql 未安装${NC}"
    echo "请安装 PostgreSQL 客户端"
    exit 1
fi

# 测试数据库连接
echo -e "${YELLOW}测试数据库连接...${NC}"
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1" &> /dev/null; then
    echo -e "${GREEN}✓ 数据库连接成功${NC}"
else
    echo -e "${RED}✗ 数据库连接失败${NC}"
    exit 1
fi

# 检查 proxies 表
echo ""
echo -e "${YELLOW}检查 proxies 表...${NC}"
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\d proxies" &> /dev/null; then
    echo -e "${GREEN}✓ proxies 表存在${NC}"
else
    echo -e "${RED}✗ proxies 表不存在${NC}"
    exit 1
fi

# 检查 node_id 字段
echo ""
echo -e "${YELLOW}检查 node_id 字段...${NC}"
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\d proxies" | grep -q "node_id"; then
    echo -e "${GREEN}✓ node_id 字段存在${NC}"
else
    echo -e "${RED}✗ node_id 字段不存在${NC}"
    echo ""
    echo "需要运行迁移:"
    echo "psql -U $DB_USER -d $DB_NAME -f internal/database/migrations/024_add_node_id_to_proxies.sql"
    exit 1
fi

# 检查索引
echo ""
echo -e "${YELLOW}检查索引...${NC}"
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\di" | grep -q "idx_proxies_node_id"; then
    echo -e "${GREEN}✓ idx_proxies_node_id 索引存在${NC}"
else
    echo -e "${YELLOW}⚠ idx_proxies_node_id 索引不存在${NC}"
fi

# 检查外键
echo ""
echo -e "${YELLOW}检查外键约束...${NC}"
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\d+ proxies" | grep -q "fk_proxies_node"; then
    echo -e "${GREEN}✓ fk_proxies_node 外键存在${NC}"
else
    echo -e "${YELLOW}⚠ fk_proxies_node 外键不存在${NC}"
fi

# 检查 nodes 表
echo ""
echo -e "${YELLOW}检查 nodes 表...${NC}"
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\d nodes" &> /dev/null; then
    echo -e "${GREEN}✓ nodes 表存在${NC}"
else
    echo -e "${YELLOW}⚠ nodes 表不存在（可能还未创建）${NC}"
fi

echo ""
echo -e "${GREEN}验证完成！${NC}"
echo ""
echo "数据库结构:"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "\d proxies"
