-- 1. 把现有所有表的 owner 改成 meme_server
DO $$
DECLARE r RECORD;
BEGIN
    FOR r IN SELECT tablename FROM pg_tables WHERE schemaname = 'public'
        LOOP
      EXECUTE 'ALTER TABLE public.' || quote_ident(r.tablename) || ' OWNER TO meme_server';
    END LOOP;
END $$;

-- 2. 让 meme_server 可以在 public schema 下建表（分区表需要）
GRANT CREATE ON SCHEMA public TO meme_server;

-- 3. 设置默认权限：以后 postgres 用户建的表自动授权给 meme_server
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO meme_server;