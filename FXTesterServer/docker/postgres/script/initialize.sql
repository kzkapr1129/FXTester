-- データベース作成
CREATE DATABASE fxtester_db;

-- データベース切り替え
\c fxtester_db

-- スキーマ作成
CREATE SCHEMA fxtester_schema;

-- ロールの作成
CREATE ROLE admin_user WITH LOGIN PASSWORD 'admin_user';
CREATE ROLE app_user WITH LOGIN PASSWORD 'app_user';

-- 権限追加
GRANT ALL PRIVILEGES ON SCHEMA fxtester_schema TO admin_user;

GRANT CONNECT ON DATABASE fxtester_db TO app_user;
GRANT USAGE, CREATE ON SCHEMA fxtester_schema TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA fxtester_schema TO app_user; -- TODO ALL TABLESではなく、テーブルごとに必要な権限を設定する
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA fxtester_schema TO app_user;

-- スキーマ切り替え
set search_path to fxtester_schema;

-- 記述方法思い出しのため一時的に残しておく ここから ==>

-- -- ロールの切り替え (※作成するタイプのオーナーをadmin_userにしたいので)
-- SET role admin_user;

-- -- タイプの作成
-- CREATE TYPE resource_type_t AS ENUM ('candle');
-- CREATE TYPE data_type_t AS ENUM ('candle_feature');
-- CREATE TYPE candle_feature_type_t AS ENUM ('power_bar', 'elliott', 'candlewick');
-- CREATE TYPE candle_t AS (
--     time TIMESTAMP WITH TIME ZONE
--     , high DECIMAL
--     , low DECIMAL
--     , open DECIMAL
--     , close DECIMAL
-- );
-- CREATE TYPE candle_feature_t AS (
--     resource_candles_id BIGINT
--     , candle_feature_type candle_feature_type_t
--     , candle_indexes BIGINT[]
-- );

-- -- タイプに対しての操作権限をBE用ユーザに付与
-- GRANT USAGE ON TYPE fxtester_schema.resource_type_t TO app_user;
-- GRANT USAGE ON TYPE fxtester_schema.data_type_t TO app_user;
-- GRANT USAGE ON TYPE fxtester_schema.candle_feature_type_t TO app_user;
-- GRANT USAGE ON TYPE fxtester_schema.candle_t TO app_user;

-- -- ロールの切り替え (※作成するテーブルのオーナーをapp_userにしたいので)
-- SET role app_user;

-- -- テーブル作成
-- --- ユーザ管理デーブル
-- CREATE TABLE fxtester_schema.user (
--   id BIGINT PRIMARY KEY
--   , refresh_token TEXT
--   , created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );

-- --- リソース管理テーブル (memo: リソースをアップロードできるのは管理者だけにするか？パーティションテーブルの性能的に100以上分割するのは好ましくないとのことなので・・・)
-- CREATE TABLE IF NOT EXISTS fxtester_schema.resource (
--   resource_id BIGINT PRIMARY KEY
--   , resource_name TEXT UNIQUE -- ユーザがアップロードしたリソースは全てのユーザで閲覧可能とし、重複データを減らすため
--   , resource_type resource_type_t
--   , relation_id BIGINT
--   , created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );

-- --- データ管理テーブル
-- CREATE TABLE IF NOT EXISTS fxtester_schema.data (
--     data_id BIGINT PRIMARY KEY
--     , data_type data_type_t
--     , relation_id BIGINT
--     , owner_user_id BIGINT
--     , created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );

-- --- ローソク足リソーステーブル
-- CREATE TABLE IF NOT EXISTS fxtester_schema.resource_candles (
--     resource_candles_id BIGINT
--     , time TIMESTAMP WITH TIME ZONE
--     , high DECIMAL NOT NULL CONSTRAINT high_price_check CHECK (high >= 0 AND high >= low)
--     , low DECIMAL NOT NULL CONSTRAINT low_price_check CHECK (low >= 0)
--     , open DECIMAL NOT NULL CONSTRAINT open_price_check CHECK (open >= 0)
--     , close DECIMAL NOT NULL CONSTRAINT close_price_check CHECK (close >= 0)
--     , PRIMARY KEY (resource_candles_id, time)
-- ) PARTITION BY LIST(resource_candles_id);

-- --- ローソク足特徴データテーブル
-- CREATE TABLE IF NOT EXISTS fxtester_schema.data_candle_feature (
--     data_candle_feature_id BIGINT
--     , candle_feature_type candle_feature_type_t
--     , resource_candles_id BIGINT
--     , candle_indexes BIGINT[]
--     , param jsonb
--     , PRIMARY KEY (data_candle_feature_id, candle_feature_type) -- 一つのIDに対して複数の特徴タイプを保存可能
-- );

-- -- シーケンスの作成 (0は初期値として使用するため統一的に1始まりとする)
-- CREATE SEQUENCE IF NOT EXISTS fxtester_schema.user_id                MINVALUE 1 OWNED BY fxtester_schema.user.id;
-- CREATE SEQUENCE IF NOT EXISTS fxtester_schema.resource_id            MINVALUE 1 OWNED BY fxtester_schema.resource.resource_id;
-- CREATE SEQUENCE IF NOT EXISTS fxtester_schema.resource_candles_id    MINVALUE 1 OWNED BY fxtester_schema.resource_candles.resource_candles_id;
-- CREATE SEQUENCE IF NOT EXISTS fxtester_schema.data_candle_feature_id MINVALUE 1 OWNED BY fxtester_schema.data_candle_feature.data_candle_feature_id;

