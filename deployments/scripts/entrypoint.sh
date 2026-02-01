#!/bin/sh
set -e

# V Panel Docker Entrypoint Script

echo "Starting V Panel..."

# 生产环境安全检查
if [ "${V_SERVER_MODE}" = "release" ]; then
    echo "Production mode detected, performing security checks..."
    
    # 检查 JWT Secret
    if [ -z "${V_JWT_SECRET}" ] || \
       [ "${V_JWT_SECRET}" = "CHANGE_ME_OR_AUTO_GENERATE_ON_FIRST_START" ] || \
       [ "${V_JWT_SECRET}" = "CHANGE_ME_OR_SYSTEM_WILL_REFUSE_TO_START" ] || \
       [ "${V_JWT_SECRET}" = "your-secure-jwt-secret-change-me" ] || \
       [ "${V_JWT_SECRET}" = "change-me-in-production" ]; then
        echo "ERROR: JWT_SECRET is not configured or using default value!"
        echo "Please set a secure JWT_SECRET in your .env file"
        echo "Generate one with: openssl rand -base64 32"
        exit 1
    fi
    
    # 检查 JWT Secret 长度
    JWT_LEN=$(echo -n "${V_JWT_SECRET}" | wc -c | tr -d ' ')
    if [ "${JWT_LEN}" -lt 32 ]; then
        echo "ERROR: JWT_SECRET is too short (${JWT_LEN} chars, minimum 32 required)"
        exit 1
    fi
    
    # 检查管理员密码
    if [ -z "${V_ADMIN_PASS}" ] || \
       [ "${V_ADMIN_PASS}" = "CHANGE_ME_OR_AUTO_GENERATE_ON_FIRST_START" ] || \
       [ "${V_ADMIN_PASS}" = "CHANGE_ME_OR_SYSTEM_WILL_REFUSE_TO_START" ] || \
       [ "${V_ADMIN_PASS}" = "admin123" ] || \
       [ "${V_ADMIN_PASS}" = "your-secure-admin-password" ]; then
        echo "ERROR: Admin password is not configured or using default value!"
        echo "Please set a secure password in your .env file"
        exit 1
    fi
    
    # 检查密码长度
    PASS_LEN=$(echo -n "${V_ADMIN_PASS}" | wc -c | tr -d ' ')
    if [ "${PASS_LEN}" -lt 12 ]; then
        echo "ERROR: Admin password is too short (${PASS_LEN} chars, minimum 12 required)"
        exit 1
    fi
    
    echo "✓ Security checks passed"
fi

# Create config from example if not exists
if [ ! -f /app/configs/config.yaml ]; then
    echo "Creating default configuration..."
    cp /app/configs/config.yaml.example /app/configs/config.yaml
fi

# Ensure data directory exists and has correct permissions
mkdir -p /app/data /app/logs

# Initialize database if needed
if [ ! -f /app/data/v.db ]; then
    echo "Initializing database..."
    touch /app/data/v.db
fi

# Print startup information
echo "Configuration:"
echo "  Server Host: ${V_SERVER_HOST:-0.0.0.0}"
echo "  Server Port: ${V_SERVER_PORT:-8080}"
echo "  Server Mode: ${V_SERVER_MODE:-release}"
echo "  Log Level: ${V_LOG_LEVEL:-info}"
echo "  Database: ${V_DB_PATH:-/app/data/v.db}"

# Execute the main command
exec "$@"
