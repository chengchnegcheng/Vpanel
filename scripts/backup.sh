#!/bin/bash

# 备份脚本
# 用于备份数据库和配置文件

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 配置
BACKUP_DIR=${BACKUP_DIR:-"backups"}
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-"vpanel"}
DB_USER=${DB_USER:-"vpanel"}
DB_PASSWORD=${DB_PASSWORD:-""}
KEEP_DAYS=${KEEP_DAYS:-7}

# 时间戳
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

show_help() {
    echo "V Panel 备份脚本"
    echo ""
    echo "用法:"
    echo "  $0 database           备份数据库"
    echo "  $0 config             备份配置文件"
    echo "  $0 all                备份所有内容"
    echo "  $0 restore <file>     恢复备份"
    echo "  $0 clean              清理旧备份"
    echo ""
    echo "环境变量:"
    echo "  BACKUP_DIR            备份目录 (默认: backups)"
    echo "  DB_HOST               数据库主机 (默认: localhost)"
    echo "  DB_PORT               数据库端口 (默认: 5432)"
    echo "  DB_NAME               数据库名称 (默认: vpanel)"
    echo "  DB_USER               数据库用户 (默认: vpanel)"
    echo "  DB_PASSWORD           数据库密码"
    echo "  KEEP_DAYS             保留天数 (默认: 7)"
    echo ""
}

backup_database() {
    echo -e "${YELLOW}备份数据库...${NC}"
    
    # 创建备份目录
    mkdir -p "$BACKUP_DIR/database"
    
    # 备份文件名
    backup_file="$BACKUP_DIR/database/vpanel_db_${TIMESTAMP}.sql"
    
    # 设置密码环境变量
    if [ -n "$DB_PASSWORD" ]; then
        export PGPASSWORD="$DB_PASSWORD"
    fi
    
    # 执行备份
    echo "备份到: $backup_file"
    if pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -F p -f "$backup_file"; then
        echo -e "${GREEN}✓ 数据库备份成功${NC}"
        
        # 压缩备份
        gzip "$backup_file"
        echo -e "${GREEN}✓ 备份已压缩: ${backup_file}.gz${NC}"
        
        # 显示文件大小
        size=$(ls -lh "${backup_file}.gz" | awk '{print $5}')
        echo "备份大小: $size"
    else
        echo -e "${RED}✗ 数据库备份失败${NC}"
        return 1
    fi
    
    # 清除密码环境变量
    unset PGPASSWORD
}