-- -- ストアドプロシージャーの追加
-- --- ローソク足リソースの保存
-- CREATE OR REPLACE PROCEDURE fxtester_schema.pr_save_resource_candles(resource_name TEXT, candles candle_t[])
-- AS
-- $$
-- DECLARE
--     new_id BIGINT;
-- begin
--     /*
--     SQL実行例
--     CALL pr_save_resource_candles('test', ARRAY[
--         ROW('2024-04-12 10:11:00+00', 4, 1, 2, 3)
--         , ROW('2024-04-12 10:12:00+00', 4, 1, 2, 3)
--         , ROW('2024-04-12 10:13:00+00', 4, 1, 2, 3)
--         , ROW('2024-04-12 10:14:00+00', 4, 1, 2, 3)
--         , ROW('2024-04-12 10:15:00+00', 4, 1, 2, 3)
--     ]::candle_t[]);
--     */
--     -- 新しいIDを取得する
-- 	SELECT nextval('resource_candles_id') INTO new_id;
--     -- パーティションテーブルの作成
--     EXECUTE 'CREATE TABLE IF NOT EXISTS fxtester_schema.resource_candles_' || new_id || ' PARTITION OF fxtester_schema.resource_candles FOR VALUES IN(''' || new_id || ''');';
--     -- ローソク足データの挿入
-- 	INSERT INTO fxtester_schema.resource_candles (resource_candles_id, time, high, low, open, close)
-- 	SELECT
-- 		new_id
-- 		, (data).time
-- 		, (data).high
-- 		, (data).low
-- 		, (data).open
-- 		, (data).close
-- 	FROM unnest(candles) AS data
-- 	ON CONFLICT (resource_candles_id, time)
-- 	DO UPDATE SET high = EXCLUDED.high, low = EXCLUDED.low, open = EXCLUDED.open, close = EXCLUDED.close;

-- 	INSERT INTO fxtester_schema.resource (resource_id, resource_name, resource_type, relation_id)
-- 	VALUES (nextval('resource_id'), resource_name, 'candle', new_id);
-- END;
-- $$
-- LANGUAGE plpgsql;

-- --- ローソク足リソースの削除
-- CREATE OR REPLACE PROCEDURE fxtester_schema.pr_delete_resource_candles(resource_candles_id BIGINT)
-- AS
-- $$
-- begin
--     -- パーティションテーブルを削除する (postgresはdropはトランザクションで利用可)
--     EXECUTE 'DROP TABLE fxtester_schema.resource_candles_' || resource_candles_id;

--     -- resourceテーブルも削除
--     DELETE FROM fxtester_schema.resource WHERE relation_id = resource_candles_id;
-- END;
-- $$
-- LANGUAGE plpgsql;

-- <== ここまで 記述方法思い出しのため一時的に残しておく

-- ロールの切り替え (※作成するテーブルのオーナーをapp_userにしたいので)
SET role app_user;

CREATE TABLE IF NOT EXISTS fxtester_schema.user (
    id BIGINT PRIMARY KEY
    , email varchar UNIQUE NOT NULL
    , access_token varchar
    , refresh_token varchar
);

-- シーケンスの作成 (0は初期値として使用するため統一的に1始まりとする)
CREATE SEQUENCE IF NOT EXISTS fxtester_schema.user_id MINVALUE 1 OWNED BY fxtester_schema.user.id;

-- 関数の追加
/**
 * 関数名: create_user
 * 機能: 新規ユーザ追加を行い、追加したユーザのIDを返却する
 * 利用例: select fxtester_schema.create_user('test@gmail.com');
 */
create or replace function fxtester_schema.create_user(in p_email varchar)
returns bigint
AS $$
DECLARE
    new_id bigint;
BEGIN
    INSERT INTO fxtester_schema.user (id, email)
    VALUES (nextval('fxtester_schema.user_id'), p_email) returning id INTO new_id;
    return new_id;
END;
$$ language plpgsql;

/**
 * 関数名: select_user_with_email
 * 機能: 指定したemailと一致するユーザ情報を返却します
 * 利用例: SELECT * FROM fxtester_schema.select_user_with_email('test@fxtester.com');
 */
CREATE OR REPLACE FUNCTION fxtester_schema.select_user_with_email(p_email varchar)
RETURNS TABLE(
    id bigint,
    email varchar,
    access_token varchar,
    refresh_token varchar
) AS $$
BEGIN
    RETURN QUERY
    SELECT *
    FROM fxtester_schema.user u
    WHERE u.email = p_email;
END;
$$ LANGUAGE plpgsql;

/**
 * ストアドプロシージャー名: update_token
 * 機能: 指定ユーザとアクセストークン、リフレッシュトークンを関連付けします
 * 利用例: call call fxtester_schema.update_token(1, 'access_token_xxx', 'refresh_token_yyy')
 */
create or replace procedure fxtester_schema.update_token(p_user_id bigint, p_access_token varchar, p_refresh_token varchar)
AS $$
BEGIN
    UPDATE fxtester_schema.user u SET access_token=p_access_token, refresh_token=p_refresh_token WHERE u.id = p_user_id;
END;
$$ language plpgsql;