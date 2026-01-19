-- 插入测试流量数据
-- 用于测试用户门户统计功能

-- 假设用户ID为1，代理ID为1
-- 插入最近30天的流量数据

DO $$
DECLARE
    test_user_id BIGINT := 1;
    test_proxy_id BIGINT := 1;
    days_back INT;
    current_date TIMESTAMP;
    upload_bytes BIGINT;
    download_bytes BIGINT;
BEGIN
    -- 检查用户是否存在
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = test_user_id) THEN
        RAISE NOTICE '用户ID % 不存在，请先创建用户', test_user_id;
        RETURN;
    END IF;

    -- 检查代理是否存在
    IF NOT EXISTS (SELECT 1 FROM proxies WHERE id = test_proxy_id) THEN
        RAISE NOTICE '代理ID % 不存在，请先创建代理', test_proxy_id;
        RETURN;
    END IF;

    -- 删除现有的测试数据（可选）
    -- DELETE FROM traffic WHERE user_id = test_user_id;

    -- 插入最近30天的数据
    FOR days_back IN 0..29 LOOP
        current_date := NOW() - (days_back || ' days')::INTERVAL;
        
        -- 生成随机流量数据（模拟真实使用）
        upload_bytes := (RANDOM() * 1000000000)::BIGINT;  -- 0-1GB
        download_bytes := (RANDOM() * 5000000000)::BIGINT;  -- 0-5GB
        
        -- 插入流量记录
        INSERT INTO traffic (user_id, proxy_id, upload, download, recorded_at, created_at)
        VALUES (
            test_user_id,
            test_proxy_id,
            upload_bytes,
            download_bytes,
            current_date,
            NOW()
        );
        
        RAISE NOTICE '插入日期: %, 上传: % MB, 下载: % MB', 
            current_date::DATE, 
            (upload_bytes / 1048576)::INT,
            (download_bytes / 1048576)::INT;
    END LOOP;

    RAISE NOTICE '成功插入 30 天的测试流量数据';
END $$;

-- 查询验证
SELECT 
    DATE(recorded_at) as date,
    COUNT(*) as records,
    SUM(upload) / 1048576 as upload_mb,
    SUM(download) / 1048576 as download_mb,
    (SUM(upload) + SUM(download)) / 1048576 as total_mb
FROM traffic
WHERE user_id = 1
GROUP BY DATE(recorded_at)
ORDER BY date DESC
LIMIT 10;