backup_config() {
    echo -e "${YELLOW}备份配置文件...${NC}"
    
    # 创建备份目录
    mkdir -p "$BACKUP_DIR/config"
    
    # 备份文件名
    backup_file="$BACKUP_DIR/config/vpanel_config_${TIMESTAMP}.tar.gz"
    
    # 要备份的文件和目录
    files_to_backup=(
        "configs"
        ".env"
    )
    
    # 检查文件是否存在
    existing_files=()
    for file in "${files_to_backup[@]}"; do
        if [ -e "$file" ]; then
            existing_files+=("$file")
        fi
    done
    
    if [ ${#existing_files[@]} -eq 0 ]; then
        echo -e "${YELLOW}⚠ 没有找到配置文件${NC}"
        return 0
    fi
    
    # 创建备份
    echo "备份到: $backup_file"
    if tar -czf "$backup_file" "${existing_files[@]}" 2>/dev/null; then
        echo -e "${GREEN}✓ 配置文件备份成功${NC}"
        
        # 显示文件大小
        size=$(ls -lh "$backup_file" | awk '{print $5}')
        echo "备份大小: $size"
        
        # 列出备份内容
        echo "备份内容:"
        tar -tzf "$backup_file" | sed 's/^/  /'
    else
        echo -e "${RED}✗ 配置文件备份失败${NC}"
        return 1
    fi
}

backup_agent_config() {
    echo -e "${YELLOW}备份 Agent 配置...${NC}"
    
    # 创建备份目录
    mkdir -p "$BACKUP_DIR/agent"
    
    # 备份文件名
    backup_file="$BACKUP_DIR/agent/agent_config_${TIMESTAMP}.tar.gz"
    
    # 要备份的文件
    files_to_backup=(
        "/etc/vpanel/agent.yaml"
        "/etc/xray/config.json"
    )
    
    # 检查文件是否存在
    existing_files=()
    for file in "${files_to_backup[@]}"; do
        if [ -f "$file" ]; then
            existing_files+=("$file")
        fi
    done
    
    if [ ${#existing_files[@]} -eq 0 ]; then
        echo -e "${YELLOW}⚠ 没有找到 Agent 配置文件${NC}"
        return 0
    fi
    
    # 创建备份
    echo "备份到: $backup_file"
    if sudo tar -czf "$backup_file" "${existing_files[@]}" 2>/dev/null; then
        echo -e "${GREEN}✓ Agent 配置备份成功${NC}"
        
        # 显示文件大小
        size=$(ls -lh "$backup_file" | awk '{print $5}')
        echo "备份大小: $size"
    else
        echo -e "${RED}✗ Agent 配置备份失败${NC}"
        return 1
    fi
}

restore_database() {
    local backup_file=$1
    
    if [ ! -f "$backup_file" ]; then
        echo -e "${RED}✗ 备份文件不存在: $backup_file${NC}"
        return 1
    fi
    
    echo -e "${YELLOW}恢复数据库...${NC}"
    echo -e "${RED}警告: 这将覆盖当前数据库！${NC}"
    read -p "确认恢复? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        echo "取消恢复"
        return 0
    fi
    
    # 解压备份文件
    temp_file="/tmp/vpanel_restore_${TIMESTAMP}.sql"
    if [[ "$backup_file" == *.gz ]]; then
        gunzip -c "$backup_file" > "$temp_file"
    else
        cp "$backup_file" "$temp_file"
    fi
    
    # 设置密码环境变量
    if [ -n "$DB_PASSWORD" ]; then
        export PGPASSWORD="$DB_PASSWORD"
    fi
    
    # 恢复数据库
    echo "从备份恢复: $backup_file"
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$temp_file"; then
        echo -e "${GREEN}✓ 数据库恢复成功${NC}"
    else
        echo -e "${RED}✗ 数据库恢复失败${NC}"
        rm -f "$temp_file"
        return 1
    fi
    
    # 清理
    rm -f "$temp_file"
    unset PGPASSWORD
}

restore_config() {
    local backup_file=$1
    
    if [ ! -f "$backup_file" ]; then
        echo -e "${RED}✗ 备份文件不存在: $backup_file${NC}"
        return 1
    fi
    
    echo -e "${YELLOW}恢复配置文件...${NC}"
    echo -e "${RED}警告: 这将覆盖当前配置！${NC}"
    read -p "确认恢复? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        echo "取消恢复"
        return 0
    fi
    
    # 恢复配置
    echo "从备份恢复: $backup_file"
    if tar -xzf "$backup_file"; then
        echo -e "${GREEN}✓ 配置文件恢复成功${NC}"
    else
        echo -e "${RED}✗ 配置文件恢复失败${NC}"
        return 1
    fi
}

clean_old_backups() {
    echo -e "${YELLOW}清理旧备份...${NC}"
    echo "保留最近 $KEEP_DAYS 天的备份"
    
    if [ ! -d "$BACKUP_DIR" ]; then
        echo "备份目录不存在"
        return 0
    fi
    
    # 查找并删除旧备份
    deleted_count=0
    while IFS= read -r -d '' file; do
        echo "删除: $file"
        rm -f "$file"
        ((deleted_count++))
    done < <(find "$BACKUP_DIR" -type f -mtime +$KEEP_DAYS -print0)
    
    if [ $deleted_count -eq 0 ]; then
        echo -e "${GREEN}✓ 没有需要清理的旧备份${NC}"
    else
        echo -e "${GREEN}✓ 已删除 $deleted_count 个旧备份${NC}"
    fi
}

list_backups() {
    echo -e "${YELLOW}可用备份:${NC}"
    echo ""
    
    if [ ! -d "$BACKUP_DIR" ]; then
        echo "没有找到备份"
        return 0
    fi
    
    echo "数据库备份:"
    if [ -d "$BACKUP_DIR/database" ]; then
        ls -lh "$BACKUP_DIR/database" | tail -n +2 || echo "  无"
    else
        echo "  无"
    fi
    
    echo ""
    echo "配置备份:"
    if [ -d "$BACKUP_DIR/config" ]; then
        ls -lh "$BACKUP_DIR/config" | tail -n +2 || echo "  无"
    else
        echo "  无"
    fi
    
    echo ""
    echo "Agent 配置备份:"
    if [ -d "$BACKUP_DIR/agent" ]; then
        ls -lh "$BACKUP_DIR/agent" | tail -n +2 || echo "  无"
    else
        echo "  无"
    fi
}

# 主逻辑
case "$1" in
    database)
        backup_database
        ;;
    config)
        backup_config
        ;;
    agent)
        backup_agent_config
        ;;
    all)
        backup_database
        echo ""
        backup_config
        echo ""
        backup_agent_config
        ;;
    restore)
        if [ -z "$2" ]; then
            echo -e "${RED}错误: 需要指定备份文件${NC}"
            echo "用法: $0 restore <backup-file>"
            exit 1
        fi
        
        # 根据文件类型恢复
        if [[ "$2" == *"_db_"* ]]; then
            restore_database "$2"
        elif [[ "$2" == *"_config_"* ]]; then
            restore_config "$2"
        else
            echo -e "${RED}错误: 无法识别备份文件类型${NC}"
            exit 1
        fi
        ;;
    clean)
        clean_old_backups
        ;;
    list)
        list_backups
        ;;
    *)
        show_help
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}完成！${NC}"
