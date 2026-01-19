-- 添加测试流量数据
-- 用于测试用户门户统计功能

-- 为用户ID 1 添加最近30天的测试流量数据
DO $$
DECLARE
    test_user_id INT := 1;
    test_proxy_id INT := 1;
    days_back INT;
    current_date TIMESTAMP;
    upload_bytes BIGINT;
    download_bytes BIGINT;
BEGIN
    -- 检查用户是否存在
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = test_user_id) THEN
        RAISE NOTICE '用户ID % 不存在，跳过添加测试数据', test_user_id;
        RETURN;
    END IF;

    -- 检查是否已有流量数据
    IF EXISTS (SELECT 1 FROM traffic WHERE user_id = test_user_id LIMIT 1) THEN
        RAISE NOTICE '用户ID % 已有流量数据', test_user_id;
        RETURN;
    END IF;

    RAISE NOTICE '为用户ID % 添加测试流量数据...', test_user_id;

    -- 添加最近30天的数据
    FOR days_back IN 0..29 LOOP
        current_date := NOW() - (days_back || ' days')::INTERVAL;
        
        -- 生成随机流量数据 (100MB - 5GB)
        upload_bytes := (100 + RANDOM() * 4900) * 1024 * 1024;
        download_bytes := (500 + RANDOM() * 4500) * 1024 * 1024;
        
        INSERT INTO traffic (user_id, proxy_id, upload, download, recorded_at, created_at)
        VALUES (
            test_user_id,
            test_proxy_id,
            upload_bytes,
            download_bytes,
            current_date,
            current_date
        );
    END LOOP;

    RAISE NOTICE '成功添加 30 条测试流量记录';
END $$;

-- 查看添加的数据
SELECT 
    COUNT(*) as total_records,
    SUM(upload) as total_upload,
    SUM(download) as total_download,
    MIN(recorded_at) as earliest_date,
    MAX(recorded_at) as latest_date
FROM traffic
WHERE user_id = 1;
