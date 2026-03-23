--
-- PostgreSQL database dump
--

-- Dumped from database version 16.2 (Debian 16.2-1.pgdg120+2)
-- Dumped by pg_dump version 16.2 (Debian 16.2-1.pgdg120+2)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: ulid; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS ulid WITH SCHEMA public;


--
-- Name: EXTENSION ulid; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION ulid IS 'ulid type and methods';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: audit_comment_template; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.audit_comment_template (
    id bigint NOT NULL,
    comment character varying(255) NOT NULL,
    auditor character varying(32) DEFAULT 'admin'::character varying NOT NULL,
    create_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP,
    update_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.audit_comment_template OWNER TO meme_server;

--
-- Name: audit_comment_template_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.audit_comment_template_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.audit_comment_template_id_seq OWNER TO meme_server;

--
-- Name: audit_comment_template_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.audit_comment_template_id_seq OWNED BY public.audit_comment_template.id;


--
-- Name: audit_comment_template_multi_lang; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.audit_comment_template_multi_lang (
    id bigint NOT NULL,
    comment json NOT NULL,
    auditor character varying(32) NOT NULL,
    create_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP,
    update_at timestamp(6) with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.audit_comment_template_multi_lang OWNER TO meme_server;

--
-- Name: audit_comment_template_multi_lang_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.audit_comment_template_multi_lang_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.audit_comment_template_multi_lang_id_seq OWNER TO meme_server;

--
-- Name: audit_comment_template_multi_lang_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.audit_comment_template_multi_lang_id_seq OWNED BY public.audit_comment_template_multi_lang.id;


--
-- Name: game_score; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.game_score (
    user_id bigint NOT NULL,
    total_score integer NOT NULL,
    frozen_score integer NOT NULL,
    daily_obtain_limit integer DEFAULT 1000 NOT NULL,
    daily_obtained integer DEFAULT 0 NOT NULL,
    last_modified timestamp with time zone DEFAULT now() NOT NULL,
    version integer DEFAULT 0 NOT NULL,
    CONSTRAINT user_score_frozen_score_check CHECK ((frozen_score >= 0)),
    CONSTRAINT user_score_total_score_check CHECK ((total_score >= 0))
);


ALTER TABLE public.game_score OWNER TO meme_server;

--
-- Name: game_score_operation_log; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.game_score_operation_log (
    log_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    delta_score integer NOT NULL,
    pre_total integer NOT NULL,
    post_total integer NOT NULL,
    pre_frozen integer,
    post_frozen integer,
    op_type character varying(20) NOT NULL,
    transaction_id character varying(36) NOT NULL,
    source_sys character varying(48) NOT NULL,
    business_name character varying(32) NOT NULL,
    reason text,
    CONSTRAINT game_score_operation_log_op_type_check CHECK (((op_type)::text = ANY (ARRAY[('GRANT'::character varying)::text, ('CONSUME'::character varying)::text, ('CONSUME_FREEZE'::character varying)::text, ('FREEZE'::character varying)::text, ('UNFREEZE'::character varying)::text])))
);


ALTER TABLE public.game_score_operation_log OWNER TO meme_server;

--
-- Name: TABLE game_score_operation_log; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON TABLE public.game_score_operation_log IS '游戏积分操作记录表';


--
-- Name: COLUMN game_score_operation_log.log_id; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.log_id IS '自增主键';


--
-- Name: COLUMN game_score_operation_log.user_id; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.user_id IS '关联用户ID';


--
-- Name: COLUMN game_score_operation_log.created_at; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.created_at IS '记录创建时间';


--
-- Name: COLUMN game_score_operation_log.delta_score; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.delta_score IS '积分变动数';


--
-- Name: COLUMN game_score_operation_log.pre_total; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.pre_total IS '操作前积分总数';


--
-- Name: COLUMN game_score_operation_log.post_total; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.post_total IS '操作后积分总数';


--
-- Name: COLUMN game_score_operation_log.pre_frozen; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.pre_frozen IS '操作前冻结积分';


--
-- Name: COLUMN game_score_operation_log.post_frozen; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.post_frozen IS '操作后冻结积分';


--
-- Name: COLUMN game_score_operation_log.op_type; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.op_type IS '操作类型';


--
-- Name: COLUMN game_score_operation_log.transaction_id; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.transaction_id IS '交易id';


--
-- Name: COLUMN game_score_operation_log.source_sys; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.source_sys IS '来源系统';


--
-- Name: COLUMN game_score_operation_log.business_name; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.game_score_operation_log.business_name IS '业务名称';


--
-- Name: game_score_operation_log_log_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.game_score_operation_log_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.game_score_operation_log_log_id_seq OWNER TO meme_server;

--
-- Name: game_score_operation_log_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.game_score_operation_log_log_id_seq OWNED BY public.game_score_operation_log.log_id;


--
-- Name: invitation_relation; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.invitation_relation (
    id bigint NOT NULL,
    inviter bigint NOT NULL,
    invitee bigint NOT NULL,
    invite_code character varying(40) NOT NULL,
    create_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.invitation_relation OWNER TO meme_server;

--
-- Name: invitation_relation_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.invitation_relation_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.invitation_relation_id_seq OWNER TO meme_server;

--
-- Name: invitation_relation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.invitation_relation_id_seq OWNED BY public.invitation_relation.id;


--
-- Name: login_log; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.login_log (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    login_time timestamp with time zone NOT NULL,
    create_time timestamp with time zone NOT NULL,
    update_time timestamp with time zone NOT NULL,
    ip_address character varying(45),
    provider character varying(32),
    provider_user_id character varying(32),
    provider_claims text,
    user_agent text NOT NULL,
    failed_reason character varying(32),
    failed_detail text
);


ALTER TABLE public.login_log OWNER TO meme_server;

--
-- Name: TABLE login_log; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON TABLE public.login_log IS '用户登录审计日志表';


--
-- Name: COLUMN login_log.id; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.id IS '自增主键';


--
-- Name: COLUMN login_log.user_id; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.user_id IS '关联用户ID';


--
-- Name: COLUMN login_log.login_time; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.login_time IS '登录（成功）时间';


--
-- Name: COLUMN login_log.create_time; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.create_time IS '记录创建时间';


--
-- Name: COLUMN login_log.update_time; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.update_time IS '记录最近更新时间';


--
-- Name: COLUMN login_log.ip_address; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.ip_address IS '用户IP地址';


--
-- Name: COLUMN login_log.provider; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.provider IS '社交媒体类型';


--
-- Name: COLUMN login_log.provider_user_id; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.provider_user_id IS '社交媒体认证完毕后返回的用户ID';


--
-- Name: COLUMN login_log.provider_claims; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.provider_claims IS '社交媒体认证完毕后返回的用户信息';


--
-- Name: COLUMN login_log.user_agent; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.user_agent IS '客户端User-Agent';


--
-- Name: COLUMN login_log.failed_reason; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.failed_reason IS '登录失败原因';


--
-- Name: COLUMN login_log.failed_detail; Type: COMMENT; Schema: public; Owner: meme_server
--

COMMENT ON COLUMN public.login_log.failed_detail IS '登录失败详情况';


--
-- Name: login_log_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.login_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.login_log_id_seq OWNER TO meme_server;

--
-- Name: login_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.login_log_id_seq OWNED BY public.login_log.id;


--
-- Name: mini_game_item; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.mini_game_item (
    id bigint NOT NULL,
    item_id character varying(32) NOT NULL,
    unit_price integer NOT NULL,
    create_at timestamp with time zone DEFAULT now(),
    update_at timestamp with time zone NOT NULL
);


ALTER TABLE public.mini_game_item OWNER TO meme_server;

--
-- Name: mini_game_item_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.mini_game_item_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.mini_game_item_id_seq OWNER TO meme_server;

--
-- Name: mini_game_item_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.mini_game_item_id_seq OWNED BY public.mini_game_item.id;


--
-- Name: oshit_task_config; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.oshit_task_config (
    id bigint NOT NULL,
    create_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp with time zone,
    task_name character varying(60) NOT NULL,
    task_type character varying(16) NOT NULL,
    task_desc character varying(128),
    CONSTRAINT oshit_task_config_task_type_check CHECK (((task_type)::text = ANY ((ARRAY['SHARE'::character varying, 'INVITE_REGISTER'::character varying, 'POSTS_UPLOAD'::character varying, 'CHECKIN'::character varying, 'WALLET_BINDING'::character varying])::text[])))
);


ALTER TABLE public.oshit_task_config OWNER TO meme_server;

--
-- Name: oshit_task_config_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.oshit_task_config_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.oshit_task_config_id_seq OWNER TO meme_server;

--
-- Name: oshit_task_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.oshit_task_config_id_seq OWNED BY public.oshit_task_config.id;


--
-- Name: score_ledger; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.score_ledger (
    ledger_id bigint NOT NULL,
    create_ts timestamp with time zone DEFAULT now() NOT NULL,
    admin_id bigint NOT NULL,
    admin_name text NOT NULL,
    user_id bigint NOT NULL,
    user_name text NOT NULL,
    delta_score integer NOT NULL,
    reason_type text NOT NULL,
    reason_note text NOT NULL,
    "precision" integer DEFAULT 3 NOT NULL,
    CONSTRAINT score_ledger_reason_type_check CHECK ((reason_type = ANY (ARRAY['reward'::text, 'deduction'::text])))
);


ALTER TABLE public.score_ledger OWNER TO meme_server;

--
-- Name: score_ledger_ledger_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.score_ledger_ledger_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.score_ledger_ledger_id_seq OWNER TO meme_server;

--
-- Name: score_ledger_ledger_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.score_ledger_ledger_id_seq OWNED BY public.score_ledger.ledger_id;


--
-- Name: score_operation_log; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.score_operation_log (
    log_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    delta_score integer NOT NULL,
    pre_total integer NOT NULL,
    post_total integer NOT NULL,
    pre_frozen integer,
    post_frozen integer,
    op_type character varying(20) NOT NULL,
    transaction_id character varying(36) NOT NULL,
    source_sys character varying(48) NOT NULL,
    business_name character varying(32) NOT NULL,
    reason text,
    CONSTRAINT score_operation_log_op_type_check CHECK (((op_type)::text = ANY ((ARRAY['GRANT'::character varying, 'CONSUME'::character varying, 'CONSUME_FREEZE'::character varying, 'FREEZE'::character varying, 'UNFREEZE'::character varying])::text[])))
);


ALTER TABLE public.score_operation_log OWNER TO meme_server;

--
-- Name: score_operation_log_log_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.score_operation_log_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.score_operation_log_log_id_seq OWNER TO meme_server;

--
-- Name: score_operation_log_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.score_operation_log_log_id_seq OWNED BY public.score_operation_log.log_id;


--
-- Name: social_media_user_info; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.social_media_user_info (
    user_id bigint NOT NULL,
    provider_user_id character varying(128) NOT NULL,
    provider character varying(32) NOT NULL,
    username character varying(128),
    email character varying(128),
    display_name character varying(128),
    profile_picture_url text,
    access_token text,
    refresh_token text,
    token_expiry timestamp with time zone,
    create_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    update_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.social_media_user_info OWNER TO meme_server;

--
-- Name: t_activity_exchange_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_activity_exchange_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RankPlatform" character varying(24) NOT NULL,
    "Handle" character varying(124) NOT NULL,
    "TxId" character varying(124) NOT NULL,
    "ScoreAmount" numeric(20,2) NOT NULL,
    "TokenAmount" numeric(78,0) NOT NULL,
    "ExchangeRate" integer DEFAULT 100 NOT NULL,
    "State" integer NOT NULL,
    "RefBlockHash" character varying(45) NOT NULL,
    "LastValidBlockHeight" bigint NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_activity_exchange_record OWNER TO postgres;

--
-- Name: t_activity_exchange_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_activity_exchange_rule (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RankPlatform" character varying(24) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "Decimals" integer NOT NULL,
    "ExchangeRate" integer DEFAULT 100 NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_activity_exchange_rule OWNER TO postgres;

--
-- Name: t_activity_rank; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_activity_rank (
    id integer NOT NULL,
    name character varying(255) DEFAULT ''::character varying NOT NULL,
    handle character varying(124) DEFAULT ''::character varying NOT NULL,
    user_code character varying(32) DEFAULT ''::character varying NOT NULL,
    rank_level character varying(32) DEFAULT ''::character varying NOT NULL,
    rank_score numeric(20,2) DEFAULT 0 NOT NULL,
    used_score numeric(20,2) DEFAULT 0 NOT NULL,
    rank_platform character varying(24) DEFAULT ''::character varying NOT NULL,
    profile text,
    create_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    update_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    creator character varying(15) DEFAULT ''::character varying NOT NULL,
    updater character varying(15) DEFAULT ''::character varying NOT NULL,
    deleted smallint DEFAULT 1 NOT NULL,
    available_score numeric(20,2) DEFAULT 0 NOT NULL,
    locked_score numeric(20,2) DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_activity_rank OWNER TO postgres;

--
-- Name: COLUMN t_activity_rank.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.id IS '主键';


--
-- Name: COLUMN t_activity_rank.name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.name IS '账户名称，例如 "0://bank.BigD33"';


--
-- Name: COLUMN t_activity_rank.handle; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.handle IS '账户@，例如 "@BigD_Energy33"';


--
-- Name: COLUMN t_activity_rank.user_code; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.user_code IS '用户编码';


--
-- Name: COLUMN t_activity_rank.rank_level; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.rank_level IS 'RANK等级';


--
-- Name: COLUMN t_activity_rank.rank_score; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.rank_score IS 'RANK分数';


--
-- Name: COLUMN t_activity_rank.rank_platform; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.rank_platform IS '平台';


--
-- Name: COLUMN t_activity_rank.profile; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.profile IS '用户头像';


--
-- Name: COLUMN t_activity_rank.create_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.create_time IS '当前记录的创建时间';


--
-- Name: COLUMN t_activity_rank.update_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.update_time IS '当前记录的更新时间';


--
-- Name: COLUMN t_activity_rank.creator; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.creator IS '当前记录的创建人';


--
-- Name: COLUMN t_activity_rank.updater; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.updater IS '当前记录的更新人';


--
-- Name: COLUMN t_activity_rank.deleted; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.deleted IS '删除标志（0: 未删除, 1: 已删除）';


--
-- Name: COLUMN t_activity_rank.available_score; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.available_score IS '可用RANK分数';


--
-- Name: COLUMN t_activity_rank.locked_score; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_activity_rank.locked_score IS '锁定RANK分数';


--
-- Name: t_activity_rank_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.t_activity_rank_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.t_activity_rank_id_seq OWNER TO postgres;

--
-- Name: t_activity_rank_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.t_activity_rank_id_seq OWNED BY public.t_activity_rank.id;


--
-- Name: t_aws_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_aws_config (
    "AccessKeyId" character varying(128) NOT NULL,
    "SecretAccessKey" character varying(128) NOT NULL,
    "Region" character varying(32) NOT NULL
);


ALTER TABLE public.t_aws_config OWNER TO postgres;

--
-- Name: t_campaign_key_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_campaign_key_config (
    "Chain" character varying(64) NOT NULL,
    "EncryptedKey" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_campaign_key_config OWNER TO postgres;

--
-- Name: t_campaign_rpc_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_campaign_rpc_config (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RpcUrl" character varying(1024) NOT NULL,
    "WssUrl" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_campaign_rpc_config OWNER TO postgres;

--
-- Name: t_campaign_scan_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_campaign_scan_info (
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "UntilTxId" character varying(128) NOT NULL,
    "BeforeTxId" character varying(128),
    "Slot" numeric(78,0)
);


ALTER TABLE public.t_campaign_scan_info OWNER TO postgres;

--
-- Name: t_campaign_tx; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_campaign_tx (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "ServiceType" integer NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_campaign_tx OWNER TO postgres;

--
-- Name: t_chain_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_chain_config (
    "Chain" character varying(64) NOT NULL,
    "RpcUrl" character varying(1024) NOT NULL,
    "WssUrl" character varying(1024) NOT NULL,
    "Decimal" integer NOT NULL,
    "Symbol" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_chain_config OWNER TO postgres;

--
-- Name: t_daily_claim_stats; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_daily_claim_stats (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "TakeShitDate" date NOT NULL,
    "TakeShitCount" integer DEFAULT 0 NOT NULL,
    "NeedLottery" boolean DEFAULT false NOT NULL,
    "LastTakeTime" timestamp without time zone,
    "TotalLottery" numeric(78,0) DEFAULT 0 NOT NULL,
    "TotalTake" numeric(78,0) DEFAULT 0 NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "LotteryCount" integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_daily_claim_stats OWNER TO postgres;

--
-- Name: t_debug_white_list; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_debug_white_list (
    "Id" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_debug_white_list OWNER TO postgres;

--
-- Name: t_ecommerce_order; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_ecommerce_order (
    "OrderId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "CustomerName" character varying(128) NOT NULL,
    "CustomerEmail" character varying(128) NOT NULL,
    "OrderStatus" integer NOT NULL,
    "PaymentStatus" integer NOT NULL,
    "PaymentMethod" character varying(64) NOT NULL,
    "ProductId" character varying(64) NOT NULL,
    "ProductName" character varying(128) NOT NULL,
    "ProductPrice" numeric(10,2) NOT NULL,
    "ProductQuantity" integer NOT NULL,
    "OrderTotal" numeric(10,2) NOT NULL,
    "TrackingNumber" character varying(64),
    "ShippingCarrier" character varying(64),
    "StoreName" character varying(128) NOT NULL,
    "StoreContactInfo" character varying(128) NOT NULL,
    "OrderDate" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_ecommerce_order OWNER TO postgres;

--
-- Name: t_facebook_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_facebook_info (
    id integer NOT NULL,
    profile_image text,
    profile_id character varying(24) NOT NULL,
    profile_url character varying(255) NOT NULL,
    username character varying(64) DEFAULT ''::character varying NOT NULL,
    push_time timestamp(3) without time zone NOT NULL,
    post_link character varying(255) DEFAULT ''::character varying NOT NULL,
    post_id character varying(24) DEFAULT ''::character varying NOT NULL,
    tags text,
    content text NOT NULL,
    comments integer DEFAULT 0 NOT NULL,
    retweets integer DEFAULT 0 NOT NULL,
    likes integer DEFAULT 0 NOT NULL,
    expired boolean DEFAULT false NOT NULL,
    statistics boolean DEFAULT false NOT NULL,
    score numeric(20,2) DEFAULT 0 NOT NULL,
    create_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    update_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    creator character varying(15) DEFAULT ''::character varying NOT NULL,
    updater character varying(15) DEFAULT ''::character varying NOT NULL,
    deleted smallint DEFAULT 1 NOT NULL
);


ALTER TABLE public.t_facebook_info OWNER TO postgres;

--
-- Name: COLUMN t_facebook_info.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.id IS '主键';


--
-- Name: COLUMN t_facebook_info.profile_image; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.profile_image IS '用户头像';


--
-- Name: COLUMN t_facebook_info.profile_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.profile_id IS '账户id';


--
-- Name: COLUMN t_facebook_info.profile_url; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.profile_url IS '个人主页连接';


--
-- Name: COLUMN t_facebook_info.username; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.username IS '账户名称';


--
-- Name: COLUMN t_facebook_info.push_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.push_time IS '推文的时间戳';


--
-- Name: COLUMN t_facebook_info.post_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.post_id IS '帖子ID';


--
-- Name: COLUMN t_facebook_info.tags; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.tags IS '标签集合';


--
-- Name: COLUMN t_facebook_info.content; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.content IS '推文内容';


--
-- Name: COLUMN t_facebook_info.comments; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.comments IS '评论数';


--
-- Name: COLUMN t_facebook_info.retweets; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.retweets IS '转发数';


--
-- Name: COLUMN t_facebook_info.likes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.likes IS '点赞数';


--
-- Name: COLUMN t_facebook_info.expired; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.expired IS '是否过期';


--
-- Name: COLUMN t_facebook_info.statistics; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.statistics IS '是否统计过';


--
-- Name: COLUMN t_facebook_info.score; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.score IS '得分';


--
-- Name: COLUMN t_facebook_info.create_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.create_time IS '当前记录的创建时间';


--
-- Name: COLUMN t_facebook_info.update_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.update_time IS '当前记录的更新时间';


--
-- Name: COLUMN t_facebook_info.creator; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.creator IS '当前记录的创建人';


--
-- Name: COLUMN t_facebook_info.updater; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.updater IS '当前记录的更新人';


--
-- Name: COLUMN t_facebook_info.deleted; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_facebook_info.deleted IS '删除标志（0: 未删除, 1: 已删除）';


--
-- Name: t_facebook_info_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.t_facebook_info_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.t_facebook_info_id_seq OWNER TO postgres;

--
-- Name: t_facebook_info_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.t_facebook_info_id_seq OWNED BY public.t_facebook_info.id;


--
-- Name: t_fee_tolerance; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_fee_tolerance (
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "MaxLessRate" numeric(5,2) NOT NULL
);


ALTER TABLE public.t_fee_tolerance OWNER TO postgres;

--
-- Name: t_game_rpc_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_game_rpc_config (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RpcUrl" character varying(1024) NOT NULL,
    "WssUrl" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_game_rpc_config OWNER TO postgres;

--
-- Name: t_game_scan_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_game_scan_info (
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "UntilTxId" character varying(128) NOT NULL,
    "BeforeTxId" character varying(128)
);


ALTER TABLE public.t_game_scan_info OWNER TO postgres;

--
-- Name: t_game_tx; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_game_tx (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "TxType" integer NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_game_tx OWNER TO postgres;

--
-- Name: t_global_daily_exchange_limit; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_global_daily_exchange_limit (
    id integer NOT NULL,
    daily_limit numeric(78,0) DEFAULT 0 NOT NULL,
    quota_date date DEFAULT CURRENT_DATE NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.t_global_daily_exchange_limit OWNER TO postgres;

--
-- Name: t_global_daily_exchange_limit_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.t_global_daily_exchange_limit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.t_global_daily_exchange_limit_id_seq OWNER TO postgres;

--
-- Name: t_global_daily_exchange_limit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.t_global_daily_exchange_limit_id_seq OWNED BY public.t_global_daily_exchange_limit.id;


--
-- Name: t_helius_api; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_helius_api (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "BaseURL" character varying(128) NOT NULL,
    "ApiKey" character varying(42) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_helius_api OWNER TO postgres;

--
-- Name: t_helius_scan_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_helius_scan_info (
    "Service" character varying(64) NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "UntilTxId" character varying(128),
    "BeforeTxId" character varying(128),
    "Slot" numeric(78,0) DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_helius_scan_info OWNER TO postgres;

--
-- Name: t_micro_service_lb; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_micro_service_lb (
    "Ip" character varying(16) NOT NULL,
    "AreaZone" character varying(32) NOT NULL,
    "Domain" character varying(128) NOT NULL
);


ALTER TABLE public.t_micro_service_lb OWNER TO postgres;

--
-- Name: t_pos_reward_key_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_pos_reward_key_config (
    "Service" character varying(64) NOT NULL,
    "EncryptedKey" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_pos_reward_key_config OWNER TO postgres;

--
-- Name: t_pos_rpc_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_pos_rpc_config (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RpcUrl" character varying(1024) NOT NULL,
    "WssUrl" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_pos_rpc_config OWNER TO postgres;

--
-- Name: t_pos_scan_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_pos_scan_info (
    "Service" character varying(64) NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "UntilTxId" character varying(128) NOT NULL,
    "BeforeTxId" character varying(128),
    "Slot" numeric(78,0) DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_pos_scan_info OWNER TO postgres;

--
-- Name: t_pos_tx; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_pos_tx (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "TxType" integer NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "State" integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_pos_tx OWNER TO postgres;

--
-- Name: t_rank_activity_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_rank_activity_rule (
    id integer NOT NULL,
    rule_code character varying(32) DEFAULT ''::character varying NOT NULL,
    rule_expr text NOT NULL,
    rule_level character varying(128) DEFAULT ''::character varying NOT NULL,
    rule_desc character varying(128) DEFAULT ''::character varying NOT NULL,
    rule_type character varying(24) DEFAULT ''::character varying NOT NULL,
    create_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    update_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    creator character varying(15) DEFAULT ''::character varying NOT NULL,
    updater character varying(15) DEFAULT ''::character varying NOT NULL,
    deleted smallint DEFAULT 1 NOT NULL
);


ALTER TABLE public.t_rank_activity_rule OWNER TO postgres;

--
-- Name: COLUMN t_rank_activity_rule.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.id IS '主键';


--
-- Name: COLUMN t_rank_activity_rule.rule_code; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.rule_code IS '规则编码';


--
-- Name: COLUMN t_rank_activity_rule.rule_expr; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.rule_expr IS '规则表达式';


--
-- Name: COLUMN t_rank_activity_rule.rule_level; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.rule_level IS '规则等级';


--
-- Name: COLUMN t_rank_activity_rule.rule_desc; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.rule_desc IS '规则说明';


--
-- Name: COLUMN t_rank_activity_rule.rule_type; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.rule_type IS '规则类型';


--
-- Name: COLUMN t_rank_activity_rule.create_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.create_time IS '当前记录的创建时间';


--
-- Name: COLUMN t_rank_activity_rule.update_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.update_time IS '当前记录的更新时间';


--
-- Name: COLUMN t_rank_activity_rule.creator; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.creator IS '当前记录的创建人';


--
-- Name: COLUMN t_rank_activity_rule.updater; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_rank_activity_rule.updater IS '当前记录的更新人';


--
-- Name: t_rank_activity_rule_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.t_rank_activity_rule_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.t_rank_activity_rule_id_seq OWNER TO postgres;

--
-- Name: t_rank_activity_rule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.t_rank_activity_rule_id_seq OWNED BY public.t_rank_activity_rule.id;


--
-- Name: t_reward_code; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_code (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RewardCode" character varying(6) NOT NULL,
    "RewardAmount" numeric(78,0) NOT NULL,
    "TxId" character varying(128) DEFAULT NULL::character varying,
    "RefBlockHash" character varying(45) DEFAULT NULL::character varying,
    "LastValidBlockHeight" bigint DEFAULT 0,
    "State" integer DEFAULT 0 NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "ExpireTime" timestamp without time zone DEFAULT (CURRENT_TIMESTAMP + '24:00:00'::interval),
    "NativeAccount" character varying(64) DEFAULT NULL::character varying
);


ALTER TABLE public.t_reward_code OWNER TO postgres;

--
-- Name: t_reward_code_fee; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_code_fee (
    "Amount" numeric(78,0) NOT NULL,
    "CostFeeRate" numeric(5,0) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_code_fee OWNER TO postgres;

--
-- Name: t_reward_code_key_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_code_key_config (
    "Chain" character varying(64) NOT NULL,
    "EncryptedKey" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_code_key_config OWNER TO postgres;

--
-- Name: t_reward_code_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_code_rule (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Decimal" integer NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "DexNativeAccount" character varying(64) NOT NULL,
    "DexFeeRate" numeric(78,0) NOT NULL,
    "MaxDexFee" numeric(78,0) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_code_rule OWNER TO postgres;

--
-- Name: t_reward_discount_rate; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_discount_rate (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Rate" numeric(3,2) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_discount_rate OWNER TO postgres;

--
-- Name: t_reward_exclude; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_exclude (
    "Chain" character varying(64) NOT NULL,
    "Account" character varying(64) NOT NULL,
    "Brand" character varying(64),
    "Remark" character varying(128),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_exclude OWNER TO postgres;

--
-- Name: t_reward_hacker; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_hacker (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_hacker OWNER TO postgres;

--
-- Name: t_reward_key_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_key_config (
    "Service" character varying(64) NOT NULL,
    "EncryptedKey" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_key_config OWNER TO postgres;

--
-- Name: t_reward_lottery; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_lottery (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "RewardAmount" numeric(78,0),
    "RewardType" integer,
    "State" integer DEFAULT 0,
    "Pending" boolean DEFAULT false,
    "Day" date NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_lottery OWNER TO postgres;

--
-- Name: t_reward_lottery_claim_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_lottery_claim_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RewardIds" public.ulid[] NOT NULL,
    "TxId" character varying(128),
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint,
    "State" integer DEFAULT 0,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_lottery_claim_record OWNER TO postgres;

--
-- Name: t_reward_private_key; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_private_key (
    "ServiceType" integer NOT NULL,
    "ServiceName" character varying(64) NOT NULL,
    "EncryptedKey" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_private_key OWNER TO postgres;

--
-- Name: t_reward_rpc_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_rpc_config (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RpcUrl" character varying(1024) NOT NULL,
    "WssUrl" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_rpc_config OWNER TO postgres;

--
-- Name: t_reward_scan_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_scan_info (
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "UntilTxId" character varying(128) NOT NULL,
    "BeforeTxId" character varying(128),
    "Slot" numeric(78,0) DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_reward_scan_info OWNER TO postgres;

--
-- Name: t_reward_test; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_test (
    "Chain" character varying(64) NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_reward_test OWNER TO postgres;

--
-- Name: t_reward_tx; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_tx (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "TxType" integer NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "State" integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_reward_tx OWNER TO postgres;

--
-- Name: t_reward_tx_scan; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_reward_tx_scan (
    "ServiceType" integer NOT NULL,
    "ServiceName" character varying(64) NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "UntilTxId" character varying(128) NOT NULL,
    "BeforeTxId" character varying(128),
    "Slot" numeric(78,0)
);


ALTER TABLE public.t_reward_tx_scan OWNER TO postgres;

--
-- Name: t_shit_house; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_shit_house (
    "HouseId" integer NOT NULL,
    "HouseName" character varying(64) NOT NULL,
    "HouseLogo" character varying(1024),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "NativeAccount" character varying(64) DEFAULT NULL::character varying,
    "Limit" numeric(78,0) DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_shit_house OWNER TO postgres;

--
-- Name: t_shit_house_token_quota; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_shit_house_token_quota (
    "HouseId" integer NOT NULL,
    "Total" numeric(78,0) DEFAULT 0 NOT NULL,
    "Locked" numeric(78,0) DEFAULT 0 NOT NULL,
    "Available" numeric(78,0) DEFAULT 0 NOT NULL,
    "Version" integer DEFAULT 0 NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "Batch" integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_shit_house_token_quota OWNER TO postgres;

--
-- Name: t_sol_airdrop_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_airdrop_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "AirDropTokenAccount" character varying(64) NOT NULL,
    "AirDropNativeAccount" character varying(64) NOT NULL,
    "ReceiptTokenAccount" character varying(64) NOT NULL,
    "ReceiptNativeAccount" character varying(64) NOT NULL,
    "Amount" numeric(78,0) NOT NULL,
    "AirDropTxId" character varying(128) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_airdrop_record OWNER TO postgres;

--
-- Name: t_sol_campaign_exchange_score_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_campaign_exchange_score_rule (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Decimal" integer NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "Rate" numeric(5,2) NOT NULL,
    "DexNativeAccount" character varying(64) NOT NULL,
    "CostRate" numeric(5,2) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_campaign_exchange_score_rule OWNER TO postgres;

--
-- Name: t_sol_determine_invite_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_determine_invite_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "InviterTokenAccount" character varying(64) NOT NULL,
    "InviterNativeAccount" character varying(64) NOT NULL,
    "InviteeTokenAccount" character varying(64) NOT NULL,
    "InviteeNativeAccount" character varying(64) NOT NULL,
    "TransferTxId" character varying(128),
    "InviteChannel" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "Level" integer DEFAULT 1
);


ALTER TABLE public.t_sol_determine_invite_record OWNER TO postgres;

--
-- Name: t_sol_exchange_campaign_score_to_token_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_exchange_campaign_score_to_token_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "ReceiptTokenAccount" character varying(64) NOT NULL,
    "ReceiptNativeAccount" character varying(64) NOT NULL,
    "Provider" character varying(64) NOT NULL,
    "UserId" character varying(64) NOT NULL,
    "ExchangeTxId" character varying(128) NOT NULL,
    "ScoreFlowId" integer NOT NULL,
    "ScoreTxId" character varying(64) NOT NULL,
    "Amount" numeric(78,0) NOT NULL,
    "Score" numeric(78,0) NOT NULL,
    "State" integer,
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_exchange_campaign_score_to_token_record OWNER TO postgres;

--
-- Name: t_sol_fee_statistics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_fee_statistics (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Slot" bigint,
    "TransactionIndex" integer,
    "BlockHash" character varying(64),
    "TransactionId" character varying(128),
    "ComputeUnitPrice" numeric(78,0),
    "ComputeUnitLimit" numeric(78,0),
    "UnitsConsumed" numeric(78,0),
    "Fee" numeric(78,0),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_fee_statistics OWNER TO postgres;

--
-- Name: t_sol_fund_flow; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_fund_flow (
    record_id public.ulid DEFAULT public.gen_ulid() NOT NULL,
    brand character varying(64) NOT NULL,
    token_symbol character varying(64) NOT NULL,
    is_token boolean NOT NULL,
    from_native_account character varying(64) NOT NULL,
    to_native_account character varying(64) NOT NULL,
    tx_id character varying(128) NOT NULL,
    direction smallint NOT NULL,
    service_type smallint NOT NULL,
    flow_type smallint NOT NULL,
    decimals smallint NOT NULL,
    amount numeric(78,0) NOT NULL,
    create_time timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_fund_flow OWNER TO postgres;

--
-- Name: t_sol_fund_flow_old; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_fund_flow_old (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "IsToken" boolean NOT NULL,
    "FromNativeAccount" character varying(64) NOT NULL,
    "ToNativeAccount" character varying(64) NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "Direction" integer NOT NULL,
    "ServiceType" integer NOT NULL,
    "FlowType" integer NOT NULL,
    "Decimals" integer NOT NULL,
    "Amount" numeric(78,0) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_fund_flow_old OWNER TO postgres;

--
-- Name: t_sol_game_buy_property_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_buy_property_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "PropertyId" public.ulid NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Price" numeric(78,0) NOT NULL,
    "Quantity" bigint NOT NULL,
    "Amount" numeric(78,0) NOT NULL,
    "State" integer DEFAULT 0,
    "ErrorCode" character varying(32) DEFAULT NULL::character varying,
    "ErrorMsg" character varying(128) DEFAULT NULL::character varying,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint
);


ALTER TABLE public.t_sol_game_buy_property_record OWNER TO postgres;

--
-- Name: t_sol_game_exchange_prize_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_exchange_prize_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "PackId" public.ulid NOT NULL,
    "PrizeId" public.ulid NOT NULL,
    "Amount" bigint NOT NULL,
    "TxId" character varying(128),
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint,
    "State" integer,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_exchange_prize_record OWNER TO postgres;

--
-- Name: t_sol_game_pack_account; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_pack_account (
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "PackId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "PackName" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_pack_account OWNER TO postgres;

--
-- Name: t_sol_game_prize_account; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_prize_account (
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "PackId" public.ulid NOT NULL,
    "PrizeId" public.ulid NOT NULL,
    "Amount" integer NOT NULL,
    "Version" integer DEFAULT 1,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_prize_account OWNER TO postgres;

--
-- Name: t_sol_game_prize_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_prize_info (
    "Product" character varying(64) NOT NULL,
    "PrizeId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "PrizeName" character varying(64) NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "Decimals" integer NOT NULL,
    "Default" boolean NOT NULL,
    "SwapAmount" numeric(78,0) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_prize_info OWNER TO postgres;

--
-- Name: t_sol_game_property_account; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_property_account (
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "PropertyId" public.ulid NOT NULL,
    "Quantity" bigint NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "Remark" character varying(128) DEFAULT NULL::character varying
);


ALTER TABLE public.t_sol_game_property_account OWNER TO postgres;

--
-- Name: t_sol_game_property_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_property_info (
    "Product" character varying(64) NOT NULL,
    "PropertyId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "PropertyName" character varying(64) NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Decimals" integer NOT NULL,
    "Price" numeric(78,0) NOT NULL
);


ALTER TABLE public.t_sol_game_property_info OWNER TO postgres;

--
-- Name: t_sol_game_property_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_property_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "PropertyId" public.ulid NOT NULL,
    "Quantity" bigint NOT NULL,
    "ActionType" character varying(10) NOT NULL,
    "Remark" character varying(128) DEFAULT NULL::character varying,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_property_record OWNER TO postgres;

--
-- Name: t_sol_game_rank_top; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_rank_top (
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "Point" bigint NOT NULL
);


ALTER TABLE public.t_sol_game_rank_top OWNER TO postgres;

--
-- Name: t_sol_game_receipt_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_receipt_info (
    "Product" character varying(64) NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_receipt_info OWNER TO postgres;

--
-- Name: t_sol_game_register_claim_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_register_claim_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Product" character varying(64) NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "UserName" character varying(64) NOT NULL,
    "Amount" bigint NOT NULL,
    "TxId" character varying(128),
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint,
    "State" integer,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_register_claim_record OWNER TO postgres;

--
-- Name: t_sol_game_register_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_register_info (
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "UserName" character varying(64) NOT NULL,
    "State" integer NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_register_info OWNER TO postgres;

--
-- Name: t_sol_game_role_account; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_role_account (
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "RoleId" public.ulid NOT NULL,
    "Purchased" boolean DEFAULT false,
    "VoteAmount" bigint DEFAULT 0,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_role_account OWNER TO postgres;

--
-- Name: t_sol_game_role_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_role_info (
    "Product" character varying(64) NOT NULL,
    "RoleId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RoleName" character varying(64) NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Decimals" integer NOT NULL,
    "Price" numeric(78,0) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_role_info OWNER TO postgres;

--
-- Name: t_sol_game_unlock_role_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_unlock_role_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "RoleId" public.ulid NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Decimals" integer NOT NULL,
    "Price" bigint NOT NULL,
    "State" integer DEFAULT 0,
    "ErrorCode" character varying(32) DEFAULT NULL::character varying,
    "ErrorMsg" character varying(128) DEFAULT NULL::character varying,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint
);


ALTER TABLE public.t_sol_game_unlock_role_record OWNER TO postgres;

--
-- Name: t_sol_game_vote_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_game_vote_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Product" character varying(64) NOT NULL,
    "UserId" public.ulid NOT NULL,
    "RoleId" public.ulid NOT NULL,
    "Quantity" bigint NOT NULL,
    "State" boolean DEFAULT false,
    "Remark" character varying(128) DEFAULT NULL::character varying,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_game_vote_record OWNER TO postgres;

--
-- Name: t_sol_native_account_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_native_account_info (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "InviteCode" character varying(16) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_native_account_info OWNER TO postgres;

--
-- Name: t_sol_official_give_token_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_official_give_token_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "ReceiptTokenAccount" character varying(64) NOT NULL,
    "ReceiptNativeAccount" character varying(64) NOT NULL,
    "DexNativeAccount" character varying(64) NOT NULL,
    "RewardTxId" character varying(128) NOT NULL,
    "Amount" numeric(78,0) NOT NULL,
    "DexFee" numeric(78,0) NOT NULL,
    "UseInviteCode" boolean NOT NULL,
    "InviteCode" character varying(16),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "State" integer,
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint,
    "DetermineInvite" boolean
);


ALTER TABLE public.t_sol_official_give_token_record OWNER TO postgres;

--
-- Name: t_sol_official_give_token_reward_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_official_give_token_reward_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "ReceiptTokenAccount" character varying(64) NOT NULL,
    "ReceiptNativeAccount" character varying(64) NOT NULL,
    "DexNativeAccount" character varying(64) NOT NULL,
    "RewardTxId" character varying(64) NOT NULL,
    "Amount" numeric(78,0) NOT NULL,
    "DexFee" numeric(78,0) NOT NULL,
    "UseInviteCode" boolean NOT NULL,
    "InviteCode" character varying(16),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_official_give_token_reward_record OWNER TO postgres;

--
-- Name: t_sol_official_give_token_reward_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_official_give_token_reward_rule (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "InviteCode" character varying(16) DEFAULT NULL::character varying,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Decimal" integer NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "DexNativeAccount" character varying(64) NOT NULL,
    "Amount" numeric(78,0) NOT NULL,
    "InviteAmount" numeric(78,0) NOT NULL,
    "DexFeeRate" numeric(78,0) NOT NULL,
    "MaxDexFee" numeric(78,0) NOT NULL,
    "Interval" integer NOT NULL,
    "Default" boolean DEFAULT false,
    "RewardInviter" boolean DEFAULT true,
    "DetermineInvite" boolean DEFAULT true,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_official_give_token_reward_rule OWNER TO postgres;

--
-- Name: t_sol_official_transfer_token_reward_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_official_transfer_token_reward_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "FromTokenAccount" character varying(64) NOT NULL,
    "FromNativeAccount" character varying(64) NOT NULL,
    "ToTokenAccount" character varying(64) NOT NULL,
    "ToNativeAccount" character varying(64) NOT NULL,
    "RewardAmount" numeric(78,0) NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_official_transfer_token_reward_record OWNER TO postgres;

--
-- Name: t_sol_official_transfer_token_reward_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_official_transfer_token_reward_rule (
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Decimal" integer NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "DexNativeAccount" character varying(64) NOT NULL,
    "RewardRate" numeric(78,0) NOT NULL,
    "RewardValidAddressRate" numeric(78,0) NOT NULL,
    "MaxRewardPerTx" numeric(78,0) NOT NULL,
    "MaxValidAddressRewardPerTx" numeric(78,0) NOT NULL,
    "DexFeeRate" numeric(78,0) NOT NULL,
    "MaxDexFee" numeric(78,0) NOT NULL,
    "Interval" integer NOT NULL,
    "CreateTime" timestamp(6) without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp(6) without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_official_transfer_token_reward_rule OWNER TO postgres;

--
-- Name: t_sol_pos_mission_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_mission_config (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RewardType" integer NOT NULL,
    "Starred" boolean NOT NULL,
    "Rate" numeric(5,2),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_mission_config OWNER TO postgres;

--
-- Name: t_sol_pos_retweet_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_retweet_config (
    "UserID" character varying(20) NOT NULL,
    "RetweetId" character varying(20) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_retweet_config OWNER TO postgres;

--
-- Name: t_sol_pos_reward; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_reward (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "GroupId" character varying(64) NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "StarLevel" integer DEFAULT 1,
    "Base" numeric(78,0),
    "Rate" numeric(5,2),
    "RewardAmount" numeric(78,0),
    "RewardType" numeric(78,0),
    "State" integer DEFAULT 0,
    "Starred" boolean NOT NULL,
    "Pending" boolean DEFAULT false,
    "Day" date NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_reward OWNER TO postgres;

--
-- Name: t_sol_pos_reward_claim_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_reward_claim_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RewardIds" public.ulid[] NOT NULL,
    "TxId" character varying(128),
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint,
    "State" integer DEFAULT 0,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_reward_claim_record OWNER TO postgres;

--
-- Name: t_sol_pos_reward_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_reward_rule (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Decimal" integer NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "DexNativeAccount" character varying(64) NOT NULL,
    "DexFeeRate" numeric(78,0) NOT NULL,
    "MaxDexFee" numeric(78,0) NOT NULL,
    "QuoteTokenAmount" numeric(78,0) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_reward_rule OWNER TO postgres;

--
-- Name: t_sol_pos_snap_shot; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_snap_shot (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "Amount" numeric(78,0),
    "StarLevel" integer DEFAULT 0,
    "Rate" numeric(5,2),
    "RangeBase" numeric(5,2),
    "Day" date NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_snap_shot OWNER TO postgres;

--
-- Name: t_sol_pos_social_media_mission_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_social_media_mission_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Provider" character varying(64) NOT NULL,
    "UserID" character varying(64) NOT NULL,
    "UserName" character varying(64) NOT NULL,
    "Type" integer NOT NULL,
    "State" integer DEFAULT 0,
    "Day" date NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_social_media_mission_record OWNER TO postgres;

--
-- Name: t_sol_pos_star_level_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_star_level_config (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "StarLevel" integer NOT NULL,
    "Rate" numeric(5,2),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_star_level_config OWNER TO postgres;

--
-- Name: t_sol_pos_star_level_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_star_level_rule (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Amount" numeric(78,2) NOT NULL,
    "GroupAmount" numeric(78,2) NOT NULL,
    "StarLevel" integer NOT NULL,
    "Rate" numeric(5,2) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_star_level_rule OWNER TO postgres;

--
-- Name: t_sol_pos_twitter_oauth2_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_pos_twitter_oauth2_config (
    "UserID" character varying(20) NOT NULL,
    "ClientId" character varying(34),
    "ClientSecret" character varying(50),
    "ApiKey" character varying(25),
    "ApiSecret" character varying(50),
    "BearerToken" character varying(256),
    "AccessToken" character varying(50),
    "AccessTokenSecret" character varying(50),
    "WebRedirectURL" character varying(1024),
    "OAuthCallBackURL" character varying(1024),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_pos_twitter_oauth2_config OWNER TO postgres;

--
-- Name: t_sol_qn_fee; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_qn_fee (
    "Id" integer NOT NULL,
    "Slot" bigint,
    "LowAvg" numeric(78,0),
    "MediumAvg" numeric(78,0),
    "HighAvg" numeric(78,0),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_qn_fee OWNER TO postgres;

--
-- Name: t_sol_scan_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_scan_info (
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "CreateTokenTxId" character varying(128) NOT NULL,
    "UntilTxId" character varying(128) NOT NULL,
    "BeforeTxId" character varying(128),
    "NativeAccount" character varying(64) DEFAULT ''::character varying,
    "TokenAccount" character varying(64) DEFAULT ''::character varying
);


ALTER TABLE public.t_sol_scan_info OWNER TO postgres;

--
-- Name: t_sol_stake_fix_interest_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_stake_fix_interest_config (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "MinAmount" numeric(78,0),
    "StakeType" integer NOT NULL,
    "FixRate" numeric(5,2),
    "IndividualRate" numeric(5,2),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_stake_fix_interest_config OWNER TO postgres;

--
-- Name: t_sol_stake_invite_dist; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_stake_invite_dist (
    "Level" integer NOT NULL
);


ALTER TABLE public.t_sol_stake_invite_dist OWNER TO postgres;

--
-- Name: t_sol_stake_invite_rate; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_stake_invite_rate (
    "Level" integer NOT NULL,
    "Rate" numeric(5,2) NOT NULL
);


ALTER TABLE public.t_sol_stake_invite_rate OWNER TO postgres;

--
-- Name: t_sol_stake_reward; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_stake_reward (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "GroupId" character varying(64) NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "StarLevel" integer DEFAULT 0,
    "Base" numeric(78,0),
    "Rate" numeric(5,2),
    "RewardAmount" numeric(78,0),
    "RewardType" integer NOT NULL,
    "State" integer DEFAULT 0,
    "Starred" boolean NOT NULL,
    "Pending" boolean DEFAULT false,
    "Day" date NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "TxId" character varying(128)
);


ALTER TABLE public.t_sol_stake_reward OWNER TO postgres;

--
-- Name: t_sol_stake_reward_claim_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_stake_reward_claim_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "RewardIds" public.ulid[] NOT NULL,
    "TxId" character varying(128),
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint,
    "State" integer DEFAULT 0,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_stake_reward_claim_record OWNER TO postgres;

--
-- Name: t_sol_stake_reward_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_stake_reward_rule (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Decimal" integer NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "StakeAdmin" character varying(64) NOT NULL,
    "ProgramId" character varying(64) NOT NULL,
    "FaucetTokenAccount" character varying(64) NOT NULL,
    "FaucetNativeAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "DexNativeAccount" character varying(64) NOT NULL,
    "QuoteTokenAmount" numeric(78,0) NOT NULL,
    "DexFeeRate" integer NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_stake_reward_rule OWNER TO postgres;

--
-- Name: t_sol_stake_snap_shot; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_stake_snap_shot (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "Amount" numeric(78,0),
    "StakeType" integer NOT NULL,
    "Day" date NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_stake_snap_shot OWNER TO postgres;

--
-- Name: t_sol_stake_star_level_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_stake_star_level_config (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "StarLevel" integer NOT NULL,
    "Rate" numeric(5,2),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_stake_star_level_config OWNER TO postgres;

--
-- Name: t_sol_stake_star_level_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_stake_star_level_rule (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Amount" numeric(78,2) NOT NULL,
    "GroupAmount" numeric(78,2) NOT NULL,
    "StarLevel" integer NOT NULL,
    "Rate" numeric(5,2) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_stake_star_level_rule OWNER TO postgres;

--
-- Name: t_sol_swap_new_token_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_swap_new_token_config (
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "OldTokenMintAccount" character varying(64) NOT NULL,
    "NewTokenMintAccount" character varying(64) NOT NULL,
    "OldDecimals" integer NOT NULL,
    "NewDecimals" integer NOT NULL,
    "ReceiptNativeAccount" character varying(64) NOT NULL,
    "ReceiptTokenAccount" character varying(64) NOT NULL,
    "SendNativeAccount" character varying(64) NOT NULL,
    "SendTokenAccount" character varying(64) NOT NULL,
    "Rate" numeric(5,2) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_swap_new_token_config OWNER TO postgres;

--
-- Name: t_sol_swap_token_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_swap_token_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "InputTokenSymbol" character varying(64) NOT NULL,
    "InputTokenMintAccount" character varying(64) NOT NULL,
    "InputTokenDecimal" integer NOT NULL,
    "OutputTokenSymbol" character varying(64) NOT NULL,
    "OutputTokenMintAccount" character varying(64) NOT NULL,
    "OutputTokenDecimal" integer NOT NULL,
    "UserNativeAccount" character varying(64) NOT NULL,
    "UserTokenAccount" character varying(64) NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "TxType" integer NOT NULL,
    "InputTokenAmount" numeric(78,0) NOT NULL,
    "OutputTokenAmount" numeric(78,0) NOT NULL,
    "State" integer,
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "HouseId" integer DEFAULT 0 NOT NULL,
    "Batch" integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.t_sol_swap_token_record OWNER TO postgres;

--
-- Name: t_sol_swap_token_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_swap_token_rule (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "InputTokenSymbol" character varying(64) NOT NULL,
    "InputTokenMintAccount" character varying(64) NOT NULL,
    "InputTokenDecimal" integer NOT NULL,
    "OutputTokenSymbol" character varying(64) NOT NULL,
    "OutputTokenMintAccount" character varying(64) NOT NULL,
    "OutputTokenDecimal" integer NOT NULL,
    "OutputTokenAccount" character varying(64) NOT NULL,
    "OutputNativeAccount" character varying(64) NOT NULL,
    "CostTokenAccount" character varying(64) NOT NULL,
    "CostNativeAccount" character varying(64) NOT NULL,
    "Rate" numeric(5,2) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_swap_token_rule OWNER TO postgres;

--
-- Name: t_sol_token_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_token_config (
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Decimal" integer NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "MinterNativeAccount" character varying(64) NOT NULL,
    "AirDropTokenAccount" character varying(64) NOT NULL,
    "AirDropNativeAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_token_config OWNER TO postgres;

--
-- Name: t_sol_transfer_checked_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_transfer_checked_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "FromTokenAccount" character varying(64) NOT NULL,
    "FromNativeAccount" character varying(64) NOT NULL,
    "ReceiptTokenAccount" character varying(64) NOT NULL,
    "ReceiptNativeAccount" character varying(64) NOT NULL,
    "OwnerNativeAccount" character varying(64) NOT NULL,
    "TransferTxId" character varying(128) NOT NULL,
    "InstructionIndex" integer NOT NULL,
    "TransferAmount" numeric(78,0) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "State" integer,
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint
);


ALTER TABLE public.t_sol_transfer_checked_record OWNER TO postgres;

--
-- Name: t_sol_transfer_reward_claim; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_transfer_reward_claim (
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Level" integer NOT NULL,
    "ClaimRatio" numeric(5,2) NOT NULL
);


ALTER TABLE public.t_sol_transfer_reward_claim OWNER TO postgres;

--
-- Name: t_sol_transfer_reward_distribution; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_transfer_reward_distribution (
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Level" integer NOT NULL
);


ALTER TABLE public.t_sol_transfer_reward_distribution OWNER TO postgres;

--
-- Name: t_sol_transfer_reward_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_transfer_reward_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "FromTokenAccount" character varying(64) NOT NULL,
    "FromNativeAccount" character varying(64) NOT NULL,
    "ReceiptTokenAccount" character varying(64) NOT NULL,
    "ReceiptNativeAccount" character varying(64) NOT NULL,
    "TransferAmount" numeric(78,0) NOT NULL,
    "TransferTxId" character varying(128) NOT NULL,
    "RewardTxId" character varying(128),
    "State" integer,
    "ErrorMessage" character varying(256),
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "Version" integer DEFAULT 0 NOT NULL,
    "RefBlockHash" character varying(45),
    "LastValidBlockHeight" bigint,
    "RewardState" integer
);


ALTER TABLE public.t_sol_transfer_reward_record OWNER TO postgres;

--
-- Name: t_sol_transfer_to_dex_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_transfer_to_dex_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "FromNativeAccount" character varying(64) NOT NULL,
    "ReceiptNativeAccount" character varying(64) NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "TransferAmount" numeric(78,0) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_transfer_to_dex_record OWNER TO postgres;

--
-- Name: t_sol_transfer_token_reward_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_sol_transfer_token_reward_rule (
    "Brand" character varying(64) NOT NULL,
    "TokenSymbol" character varying(64) NOT NULL,
    "Decimal" integer NOT NULL,
    "TokenMintAccount" character varying(64) NOT NULL,
    "RewardTokenAccount" character varying(64) NOT NULL,
    "RewardNativeAccount" character varying(64) NOT NULL,
    "AccountExistSlot" bigint NOT NULL,
    "AccountExistBufferSlot" bigint NOT NULL,
    "RewardRate" numeric(78,0) NOT NULL,
    "MaxRewardPerTx" numeric(78,0) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_sol_transfer_token_reward_rule OWNER TO postgres;

--
-- Name: t_stake_amm_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_stake_amm_config (
    "QuoteToken" character varying(64) NOT NULL,
    "PublicKey" character varying(64) NOT NULL
);


ALTER TABLE public.t_stake_amm_config OWNER TO postgres;

--
-- Name: t_stake_area_leader; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_stake_area_leader (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_stake_area_leader OWNER TO postgres;

--
-- Name: t_stake_buy_token; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_stake_buy_token (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "TxId" character varying(128) NOT NULL,
    "Slot" numeric(78,0) NOT NULL,
    "Source" character varying(64) NOT NULL,
    "Destination" character varying(64) NOT NULL,
    "Amount" numeric(78,0) NOT NULL,
    "Locked" boolean DEFAULT false NOT NULL,
    "LockedBy" character varying(128),
    "LockedAt" timestamp(6) with time zone,
    "StakedAmount" numeric(78,0) DEFAULT 0 NOT NULL,
    "RemainingAmount" numeric(78,0) DEFAULT 0 NOT NULL,
    "CreatedAt" timestamp(6) with time zone DEFAULT now()
);


ALTER TABLE public.t_stake_buy_token OWNER TO postgres;

--
-- Name: t_stake_record; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_stake_record (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Staker" character varying(64),
    "StakeAmount" numeric(78,0),
    "State" integer DEFAULT 0 NOT NULL,
    "StakeTxHash" character varying(128),
    "UsedTxIds" text,
    "LockedTxIds" text,
    "CreatedAt" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdatedAt" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_stake_record OWNER TO postgres;

--
-- Name: t_stake_token_pool; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_stake_token_pool (
    "Source" character varying(64) NOT NULL,
    "FromTokenAccount" character varying(64) NOT NULL
);


ALTER TABLE public.t_stake_token_pool OWNER TO postgres;

--
-- Name: t_stake_total_area_leader; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_stake_total_area_leader (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "NativeAccount" character varying(64) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_stake_total_area_leader OWNER TO postgres;

--
-- Name: t_swap_key_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_swap_key_config (
    "Chain" character varying(64) NOT NULL,
    "EncryptedKey" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_swap_key_config OWNER TO postgres;

--
-- Name: t_swap_scan_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_swap_scan_info (
    "NativeAccount" character varying(64) NOT NULL,
    "TokenAccount" character varying(64) NOT NULL,
    "UntilTxId" character varying(128) NOT NULL,
    "BeforeTxId" character varying(128),
    "Slot" numeric(78,0)
);


ALTER TABLE public.t_swap_scan_info OWNER TO postgres;

--
-- Name: t_system_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_system_config (
    "Env" integer NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "UpdateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_system_config OWNER TO postgres;

--
-- Name: t_tweets_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_tweets_info (
    id integer NOT NULL,
    name character varying(255) DEFAULT ''::character varying NOT NULL,
    handle character varying(124) DEFAULT ''::character varying NOT NULL,
    push_time timestamp(3) without time zone NOT NULL,
    verified boolean DEFAULT false NOT NULL,
    content text NOT NULL,
    comments integer DEFAULT 0 NOT NULL,
    retweets integer DEFAULT 0 NOT NULL,
    likes integer DEFAULT 0 NOT NULL,
    analytics integer DEFAULT 0 NOT NULL,
    tags text,
    mentions text,
    emojis text,
    profile_image text,
    tweet_link text,
    tweet_id bigint NOT NULL,
    expired boolean DEFAULT false NOT NULL,
    statistics boolean DEFAULT false NOT NULL,
    score numeric(20,2) DEFAULT 0 NOT NULL,
    create_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    update_time timestamp(3) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    creator character varying(15) DEFAULT ''::character varying NOT NULL,
    updater character varying(15) DEFAULT ''::character varying NOT NULL,
    deleted smallint DEFAULT 1 NOT NULL
);


ALTER TABLE public.t_tweets_info OWNER TO postgres;

--
-- Name: COLUMN t_tweets_info.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.id IS '主键';


--
-- Name: COLUMN t_tweets_info.name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.name IS '账户名称，例如 "0://bank.BigD33"';


--
-- Name: COLUMN t_tweets_info.handle; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.handle IS '账户@，例如 "@BigD_Energy33"';


--
-- Name: COLUMN t_tweets_info.push_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.push_time IS '推文的时间戳';


--
-- Name: COLUMN t_tweets_info.verified; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.verified IS '账户是否已验证';


--
-- Name: COLUMN t_tweets_info.content; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.content IS '推文内容';


--
-- Name: COLUMN t_tweets_info.comments; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.comments IS '评论数';


--
-- Name: COLUMN t_tweets_info.retweets; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.retweets IS '转发数';


--
-- Name: COLUMN t_tweets_info.likes; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.likes IS '点赞数';


--
-- Name: COLUMN t_tweets_info.analytics; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.analytics IS '查看数量';


--
-- Name: COLUMN t_tweets_info.tags; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.tags IS '推文中的标签';


--
-- Name: COLUMN t_tweets_info.mentions; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.mentions IS '推文中提及的用户';


--
-- Name: COLUMN t_tweets_info.emojis; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.emojis IS '推文中使用的表情符号';


--
-- Name: COLUMN t_tweets_info.profile_image; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.profile_image IS '用户的头像链接';


--
-- Name: COLUMN t_tweets_info.tweet_link; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.tweet_link IS '推文的链接';


--
-- Name: COLUMN t_tweets_info.tweet_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.tweet_id IS '推文的唯一标识符（Tweet_ID）';


--
-- Name: COLUMN t_tweets_info.create_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.create_time IS '当前记录的创建时间';


--
-- Name: COLUMN t_tweets_info.update_time; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.update_time IS '当前记录的更新时间';


--
-- Name: COLUMN t_tweets_info.creator; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.creator IS '当前记录的创建人';


--
-- Name: COLUMN t_tweets_info.updater; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.updater IS '当前记录的更新人';


--
-- Name: COLUMN t_tweets_info.deleted; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.t_tweets_info.deleted IS '删除标志（0: 未删除, 1: 已删除）';


--
-- Name: t_tweets_info_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.t_tweets_info_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.t_tweets_info_id_seq OWNER TO postgres;

--
-- Name: t_tweets_info_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.t_tweets_info_id_seq OWNED BY public.t_tweets_info.id;


--
-- Name: t_user_daily_exchange_quota; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_user_daily_exchange_quota (
    id integer NOT NULL,
    user_id character varying(64) NOT NULL,
    quota_date date NOT NULL,
    max_quota numeric(78,0) DEFAULT 50000000 NOT NULL,
    frozen_quota numeric(78,0) DEFAULT 0 NOT NULL,
    available_quota numeric(78,0) DEFAULT 50000000 NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.t_user_daily_exchange_quota OWNER TO postgres;

--
-- Name: t_user_daily_exchange_quota_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.t_user_daily_exchange_quota_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.t_user_daily_exchange_quota_id_seq OWNER TO postgres;

--
-- Name: t_user_daily_exchange_quota_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.t_user_daily_exchange_quota_id_seq OWNED BY public.t_user_daily_exchange_quota.id;


--
-- Name: t_user_wallet_rpc_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.t_user_wallet_rpc_config (
    "RecordId" public.ulid DEFAULT public.gen_ulid() NOT NULL,
    "Chain" character varying(1024) NOT NULL,
    "RpcUrl" character varying(1024) NOT NULL,
    "WssUrl" character varying(1024) NOT NULL,
    "CreateTime" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.t_user_wallet_rpc_config OWNER TO postgres;

--
-- Name: user_activity_message; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_activity_message (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    row_id bigint NOT NULL,
    create_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp with time zone NOT NULL,
    read_status boolean DEFAULT false NOT NULL,
    message_type character varying(16) NOT NULL,
    CONSTRAINT user_activity_message_message_type_check CHECK (((message_type)::text = ANY ((ARRAY['POST'::character varying, 'REWARD'::character varying, 'SHARE_REWARD'::character varying, 'EXCHANGE'::character varying])::text[])))
);


ALTER TABLE public.user_activity_message OWNER TO meme_server;

--
-- Name: user_activity_message_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_activity_message_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_activity_message_id_seq OWNER TO meme_server;

--
-- Name: user_activity_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_activity_message_id_seq OWNED BY public.user_activity_message.id;


--
-- Name: user_basic; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_basic (
    user_id bigint NOT NULL,
    invite_by bigint,
    create_at timestamp with time zone DEFAULT now(),
    update_at timestamp with time zone NOT NULL,
    kol boolean DEFAULT false NOT NULL,
    username character varying(50),
    nick_name character varying(50) DEFAULT 'Shittie'::character varying,
    email character varying(128),
    provider_user_id character varying(32),
    provider character varying(32)
);


ALTER TABLE public.user_basic OWNER TO meme_server;

--
-- Name: user_basic_old; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_basic_old (
    user_id bigint NOT NULL,
    invite_by bigint,
    create_at timestamp with time zone DEFAULT now(),
    update_at timestamp with time zone NOT NULL,
    kol boolean DEFAULT false NOT NULL,
    username character varying(50),
    provider_user_id character varying(32) NOT NULL,
    provider character varying(32) NOT NULL
);


ALTER TABLE public.user_basic_old OWNER TO meme_server;

--
-- Name: user_basic_user_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_basic_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_basic_user_id_seq OWNER TO meme_server;

--
-- Name: user_basic_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_basic_user_id_seq OWNED BY public.user_basic.user_id;


--
-- Name: user_extend_info; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_extend_info (
    user_id bigint NOT NULL,
    create_at timestamp with time zone DEFAULT now(),
    update_at timestamp with time zone NOT NULL,
    birthday date,
    region_code character varying(6),
    phone_number character varying(11),
    avatar_picture character varying(100),
    location character varying(256)
);


ALTER TABLE public.user_extend_info OWNER TO meme_server;

--
-- Name: user_favorite_product; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_favorite_product (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    product_id character varying(36) NOT NULL,
    create_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.user_favorite_product OWNER TO meme_server;

--
-- Name: user_favorite_product_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_favorite_product_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_favorite_product_id_seq OWNER TO meme_server;

--
-- Name: user_favorite_product_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_favorite_product_id_seq OWNED BY public.user_favorite_product.id;


--
-- Name: user_game_item; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_game_item (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    item_id character varying(32),
    item_number integer NOT NULL,
    create_at timestamp with time zone DEFAULT now(),
    update_at timestamp with time zone NOT NULL,
    CONSTRAINT game_score_number_check CHECK ((item_number >= 0))
);


ALTER TABLE public.user_game_item OWNER TO meme_server;

--
-- Name: user_game_item_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_game_item_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_game_item_id_seq OWNER TO meme_server;

--
-- Name: user_game_item_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_game_item_id_seq OWNED BY public.user_game_item.id;


--
-- Name: user_game_item_op_record; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_game_item_op_record (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    item_id character varying(32),
    item_number integer NOT NULL,
    create_at timestamp with time zone DEFAULT now(),
    op_type character varying(20) NOT NULL,
    game_score_log_id bigint,
    CONSTRAINT game_score_number_check CHECK ((item_number > 0)),
    CONSTRAINT user_game_item_op_record_op_type_check CHECK (((op_type)::text = ANY (ARRAY[('EXCHANGE'::character varying)::text, ('CONSUME'::character varying)::text])))
);


ALTER TABLE public.user_game_item_op_record OWNER TO meme_server;

--
-- Name: user_game_item_op_record_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_game_item_op_record_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_game_item_op_record_id_seq OWNER TO meme_server;

--
-- Name: user_game_item_op_record_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_game_item_op_record_id_seq OWNED BY public.user_game_item_op_record.id;


--
-- Name: user_invitation_enroll; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_invitation_enroll (
    id bigint NOT NULL,
    inviter bigint NOT NULL,
    inviter_invite_by bigint,
    create_time timestamp with time zone,
    update_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    kol boolean NOT NULL,
    invite_code character varying(40) NOT NULL
);


ALTER TABLE public.user_invitation_enroll OWNER TO meme_server;

--
-- Name: user_invitation_enroll_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_invitation_enroll_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_invitation_enroll_id_seq OWNER TO meme_server;

--
-- Name: user_invitation_enroll_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_invitation_enroll_id_seq OWNED BY public.user_invitation_enroll.id;


--
-- Name: user_posts_task; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_posts_task (
    task_id bigint NOT NULL,
    user_id bigint NOT NULL,
    likes integer DEFAULT 0,
    shares integer DEFAULT 0,
    post_comments integer DEFAULT 0,
    create_time timestamp with time zone NOT NULL,
    update_time timestamp with time zone NOT NULL,
    task_name character varying(32) NOT NULL,
    task_status character varying(16) DEFAULT 'PENDING'::character varying NOT NULL,
    social_media character varying(16),
    screenshot_key character varying(100),
    screenshot_postfix character varying(8),
    post_url character varying(256),
    auditor character varying(16) DEFAULT NULL::character varying,
    audit_comment text,
    related_data jsonb,
    CONSTRAINT user_posts_task_social_media_check CHECK (((social_media)::text = ANY (ARRAY[('facebook'::character varying)::text, ('x'::character varying)::text, ('instagram'::character varying)::text, ('tiktok'::character varying)::text, ('instagram'::character varying)::text]))),
    CONSTRAINT user_posts_task_task_status_check CHECK (((task_status)::text = ANY (ARRAY[('PENDING'::character varying)::text, ('UPLOADED'::character varying)::text, ('APPROVED'::character varying)::text, ('REJECTED'::character varying)::text, ('WARNED_REJECTED'::character varying)::text, ('CRAWLED'::character varying)::text, ('MANUAL_REWARD'::character varying)::text, ('REWARDED'::character varying)::text, ('FAIL_TO_REWARD'::character varying)::text])))
);


ALTER TABLE public.user_posts_task OWNER TO meme_server;

--
-- Name: user_posts_task_record; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_posts_task_record (
    id bigint NOT NULL,
    task_id bigint NOT NULL,
    user_id bigint NOT NULL,
    create_time timestamp with time zone NOT NULL,
    task_name character varying(32) NOT NULL,
    task_status character varying(16) NOT NULL,
    post_url character varying(256),
    screenshot_key character varying(100),
    auditor character varying(16) DEFAULT NULL::character varying,
    audit_comment text,
    fail_reason text
);


ALTER TABLE public.user_posts_task_record OWNER TO meme_server;

--
-- Name: user_posts_task_record_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_posts_task_record_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_posts_task_record_id_seq OWNER TO meme_server;

--
-- Name: user_posts_task_record_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_posts_task_record_id_seq OWNED BY public.user_posts_task_record.id;


--
-- Name: user_posts_task_task_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_posts_task_task_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_posts_task_task_id_seq OWNER TO meme_server;

--
-- Name: user_posts_task_task_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_posts_task_task_id_seq OWNED BY public.user_posts_task.task_id;


--
-- Name: user_reward_record; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_reward_record (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    task_id bigint NOT NULL,
    create_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp with time zone,
    score integer,
    task_name character varying(20) NOT NULL,
    reward_name character varying(20) NOT NULL,
    reward_type character varying(16) NOT NULL,
    reward_reason character varying(32) NOT NULL,
    fail_reason text,
    CONSTRAINT user_reward_record_reward_name_check CHECK (((reward_name)::text = ANY ((ARRAY['POST'::character varying, 'CHECKIN'::character varying, 'WALLET_BINDING'::character varying])::text[]))),
    CONSTRAINT user_reward_record_reward_reason_check CHECK (((reward_reason)::text = ANY (ARRAY[('POSTS_UPLOADED'::character varying)::text, ('INVITE_REGISTER'::character varying)::text, ('POSTS_APPROVED'::character varying)::text, ('POSTS_APPRECIATED'::character varying)::text, ('POSTS_INFLUENCE'::character varying)::text]))),
    CONSTRAINT user_reward_record_reward_type_check CHECK (((reward_type)::text = ANY ((ARRAY['SCORE'::character varying, 'TANGIBLE'::character varying])::text[])))
);


ALTER TABLE public.user_reward_record OWNER TO meme_server;

--
-- Name: user_reward_record_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_reward_record_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_reward_record_id_seq OWNER TO meme_server;

--
-- Name: user_reward_record_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_reward_record_id_seq OWNED BY public.user_reward_record.id;


--
-- Name: user_score; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_score (
    user_id bigint NOT NULL,
    total_score integer NOT NULL,
    frozen_score integer NOT NULL,
    daily_obtain_limit integer DEFAULT 1000 NOT NULL,
    daily_obtained integer DEFAULT 0 NOT NULL,
    last_modified timestamp with time zone DEFAULT now() NOT NULL,
    version integer DEFAULT 0 NOT NULL,
    CONSTRAINT user_score_frozen_score_check CHECK ((frozen_score >= 0)),
    CONSTRAINT user_score_total_score_check CHECK ((total_score >= 0))
);


ALTER TABLE public.user_score OWNER TO meme_server;

--
-- Name: user_share_reward_record; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_share_reward_record (
    id bigint NOT NULL,
    beneficiary bigint NOT NULL,
    sharing_task bigint NOT NULL,
    create_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    score integer,
    contributor bigint NOT NULL,
    contributor_task bigint NOT NULL,
    contributor_task_name character varying(32) NOT NULL,
    contributor_name character varying(128),
    reward_type character varying(16) NOT NULL,
    reward_reason character varying(40) NOT NULL,
    fail_reason text,
    CONSTRAINT user_share_reward_record_reward_reason_check CHECK (((reward_reason)::text = ANY ((ARRAY['PASSIVE_INCOME_OF_POSTS_UPLOADED'::character varying, 'PASSIVE_INCOME_OF_POSTS_INFLUENCE'::character varying])::text[]))),
    CONSTRAINT user_share_reward_record_reward_type_check CHECK (((reward_type)::text = ANY ((ARRAY['SCORE'::character varying, 'TANGIBLE'::character varying])::text[])))
);


ALTER TABLE public.user_share_reward_record OWNER TO meme_server;

--
-- Name: user_share_reward_record_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_share_reward_record_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_share_reward_record_id_seq OWNER TO meme_server;

--
-- Name: user_share_reward_record_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_share_reward_record_id_seq OWNED BY public.user_share_reward_record.id;


--
-- Name: user_sharing_task; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_sharing_task (
    task_id bigint NOT NULL,
    user_id bigint NOT NULL,
    create_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp with time zone,
    counts integer DEFAULT 0,
    gained_score bigint DEFAULT 0
);


ALTER TABLE public.user_sharing_task OWNER TO meme_server;

--
-- Name: user_sharing_task_event_record; Type: TABLE; Schema: public; Owner: meme_server
--

CREATE TABLE public.user_sharing_task_event_record (
    id bigint NOT NULL,
    inviter bigint NOT NULL,
    sharing_task bigint NOT NULL,
    invitee bigint NOT NULL,
    invitee_task bigint,
    create_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    invitee_task_name character varying(32),
    task_status character varying(32) NOT NULL,
    fail_reason text
);


ALTER TABLE public.user_sharing_task_event_record OWNER TO meme_server;

--
-- Name: user_sharing_task_event_record_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_sharing_task_event_record_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_sharing_task_event_record_id_seq OWNER TO meme_server;

--
-- Name: user_sharing_task_event_record_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_sharing_task_event_record_id_seq OWNED BY public.user_sharing_task_event_record.id;


--
-- Name: user_sharing_task_task_id_seq; Type: SEQUENCE; Schema: public; Owner: meme_server
--

CREATE SEQUENCE public.user_sharing_task_task_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_sharing_task_task_id_seq OWNER TO meme_server;

--
-- Name: user_sharing_task_task_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: meme_server
--

ALTER SEQUENCE public.user_sharing_task_task_id_seq OWNED BY public.user_sharing_task.task_id;


--
-- Name: audit_comment_template id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.audit_comment_template ALTER COLUMN id SET DEFAULT nextval('public.audit_comment_template_id_seq'::regclass);


--
-- Name: audit_comment_template_multi_lang id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.audit_comment_template_multi_lang ALTER COLUMN id SET DEFAULT nextval('public.audit_comment_template_multi_lang_id_seq'::regclass);


--
-- Name: game_score_operation_log log_id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.game_score_operation_log ALTER COLUMN log_id SET DEFAULT nextval('public.game_score_operation_log_log_id_seq'::regclass);


--
-- Name: invitation_relation id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.invitation_relation ALTER COLUMN id SET DEFAULT nextval('public.invitation_relation_id_seq'::regclass);


--
-- Name: login_log id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.login_log ALTER COLUMN id SET DEFAULT nextval('public.login_log_id_seq'::regclass);


--
-- Name: mini_game_item id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.mini_game_item ALTER COLUMN id SET DEFAULT nextval('public.mini_game_item_id_seq'::regclass);


--
-- Name: oshit_task_config id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.oshit_task_config ALTER COLUMN id SET DEFAULT nextval('public.oshit_task_config_id_seq'::regclass);


--
-- Name: score_ledger ledger_id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.score_ledger ALTER COLUMN ledger_id SET DEFAULT nextval('public.score_ledger_ledger_id_seq'::regclass);


--
-- Name: score_operation_log log_id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.score_operation_log ALTER COLUMN log_id SET DEFAULT nextval('public.score_operation_log_log_id_seq'::regclass);


--
-- Name: t_activity_rank id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_activity_rank ALTER COLUMN id SET DEFAULT nextval('public.t_activity_rank_id_seq'::regclass);


--
-- Name: t_facebook_info id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_facebook_info ALTER COLUMN id SET DEFAULT nextval('public.t_facebook_info_id_seq'::regclass);


--
-- Name: t_global_daily_exchange_limit id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_global_daily_exchange_limit ALTER COLUMN id SET DEFAULT nextval('public.t_global_daily_exchange_limit_id_seq'::regclass);


--
-- Name: t_rank_activity_rule id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_rank_activity_rule ALTER COLUMN id SET DEFAULT nextval('public.t_rank_activity_rule_id_seq'::regclass);


--
-- Name: t_tweets_info id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_tweets_info ALTER COLUMN id SET DEFAULT nextval('public.t_tweets_info_id_seq'::regclass);


--
-- Name: t_user_daily_exchange_quota id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_user_daily_exchange_quota ALTER COLUMN id SET DEFAULT nextval('public.t_user_daily_exchange_quota_id_seq'::regclass);


--
-- Name: user_activity_message id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_activity_message ALTER COLUMN id SET DEFAULT nextval('public.user_activity_message_id_seq'::regclass);


--
-- Name: user_basic user_id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_basic ALTER COLUMN user_id SET DEFAULT nextval('public.user_basic_user_id_seq'::regclass);


--
-- Name: user_favorite_product id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_favorite_product ALTER COLUMN id SET DEFAULT nextval('public.user_favorite_product_id_seq'::regclass);


--
-- Name: user_game_item id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_game_item ALTER COLUMN id SET DEFAULT nextval('public.user_game_item_id_seq'::regclass);


--
-- Name: user_game_item_op_record id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_game_item_op_record ALTER COLUMN id SET DEFAULT nextval('public.user_game_item_op_record_id_seq'::regclass);


--
-- Name: user_invitation_enroll id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_invitation_enroll ALTER COLUMN id SET DEFAULT nextval('public.user_invitation_enroll_id_seq'::regclass);


--
-- Name: user_posts_task task_id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_posts_task ALTER COLUMN task_id SET DEFAULT nextval('public.user_posts_task_task_id_seq'::regclass);


--
-- Name: user_posts_task_record id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_posts_task_record ALTER COLUMN id SET DEFAULT nextval('public.user_posts_task_record_id_seq'::regclass);


--
-- Name: user_reward_record id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_reward_record ALTER COLUMN id SET DEFAULT nextval('public.user_reward_record_id_seq'::regclass);


--
-- Name: user_share_reward_record id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_share_reward_record ALTER COLUMN id SET DEFAULT nextval('public.user_share_reward_record_id_seq'::regclass);


--
-- Name: user_sharing_task task_id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_sharing_task ALTER COLUMN task_id SET DEFAULT nextval('public.user_sharing_task_task_id_seq'::regclass);


--
-- Name: user_sharing_task_event_record id; Type: DEFAULT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_sharing_task_event_record ALTER COLUMN id SET DEFAULT nextval('public.user_sharing_task_event_record_id_seq'::regclass);


--
-- Name: t_sol_game_buy_property_record SolGameBuyProperty_TxId; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_buy_property_record
    ADD CONSTRAINT "SolGameBuyProperty_TxId" UNIQUE ("Product", "TxId");


--
-- Name: t_sol_game_exchange_prize_record SolGameExchangePrizeRecord_TxId; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_exchange_prize_record
    ADD CONSTRAINT "SolGameExchangePrizeRecord_TxId" UNIQUE ("TxId");


--
-- Name: t_sol_game_property_info SolGamePropertyInfo_PropertyName; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_property_info
    ADD CONSTRAINT "SolGamePropertyInfo_PropertyName" UNIQUE ("Product", "PropertyName");


--
-- Name: t_sol_game_register_claim_record SolGameRegisterClaim_TxId; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_register_claim_record
    ADD CONSTRAINT "SolGameRegisterClaim_TxId" UNIQUE ("TxId");


--
-- Name: t_sol_game_role_info SolGameRoleInfo_Product_RoleName; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_role_info
    ADD CONSTRAINT "SolGameRoleInfo_Product_RoleName" UNIQUE ("Product", "RoleName");


--
-- Name: t_sol_game_unlock_role_record SolGameUnlockRecord_TxId; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_unlock_role_record
    ADD CONSTRAINT "SolGameUnlockRecord_TxId" UNIQUE ("Product", "TxId");


--
-- Name: t_sol_fee_statistics Sol_Fee_Slot_TxIndex; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_fee_statistics
    ADD CONSTRAINT "Sol_Fee_Slot_TxIndex" UNIQUE ("Slot", "TransactionIndex");


--
-- Name: t_sol_fee_statistics Sol_Fee_Statistics_TxId; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_fee_statistics
    ADD CONSTRAINT "Sol_Fee_Statistics_TxId" UNIQUE ("TransactionId");


--
-- Name: audit_comment_template audit_comment_template_comment_key; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.audit_comment_template
    ADD CONSTRAINT audit_comment_template_comment_key UNIQUE (comment);


--
-- Name: audit_comment_template_multi_lang audit_comment_template_multi_lang_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.audit_comment_template_multi_lang
    ADD CONSTRAINT audit_comment_template_multi_lang_pkey PRIMARY KEY (id);


--
-- Name: audit_comment_template audit_comment_template_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.audit_comment_template
    ADD CONSTRAINT audit_comment_template_pkey PRIMARY KEY (id);


--
-- Name: game_score_operation_log game_score_operation_log_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.game_score_operation_log
    ADD CONSTRAINT game_score_operation_log_pkey PRIMARY KEY (log_id);


--
-- Name: game_score game_score_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.game_score
    ADD CONSTRAINT game_score_pkey PRIMARY KEY (user_id);


--
-- Name: invitation_relation invitation_relation_invitee_key; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.invitation_relation
    ADD CONSTRAINT invitation_relation_invitee_key UNIQUE (invitee);


--
-- Name: invitation_relation invitation_relation_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.invitation_relation
    ADD CONSTRAINT invitation_relation_pkey PRIMARY KEY (id);


--
-- Name: login_log login_log_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.login_log
    ADD CONSTRAINT login_log_pkey PRIMARY KEY (id);


--
-- Name: mini_game_item mini_game_item_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.mini_game_item
    ADD CONSTRAINT mini_game_item_pkey PRIMARY KEY (id);


--
-- Name: oshit_task_config oshit_task_config_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.oshit_task_config
    ADD CONSTRAINT oshit_task_config_pkey PRIMARY KEY (id);


--
-- Name: t_reward_hacker reward_hacker_unique_native_account; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_hacker
    ADD CONSTRAINT reward_hacker_unique_native_account UNIQUE ("NativeAccount");


--
-- Name: score_ledger score_ledger_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.score_ledger
    ADD CONSTRAINT score_ledger_pkey PRIMARY KEY (ledger_id);


--
-- Name: score_operation_log score_operation_log_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.score_operation_log
    ADD CONSTRAINT score_operation_log_pkey PRIMARY KEY (log_id);


--
-- Name: social_media_user_info social_media_user_info_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.social_media_user_info
    ADD CONSTRAINT social_media_user_info_pkey PRIMARY KEY (user_id);


--
-- Name: t_activity_exchange_record t_activity_exchange_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_activity_exchange_record
    ADD CONSTRAINT t_activity_exchange_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_activity_exchange_rule t_activity_exchange_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_activity_exchange_rule
    ADD CONSTRAINT t_activity_exchange_rule_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_activity_rank t_activity_rank_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_activity_rank
    ADD CONSTRAINT t_activity_rank_pkey PRIMARY KEY (id);


--
-- Name: t_aws_config t_aws_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_aws_config
    ADD CONSTRAINT t_aws_config_pkey PRIMARY KEY ("AccessKeyId");


--
-- Name: t_campaign_key_config t_campaign_key_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_campaign_key_config
    ADD CONSTRAINT t_campaign_key_config_pkey PRIMARY KEY ("Chain");


--
-- Name: t_campaign_rpc_config t_campaign_rpc_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_campaign_rpc_config
    ADD CONSTRAINT t_campaign_rpc_config_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_campaign_scan_info t_campaign_scan_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_campaign_scan_info
    ADD CONSTRAINT t_campaign_scan_info_pkey PRIMARY KEY ("NativeAccount");


--
-- Name: t_campaign_tx t_campaign_tx_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_campaign_tx
    ADD CONSTRAINT t_campaign_tx_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_chain_config t_chain_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_chain_config
    ADD CONSTRAINT t_chain_config_pkey PRIMARY KEY ("Chain");


--
-- Name: t_daily_claim_stats t_daily_claim_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_daily_claim_stats
    ADD CONSTRAINT t_daily_claim_stats_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_debug_white_list t_debug_white_list_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_debug_white_list
    ADD CONSTRAINT t_debug_white_list_pkey PRIMARY KEY ("Id");


--
-- Name: t_ecommerce_order t_ecommerce_order_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_ecommerce_order
    ADD CONSTRAINT t_ecommerce_order_pkey PRIMARY KEY ("OrderId");


--
-- Name: t_facebook_info t_facebook_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_facebook_info
    ADD CONSTRAINT t_facebook_info_pkey PRIMARY KEY (id);


--
-- Name: t_facebook_info t_facebook_info_post_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_facebook_info
    ADD CONSTRAINT t_facebook_info_post_id_key UNIQUE (post_id);


--
-- Name: t_fee_tolerance t_fee_tolerance_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_fee_tolerance
    ADD CONSTRAINT t_fee_tolerance_pkey PRIMARY KEY ("Brand", "TokenSymbol");


--
-- Name: t_game_rpc_config t_game_rpc_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_game_rpc_config
    ADD CONSTRAINT t_game_rpc_config_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_game_scan_info t_game_scan_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_game_scan_info
    ADD CONSTRAINT t_game_scan_info_pkey PRIMARY KEY ("NativeAccount");


--
-- Name: t_game_tx t_game_tx_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_game_tx
    ADD CONSTRAINT t_game_tx_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_global_daily_exchange_limit t_global_daily_exchange_limit_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_global_daily_exchange_limit
    ADD CONSTRAINT t_global_daily_exchange_limit_pkey PRIMARY KEY (id);


--
-- Name: t_global_daily_exchange_limit t_global_daily_exchange_limit_quota_date_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_global_daily_exchange_limit
    ADD CONSTRAINT t_global_daily_exchange_limit_quota_date_key UNIQUE (quota_date);


--
-- Name: t_helius_api t_helius_api_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_helius_api
    ADD CONSTRAINT t_helius_api_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_helius_scan_info t_helius_scan_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_helius_scan_info
    ADD CONSTRAINT t_helius_scan_info_pkey PRIMARY KEY ("Service");


--
-- Name: t_micro_service_lb t_micro_service_lb_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_micro_service_lb
    ADD CONSTRAINT t_micro_service_lb_pkey PRIMARY KEY ("Ip");


--
-- Name: t_pos_reward_key_config t_pos_reward_key_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_pos_reward_key_config
    ADD CONSTRAINT t_pos_reward_key_config_pkey PRIMARY KEY ("Service");


--
-- Name: t_pos_rpc_config t_pos_rpc_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_pos_rpc_config
    ADD CONSTRAINT t_pos_rpc_config_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_pos_scan_info t_pos_scan_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_pos_scan_info
    ADD CONSTRAINT t_pos_scan_info_pkey PRIMARY KEY ("Service");


--
-- Name: t_pos_tx t_pos_tx_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_pos_tx
    ADD CONSTRAINT t_pos_tx_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_rank_activity_rule t_rank_activity_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_rank_activity_rule
    ADD CONSTRAINT t_rank_activity_rule_pkey PRIMARY KEY (id);


--
-- Name: t_rank_activity_rule t_rank_activity_rule_rule_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_rank_activity_rule
    ADD CONSTRAINT t_rank_activity_rule_rule_code_key UNIQUE (rule_code);


--
-- Name: t_reward_code_fee t_reward_code_fee_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_code_fee
    ADD CONSTRAINT t_reward_code_fee_pkey PRIMARY KEY ("Amount");


--
-- Name: t_reward_code_key_config t_reward_code_key_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_code_key_config
    ADD CONSTRAINT t_reward_code_key_config_pkey PRIMARY KEY ("Chain");


--
-- Name: t_reward_code t_reward_code_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_code
    ADD CONSTRAINT t_reward_code_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_reward_code_rule t_reward_code_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_code_rule
    ADD CONSTRAINT t_reward_code_rule_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_reward_discount_rate t_reward_discount_rate_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_discount_rate
    ADD CONSTRAINT t_reward_discount_rate_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_reward_exclude t_reward_exclude_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_exclude
    ADD CONSTRAINT t_reward_exclude_pkey PRIMARY KEY ("Chain", "Account");


--
-- Name: t_reward_hacker t_reward_hacker_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_hacker
    ADD CONSTRAINT t_reward_hacker_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_reward_key_config t_reward_key_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_key_config
    ADD CONSTRAINT t_reward_key_config_pkey PRIMARY KEY ("Service");


--
-- Name: t_reward_lottery_claim_record t_reward_lottery_claim_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_lottery_claim_record
    ADD CONSTRAINT t_reward_lottery_claim_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_reward_lottery t_reward_lottery_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_lottery
    ADD CONSTRAINT t_reward_lottery_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_reward_private_key t_reward_private_key_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_private_key
    ADD CONSTRAINT t_reward_private_key_pkey PRIMARY KEY ("ServiceType");


--
-- Name: t_reward_rpc_config t_reward_rpc_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_rpc_config
    ADD CONSTRAINT t_reward_rpc_config_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_reward_scan_info t_reward_scan_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_scan_info
    ADD CONSTRAINT t_reward_scan_info_pkey PRIMARY KEY ("NativeAccount");


--
-- Name: t_reward_test t_reward_test_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_test
    ADD CONSTRAINT t_reward_test_pkey PRIMARY KEY ("Chain", "NativeAccount");


--
-- Name: t_reward_tx t_reward_tx_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_reward_tx
    ADD CONSTRAINT t_reward_tx_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_shit_house t_shit_house_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_shit_house
    ADD CONSTRAINT t_shit_house_pkey PRIMARY KEY ("HouseId");


--
-- Name: t_shit_house_token_quota t_shit_house_token_quota_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_shit_house_token_quota
    ADD CONSTRAINT t_shit_house_token_quota_pkey PRIMARY KEY ("HouseId");


--
-- Name: t_sol_airdrop_record t_sol_airdrop_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_airdrop_record
    ADD CONSTRAINT t_sol_airdrop_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_campaign_exchange_score_rule t_sol_campaign_exchange_score_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_campaign_exchange_score_rule
    ADD CONSTRAINT t_sol_campaign_exchange_score_rule_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_determine_invite_record t_sol_determine_invite_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_determine_invite_record
    ADD CONSTRAINT t_sol_determine_invite_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_exchange_campaign_score_to_token_record t_sol_exchange_campaign_score_to_token_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_exchange_campaign_score_to_token_record
    ADD CONSTRAINT t_sol_exchange_campaign_score_to_token_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_fee_statistics t_sol_fee_statistics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_fee_statistics
    ADD CONSTRAINT t_sol_fee_statistics_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_fund_flow t_sol_fund_flow_new_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_fund_flow
    ADD CONSTRAINT t_sol_fund_flow_new_pkey PRIMARY KEY (record_id);


--
-- Name: t_sol_fund_flow_old t_sol_fund_flow_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_fund_flow_old
    ADD CONSTRAINT t_sol_fund_flow_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_game_buy_property_record t_sol_game_buy_property_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_buy_property_record
    ADD CONSTRAINT t_sol_game_buy_property_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_game_exchange_prize_record t_sol_game_exchange_prize_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_exchange_prize_record
    ADD CONSTRAINT t_sol_game_exchange_prize_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_game_pack_account t_sol_game_pack_account_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_pack_account
    ADD CONSTRAINT t_sol_game_pack_account_pkey PRIMARY KEY ("Product", "UserId", "PackId");


--
-- Name: t_sol_game_prize_account t_sol_game_prize_account_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_prize_account
    ADD CONSTRAINT t_sol_game_prize_account_pkey PRIMARY KEY ("Product", "UserId", "PackId", "PrizeId");


--
-- Name: t_sol_game_prize_info t_sol_game_prize_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_prize_info
    ADD CONSTRAINT t_sol_game_prize_info_pkey PRIMARY KEY ("Product", "PrizeId");


--
-- Name: t_sol_game_property_account t_sol_game_property_account_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_property_account
    ADD CONSTRAINT t_sol_game_property_account_pkey PRIMARY KEY ("Product", "UserId", "PropertyId");


--
-- Name: t_sol_game_property_info t_sol_game_property_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_property_info
    ADD CONSTRAINT t_sol_game_property_info_pkey PRIMARY KEY ("Product", "PropertyId");


--
-- Name: t_sol_game_property_record t_sol_game_property_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_property_record
    ADD CONSTRAINT t_sol_game_property_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_game_rank_top t_sol_game_rank_top_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_rank_top
    ADD CONSTRAINT t_sol_game_rank_top_pkey PRIMARY KEY ("Product", "UserId");


--
-- Name: t_sol_game_receipt_info t_sol_game_receipt_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_receipt_info
    ADD CONSTRAINT t_sol_game_receipt_info_pkey PRIMARY KEY ("Product", "Brand", "TokenSymbol");


--
-- Name: t_sol_game_register_claim_record t_sol_game_register_claim_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_register_claim_record
    ADD CONSTRAINT t_sol_game_register_claim_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_game_register_info t_sol_game_register_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_register_info
    ADD CONSTRAINT t_sol_game_register_info_pkey PRIMARY KEY ("Product", "UserId");


--
-- Name: t_sol_game_role_account t_sol_game_role_account_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_role_account
    ADD CONSTRAINT t_sol_game_role_account_pkey PRIMARY KEY ("Product", "UserId", "RoleId");


--
-- Name: t_sol_game_role_info t_sol_game_role_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_role_info
    ADD CONSTRAINT t_sol_game_role_info_pkey PRIMARY KEY ("Product", "RoleId");


--
-- Name: t_sol_game_unlock_role_record t_sol_game_unlock_role_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_unlock_role_record
    ADD CONSTRAINT t_sol_game_unlock_role_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_game_vote_record t_sol_game_vote_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_game_vote_record
    ADD CONSTRAINT t_sol_game_vote_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_native_account_info t_sol_native_account_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_native_account_info
    ADD CONSTRAINT t_sol_native_account_info_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_official_give_token_record t_sol_official_give_token_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_official_give_token_record
    ADD CONSTRAINT t_sol_official_give_token_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_official_give_token_reward_record t_sol_official_give_token_reward_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_official_give_token_reward_record
    ADD CONSTRAINT t_sol_official_give_token_reward_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_official_give_token_reward_rule t_sol_official_give_token_reward_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_official_give_token_reward_rule
    ADD CONSTRAINT t_sol_official_give_token_reward_rule_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_official_transfer_token_reward_record t_sol_official_transfer_token_reward_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_official_transfer_token_reward_record
    ADD CONSTRAINT t_sol_official_transfer_token_reward_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_official_transfer_token_reward_rule t_sol_official_transfer_token_reward_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_official_transfer_token_reward_rule
    ADD CONSTRAINT t_sol_official_transfer_token_reward_rule_pkey PRIMARY KEY ("Brand", "TokenSymbol");


--
-- Name: t_sol_pos_mission_config t_sol_pos_mission_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_mission_config
    ADD CONSTRAINT t_sol_pos_mission_config_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_pos_retweet_config t_sol_pos_retweet_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_retweet_config
    ADD CONSTRAINT t_sol_pos_retweet_config_pkey PRIMARY KEY ("UserID", "RetweetId");


--
-- Name: t_sol_pos_reward_claim_record t_sol_pos_reward_claim_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_reward_claim_record
    ADD CONSTRAINT t_sol_pos_reward_claim_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_pos_reward t_sol_pos_reward_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_reward
    ADD CONSTRAINT t_sol_pos_reward_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_pos_reward_rule t_sol_pos_reward_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_reward_rule
    ADD CONSTRAINT t_sol_pos_reward_rule_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_pos_snap_shot t_sol_pos_snap_shot_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_snap_shot
    ADD CONSTRAINT t_sol_pos_snap_shot_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_pos_social_media_mission_record t_sol_pos_social_media_mission_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_social_media_mission_record
    ADD CONSTRAINT t_sol_pos_social_media_mission_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_pos_star_level_config t_sol_pos_star_level_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_star_level_config
    ADD CONSTRAINT t_sol_pos_star_level_config_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_pos_star_level_rule t_sol_pos_star_level_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_star_level_rule
    ADD CONSTRAINT t_sol_pos_star_level_rule_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_pos_twitter_oauth2_config t_sol_pos_twitter_oauth2_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_pos_twitter_oauth2_config
    ADD CONSTRAINT t_sol_pos_twitter_oauth2_config_pkey PRIMARY KEY ("UserID");


--
-- Name: t_sol_qn_fee t_sol_qn_fee_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_qn_fee
    ADD CONSTRAINT t_sol_qn_fee_pkey PRIMARY KEY ("Id");


--
-- Name: t_sol_scan_info t_sol_scan_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_scan_info
    ADD CONSTRAINT t_sol_scan_info_pkey PRIMARY KEY ("Brand", "TokenSymbol");


--
-- Name: t_sol_stake_fix_interest_config t_sol_stake_fix_interest_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_stake_fix_interest_config
    ADD CONSTRAINT t_sol_stake_fix_interest_config_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_stake_invite_rate t_sol_stake_invite_rate_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_stake_invite_rate
    ADD CONSTRAINT t_sol_stake_invite_rate_pkey PRIMARY KEY ("Level");


--
-- Name: t_sol_stake_reward_claim_record t_sol_stake_reward_claim_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_stake_reward_claim_record
    ADD CONSTRAINT t_sol_stake_reward_claim_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_stake_reward t_sol_stake_reward_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_stake_reward
    ADD CONSTRAINT t_sol_stake_reward_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_stake_reward_rule t_sol_stake_reward_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_stake_reward_rule
    ADD CONSTRAINT t_sol_stake_reward_rule_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_stake_snap_shot t_sol_stake_snap_shot_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_stake_snap_shot
    ADD CONSTRAINT t_sol_stake_snap_shot_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_stake_star_level_config t_sol_stake_star_level_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_stake_star_level_config
    ADD CONSTRAINT t_sol_stake_star_level_config_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_stake_star_level_rule t_sol_stake_star_level_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_stake_star_level_rule
    ADD CONSTRAINT t_sol_stake_star_level_rule_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_swap_new_token_config t_sol_swap_new_token_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_swap_new_token_config
    ADD CONSTRAINT t_sol_swap_new_token_config_pkey PRIMARY KEY ("Brand", "TokenSymbol");


--
-- Name: t_sol_swap_token_record t_sol_swap_token_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_swap_token_record
    ADD CONSTRAINT t_sol_swap_token_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_swap_token_rule t_sol_swap_token_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_swap_token_rule
    ADD CONSTRAINT t_sol_swap_token_rule_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_token_config t_sol_token_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_token_config
    ADD CONSTRAINT t_sol_token_config_pkey PRIMARY KEY ("Brand", "TokenSymbol");


--
-- Name: t_sol_transfer_checked_record t_sol_transfer_checked_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_transfer_checked_record
    ADD CONSTRAINT t_sol_transfer_checked_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_transfer_reward_claim t_sol_transfer_reward_claim_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_transfer_reward_claim
    ADD CONSTRAINT t_sol_transfer_reward_claim_pkey PRIMARY KEY ("Brand", "TokenSymbol", "Level");


--
-- Name: t_sol_transfer_reward_distribution t_sol_transfer_reward_distribution_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_transfer_reward_distribution
    ADD CONSTRAINT t_sol_transfer_reward_distribution_pkey PRIMARY KEY ("Brand", "TokenSymbol", "Level");


--
-- Name: t_sol_transfer_reward_record t_sol_transfer_reward_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_transfer_reward_record
    ADD CONSTRAINT t_sol_transfer_reward_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_transfer_to_dex_record t_sol_transfer_to_dex_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_transfer_to_dex_record
    ADD CONSTRAINT t_sol_transfer_to_dex_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_transfer_token_reward_rule t_sol_transfer_token_reward_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_transfer_token_reward_rule
    ADD CONSTRAINT t_sol_transfer_token_reward_rule_pkey PRIMARY KEY ("Brand", "TokenSymbol");


--
-- Name: t_stake_amm_config t_stake_amm_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_stake_amm_config
    ADD CONSTRAINT t_stake_amm_config_pkey PRIMARY KEY ("QuoteToken");


--
-- Name: t_stake_area_leader t_stake_area_leader_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_stake_area_leader
    ADD CONSTRAINT t_stake_area_leader_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_stake_buy_token t_stake_buy_token_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_stake_buy_token
    ADD CONSTRAINT t_stake_buy_token_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_stake_buy_token t_stake_buy_token_txid_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_stake_buy_token
    ADD CONSTRAINT t_stake_buy_token_txid_unique UNIQUE ("TxId");


--
-- Name: t_stake_record t_stake_record_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_stake_record
    ADD CONSTRAINT t_stake_record_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_stake_token_pool t_stake_token_pool_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_stake_token_pool
    ADD CONSTRAINT t_stake_token_pool_pkey PRIMARY KEY ("Source");


--
-- Name: t_stake_total_area_leader t_stake_total_area_leader_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_stake_total_area_leader
    ADD CONSTRAINT t_stake_total_area_leader_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_swap_key_config t_swap_key_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_swap_key_config
    ADD CONSTRAINT t_swap_key_config_pkey PRIMARY KEY ("Chain");


--
-- Name: t_swap_scan_info t_swap_scan_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_swap_scan_info
    ADD CONSTRAINT t_swap_scan_info_pkey PRIMARY KEY ("NativeAccount");


--
-- Name: t_tweets_info t_tweets_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_tweets_info
    ADD CONSTRAINT t_tweets_info_pkey PRIMARY KEY (id);


--
-- Name: t_tweets_info t_tweets_info_tweet_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_tweets_info
    ADD CONSTRAINT t_tweets_info_tweet_id_key UNIQUE (tweet_id);


--
-- Name: t_user_daily_exchange_quota t_user_daily_exchange_quota_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_user_daily_exchange_quota
    ADD CONSTRAINT t_user_daily_exchange_quota_pkey PRIMARY KEY (id);


--
-- Name: t_user_daily_exchange_quota t_user_daily_exchange_quota_user_id_quota_date_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_user_daily_exchange_quota
    ADD CONSTRAINT t_user_daily_exchange_quota_user_id_quota_date_key UNIQUE (user_id, quota_date);


--
-- Name: t_user_wallet_rpc_config t_user_wallet_rpc_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_user_wallet_rpc_config
    ADD CONSTRAINT t_user_wallet_rpc_config_pkey PRIMARY KEY ("RecordId");


--
-- Name: t_sol_fund_flow_old uniq_txid_toaccount_flowtype; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_fund_flow_old
    ADD CONSTRAINT uniq_txid_toaccount_flowtype UNIQUE ("TxId", "ToNativeAccount", "FlowType");


--
-- Name: t_sol_official_give_token_reward_rule unique_brand_default; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_official_give_token_reward_rule
    ADD CONSTRAINT unique_brand_default UNIQUE ("Brand", "TokenSymbol", "Default");


--
-- Name: t_sol_official_give_token_reward_rule unique_brand_invite_code; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_official_give_token_reward_rule
    ADD CONSTRAINT unique_brand_invite_code UNIQUE ("Brand", "TokenSymbol", "InviteCode");


--
-- Name: t_sol_native_account_info unique_brand_token_account; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_native_account_info
    ADD CONSTRAINT unique_brand_token_account UNIQUE ("Brand", "TokenSymbol", "NativeAccount");


--
-- Name: t_sol_native_account_info unique_invite_code; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_native_account_info
    ADD CONSTRAINT unique_invite_code UNIQUE ("InviteCode");


--
-- Name: t_activity_rank unique_name_user_rank; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_activity_rank
    ADD CONSTRAINT unique_name_user_rank UNIQUE (name, user_code, rank_platform);


--
-- Name: t_daily_claim_stats unique_nativeaccount_claimdate; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_daily_claim_stats
    ADD CONSTRAINT unique_nativeaccount_claimdate UNIQUE ("NativeAccount", "TakeShitDate");


--
-- Name: social_media_user_info unique_provider_user_id; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.social_media_user_info
    ADD CONSTRAINT unique_provider_user_id UNIQUE (provider_user_id, provider);


--
-- Name: t_activity_exchange_rule unique_rank_platform; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_activity_exchange_rule
    ADD CONSTRAINT unique_rank_platform UNIQUE ("RankPlatform");


--
-- Name: t_sol_transfer_checked_record unique_transfer_info; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_sol_transfer_checked_record
    ADD CONSTRAINT unique_transfer_info UNIQUE ("TransferTxId", "InstructionIndex");


--
-- Name: t_activity_exchange_record unique_tx_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_activity_exchange_record
    ADD CONSTRAINT unique_tx_id UNIQUE ("TxId");


--
-- Name: user_activity_message user_activity_message_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_activity_message
    ADD CONSTRAINT user_activity_message_pkey PRIMARY KEY (id);


--
-- Name: user_basic user_basic_email_key; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_basic
    ADD CONSTRAINT user_basic_email_key UNIQUE (email);


--
-- Name: user_basic user_basic_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_basic
    ADD CONSTRAINT user_basic_pkey PRIMARY KEY (user_id);


--
-- Name: user_basic user_basic_username_key; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_basic
    ADD CONSTRAINT user_basic_username_key UNIQUE (username);


--
-- Name: user_extend_info user_extend_info_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_extend_info
    ADD CONSTRAINT user_extend_info_pkey PRIMARY KEY (user_id);


--
-- Name: user_favorite_product user_favorite_product_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_favorite_product
    ADD CONSTRAINT user_favorite_product_pkey PRIMARY KEY (id);


--
-- Name: user_game_item_op_record user_game_item_op_record_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_game_item_op_record
    ADD CONSTRAINT user_game_item_op_record_pkey PRIMARY KEY (id);


--
-- Name: user_game_item user_game_item_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_game_item
    ADD CONSTRAINT user_game_item_pkey PRIMARY KEY (id);


--
-- Name: user_invitation_enroll user_invitation_enroll_invite_code_key; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_invitation_enroll
    ADD CONSTRAINT user_invitation_enroll_invite_code_key UNIQUE (invite_code);


--
-- Name: user_invitation_enroll user_invitation_enroll_inviter_key; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_invitation_enroll
    ADD CONSTRAINT user_invitation_enroll_inviter_key UNIQUE (inviter);


--
-- Name: user_invitation_enroll user_invitation_enroll_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_invitation_enroll
    ADD CONSTRAINT user_invitation_enroll_pkey PRIMARY KEY (id);


--
-- Name: user_posts_task user_posts_task_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_posts_task
    ADD CONSTRAINT user_posts_task_pkey PRIMARY KEY (task_id);


--
-- Name: user_posts_task_record user_posts_task_record_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_posts_task_record
    ADD CONSTRAINT user_posts_task_record_pkey PRIMARY KEY (id);


--
-- Name: user_reward_record user_reward_record_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_reward_record
    ADD CONSTRAINT user_reward_record_pkey PRIMARY KEY (id);


--
-- Name: user_score user_score_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_score
    ADD CONSTRAINT user_score_pkey PRIMARY KEY (user_id);


--
-- Name: user_share_reward_record user_share_reward_record_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_share_reward_record
    ADD CONSTRAINT user_share_reward_record_pkey PRIMARY KEY (id);


--
-- Name: user_sharing_task_event_record user_sharing_task_event_record_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_sharing_task_event_record
    ADD CONSTRAINT user_sharing_task_event_record_pkey PRIMARY KEY (id);


--
-- Name: user_sharing_task user_sharing_task_pkey; Type: CONSTRAINT; Schema: public; Owner: meme_server
--

ALTER TABLE ONLY public.user_sharing_task
    ADD CONSTRAINT user_sharing_task_pkey PRIMARY KEY (task_id);


--
-- Name: SolGameRoleAccount_Index_RoleId; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "SolGameRoleAccount_Index_RoleId" ON public.t_sol_game_role_account USING btree ("RoleId");


--
-- Name: idx_activity_rank_score_desc; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_activity_rank_score_desc ON public.t_activity_rank USING btree (rank_score DESC, handle, user_code);


--
-- Name: idx_create_time; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_create_time ON public.login_log USING btree (create_time);


--
-- Name: idx_facebook_info_expired_statistics_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_facebook_info_expired_statistics_name ON public.t_facebook_info USING btree (expired, statistics, username);


--
-- Name: idx_facebook_info_post_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_facebook_info_post_id ON public.t_facebook_info USING btree (post_id);


--
-- Name: idx_facebook_info_score_id_handle_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_facebook_info_score_id_handle_time ON public.t_facebook_info USING btree (score DESC, id, profile_id, push_time);


--
-- Name: idx_fund_flow_no_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_fund_flow_no_time ON public.t_sol_fund_flow_old USING btree ("ToNativeAccount", "ServiceType", "FlowType") INCLUDE ("Amount");


--
-- Name: idx_gsol_created_at; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_gsol_created_at ON public.game_score_operation_log USING btree (created_at);


--
-- Name: idx_gsol_trans_id; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_gsol_trans_id ON public.game_score_operation_log USING btree (transaction_id);


--
-- Name: idx_gsol_user_id; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_gsol_user_id ON public.game_score_operation_log USING btree (user_id, op_type);


--
-- Name: idx_invite_by; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_invite_by ON public.user_invitation_enroll USING btree (inviter_invite_by);


--
-- Name: idx_invite_code; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_invite_code ON public.user_invitation_enroll USING btree (invite_code);


--
-- Name: idx_inviter; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_inviter ON public.user_invitation_enroll USING btree (inviter);


--
-- Name: idx_login_time; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_login_time ON public.login_log USING btree (user_id, login_time);


--
-- Name: idx_provider_info; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_provider_info ON public.login_log USING btree (provider_user_id, provider);


--
-- Name: idx_score_ledger_create_ts; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_score_ledger_create_ts ON public.score_ledger USING btree (create_ts);


--
-- Name: idx_sol_trans_id; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_sol_trans_id ON public.score_operation_log USING btree (transaction_id);


--
-- Name: idx_sol_transfer_checked_from_state_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_sol_transfer_checked_from_state_time ON public.t_sol_transfer_checked_record USING btree ("FromNativeAccount", "State", "CreateTime") INCLUDE ("TransferAmount");


--
-- Name: idx_sol_user_id; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_sol_user_id ON public.score_operation_log USING btree (user_id, op_type);


--
-- Name: idx_stake_buy_token_locked; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stake_buy_token_locked ON public.t_stake_buy_token USING btree ("Locked");


--
-- Name: idx_stake_buy_token_locked_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stake_buy_token_locked_at ON public.t_stake_buy_token USING btree ("LockedAt");


--
-- Name: idx_stake_buy_token_locked_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stake_buy_token_locked_by ON public.t_stake_buy_token USING btree ("LockedBy");


--
-- Name: idx_stake_buy_token_slot; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_stake_buy_token_slot ON public.t_stake_buy_token USING btree ("Slot");


--
-- Name: idx_t_sol_fund_flow_covering; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_t_sol_fund_flow_covering ON public.t_sol_fund_flow_old USING btree ("ToNativeAccount", "ServiceType", "FlowType", "CreateTime") INCLUDE ("Amount");


--
-- Name: idx_t_sol_fund_flow_new_covering; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_t_sol_fund_flow_new_covering ON public.t_sol_fund_flow USING btree (to_native_account, service_type, flow_type, create_time) INCLUDE (amount);


--
-- Name: idx_t_sol_fund_flow_new_unique_tx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_t_sol_fund_flow_new_unique_tx ON public.t_sol_fund_flow USING btree (tx_id, to_native_account, flow_type);


--
-- Name: idx_toam_user_record; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_toam_user_record ON public.user_activity_message USING btree (user_id, message_type, row_id);


--
-- Name: idx_tweets_info_expired_statistics_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tweets_info_expired_statistics_name ON public.t_tweets_info USING btree (expired, statistics, name);


--
-- Name: idx_tweets_info_score_id_handle_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tweets_info_score_id_handle_time ON public.t_tweets_info USING btree (score DESC, id, handle, push_time);


--
-- Name: idx_tweets_info_tweet_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tweets_info_tweet_id ON public.t_tweets_info USING btree (tweet_id);


--
-- Name: idx_uei_avatar_picture; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uei_avatar_picture ON public.user_extend_info USING btree (avatar_picture);


--
-- Name: idx_uei_birthday; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uei_birthday ON public.user_extend_info USING btree (avatar_picture);


--
-- Name: idx_uei_phone_number; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uei_phone_number ON public.user_extend_info USING btree (phone_number, region_code);


--
-- Name: idx_ufp_user_product; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE UNIQUE INDEX idx_ufp_user_product ON public.user_favorite_product USING btree (user_id, product_id);


--
-- Name: idx_ugi_user_item; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE UNIQUE INDEX idx_ugi_user_item ON public.user_game_item USING btree (user_id, item_id);


--
-- Name: idx_ugior_score_log; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_ugior_score_log ON public.user_game_item_op_record USING btree (game_score_log_id);


--
-- Name: idx_ugior_user_item; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_ugior_user_item ON public.user_game_item_op_record USING btree (user_id, item_id);


--
-- Name: idx_update_time; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_update_time ON public.login_log USING btree (update_time);


--
-- Name: idx_upt_creat_time; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_upt_creat_time ON public.user_posts_task USING btree (create_time);


--
-- Name: idx_upt_screenshot; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE UNIQUE INDEX idx_upt_screenshot ON public.user_posts_task USING btree (screenshot_key);


--
-- Name: idx_upt_task_name; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_upt_task_name ON public.oshit_task_config USING btree (task_name, task_type);


--
-- Name: idx_upt_task_type; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_upt_task_type ON public.oshit_task_config USING btree (task_type, task_name);


--
-- Name: idx_upt_url; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE UNIQUE INDEX idx_upt_url ON public.user_posts_task USING btree (post_url);


--
-- Name: idx_upt_user_id; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_upt_user_id ON public.user_posts_task USING btree (user_id, create_time);


--
-- Name: idx_uptr_create_time; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uptr_create_time ON public.user_posts_task_record USING btree (create_time);


--
-- Name: idx_uptr_url; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uptr_url ON public.user_posts_task_record USING btree (post_url);


--
-- Name: idx_uptr_user_id; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uptr_user_id ON public.user_posts_task_record USING btree (user_id);


--
-- Name: idx_urr_task_id; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_urr_task_id ON public.user_reward_record USING btree (task_id);


--
-- Name: idx_urr_user_id; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_urr_user_id ON public.user_reward_record USING btree (user_id);


--
-- Name: idx_user_basic_email; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE UNIQUE INDEX idx_user_basic_email ON public.user_basic USING btree (email);


--
-- Name: idx_user_basic_nick_name; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_user_basic_nick_name ON public.user_basic USING btree (nick_name);


--
-- Name: idx_user_basic_provider_user; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE UNIQUE INDEX idx_user_basic_provider_user ON public.user_basic USING btree (provider_user_id, provider);


--
-- Name: idx_user_login_time; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_user_login_time ON public.login_log USING btree (user_id, login_time);


--
-- Name: idx_usrr_beneficiary; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_usrr_beneficiary ON public.user_share_reward_record USING btree (beneficiary);


--
-- Name: idx_usrr_contributor; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_usrr_contributor ON public.user_share_reward_record USING btree (contributor);


--
-- Name: idx_usrr_sharing_contributor_task; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_usrr_sharing_contributor_task ON public.user_share_reward_record USING btree (contributor_task);


--
-- Name: idx_usrr_sharing_task; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_usrr_sharing_task ON public.user_share_reward_record USING btree (sharing_task);


--
-- Name: idx_ust_user_id; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_ust_user_id ON public.user_sharing_task USING btree (user_id);


--
-- Name: idx_uster_invitee; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uster_invitee ON public.user_sharing_task_event_record USING btree (invitee);


--
-- Name: idx_uster_invitee_task_name; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uster_invitee_task_name ON public.user_sharing_task_event_record USING btree (invitee_task_name);


--
-- Name: idx_uster_inviter; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uster_inviter ON public.user_sharing_task_event_record USING btree (inviter);


--
-- Name: idx_uster_sharing_task; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX idx_uster_sharing_task ON public.user_sharing_task_event_record USING btree (sharing_task);


--
-- Name: invite_relation_idx_invite_code; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX invite_relation_idx_invite_code ON public.invitation_relation USING btree (invite_code);


--
-- Name: invite_relation_idx_invitee; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE UNIQUE INDEX invite_relation_idx_invitee ON public.invitation_relation USING btree (invitee);


--
-- Name: invite_relation_idx_inviter; Type: INDEX; Schema: public; Owner: meme_server
--

CREATE INDEX invite_relation_idx_inviter ON public.invitation_relation USING btree (inviter);


--
-- Name: t_activity_exchange_record_RankPlatform_Handle_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_activity_exchange_record_RankPlatform_Handle_idx" ON public.t_activity_exchange_record USING btree ("RankPlatform", "Handle");


--
-- Name: t_campaign_scan_info_TokenAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_campaign_scan_info_TokenAccount_idx" ON public.t_campaign_scan_info USING btree ("TokenAccount");


--
-- Name: t_campaign_tx_TxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_campaign_tx_TxId_idx" ON public.t_campaign_tx USING btree ("TxId");


--
-- Name: t_debug_white_list_NativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_debug_white_list_NativeAccount_idx" ON public.t_debug_white_list USING btree ("NativeAccount");


--
-- Name: t_game_scan_info_TokenAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_game_scan_info_TokenAccount_idx" ON public.t_game_scan_info USING btree ("TokenAccount");


--
-- Name: t_game_tx_TxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_game_tx_TxId_idx" ON public.t_game_tx USING btree ("TxId");


--
-- Name: t_pos_tx_TxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_pos_tx_TxId_idx" ON public.t_pos_tx USING btree ("TxId");


--
-- Name: t_reward_code_ExpireTime_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_reward_code_ExpireTime_idx" ON public.t_reward_code USING btree ("ExpireTime");


--
-- Name: t_reward_code_RewardCode_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_reward_code_RewardCode_idx" ON public.t_reward_code USING btree ("RewardCode");


--
-- Name: t_reward_lottery_NativeAccount_Day_RewardType_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_reward_lottery_NativeAccount_Day_RewardType_idx" ON public.t_reward_lottery USING btree ("NativeAccount", "Day", "RewardType");


--
-- Name: t_reward_lottery_NativeAccount_State_Pending_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_reward_lottery_NativeAccount_State_Pending_idx" ON public.t_reward_lottery USING btree ("NativeAccount", "State", "Pending");


--
-- Name: t_reward_lottery_claim_record_State_CreateTime_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_reward_lottery_claim_record_State_CreateTime_idx" ON public.t_reward_lottery_claim_record USING btree ("State", "CreateTime");


--
-- Name: t_reward_lottery_claim_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_reward_lottery_claim_record_State_LastValidBlockHeight_idx" ON public.t_reward_lottery_claim_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_reward_lottery_claim_record_TxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_reward_lottery_claim_record_TxId_idx" ON public.t_reward_lottery_claim_record USING btree ("TxId");


--
-- Name: t_reward_scan_info_TokenAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_reward_scan_info_TokenAccount_idx" ON public.t_reward_scan_info USING btree ("TokenAccount");


--
-- Name: t_reward_tx_TxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_reward_tx_TxId_idx" ON public.t_reward_tx USING btree ("TxId");


--
-- Name: t_shit_house_token_quota_Batch_HouseId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_shit_house_token_quota_Batch_HouseId_idx" ON public.t_shit_house_token_quota USING btree ("Batch", "HouseId");


--
-- Name: t_sol_determine_invite_record_Brand_TokenSymbol_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_determine_invite_record_Brand_TokenSymbol_idx" ON public.t_sol_determine_invite_record USING btree ("Brand", "TokenSymbol");


--
-- Name: t_sol_determine_invite_record_InviteeNativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_determine_invite_record_InviteeNativeAccount_idx" ON public.t_sol_determine_invite_record USING btree ("InviteeNativeAccount");


--
-- Name: t_sol_determine_invite_record_InviterNativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_determine_invite_record_InviterNativeAccount_idx" ON public.t_sol_determine_invite_record USING btree ("InviterNativeAccount");


--
-- Name: t_sol_determine_invite_record_InviterTokenAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_determine_invite_record_InviterTokenAccount_idx" ON public.t_sol_determine_invite_record USING btree ("InviterTokenAccount");


--
-- Name: t_sol_exchange_campaign_score_to_token_ReceiptNativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_exchange_campaign_score_to_token_ReceiptNativeAccount_idx" ON public.t_sol_exchange_campaign_score_to_token_record USING btree ("ReceiptNativeAccount");


--
-- Name: t_sol_exchange_campaign_score_to_token_reco_Provider_UserId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_exchange_campaign_score_to_token_reco_Provider_UserId_idx" ON public.t_sol_exchange_campaign_score_to_token_record USING btree ("Provider", "UserId");


--
-- Name: t_sol_exchange_campaign_score_to_token_record_ExchangeTxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_exchange_campaign_score_to_token_record_ExchangeTxId_idx" ON public.t_sol_exchange_campaign_score_to_token_record USING btree ("ExchangeTxId");


--
-- Name: t_sol_fee_statistics_ComputeUnitPrice_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_fee_statistics_ComputeUnitPrice_idx" ON public.t_sol_fee_statistics USING btree ("ComputeUnitPrice");


--
-- Name: t_sol_game_buy_property_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_game_buy_property_record_State_LastValidBlockHeight_idx" ON public.t_sol_game_buy_property_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_sol_game_exchange_prize_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_game_exchange_prize_record_State_LastValidBlockHeight_idx" ON public.t_sol_game_exchange_prize_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_sol_game_register_claim_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_game_register_claim_record_State_LastValidBlockHeight_idx" ON public.t_sol_game_register_claim_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_sol_game_unlock_role_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_game_unlock_role_record_State_LastValidBlockHeight_idx" ON public.t_sol_game_unlock_role_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_sol_official_give_token_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_official_give_token_record_State_LastValidBlockHeight_idx" ON public.t_sol_official_give_token_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_sol_pos_mission_config_RewardType_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_pos_mission_config_RewardType_idx" ON public.t_sol_pos_mission_config USING btree ("RewardType");


--
-- Name: t_sol_pos_reward_GroupId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_pos_reward_GroupId_idx" ON public.t_sol_pos_reward USING btree ("GroupId");


--
-- Name: t_sol_pos_reward_NativeAccount_Day_RewardType_Starred_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_sol_pos_reward_NativeAccount_Day_RewardType_Starred_idx" ON public.t_sol_pos_reward USING btree ("NativeAccount", "Day", "RewardType", "Starred");


--
-- Name: t_sol_pos_reward_NativeAccount_State_Pending_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_pos_reward_NativeAccount_State_Pending_idx" ON public.t_sol_pos_reward USING btree ("NativeAccount", "State", "Pending");


--
-- Name: t_sol_pos_reward_claim_record_State_CreateTime_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_pos_reward_claim_record_State_CreateTime_idx" ON public.t_sol_pos_reward_claim_record USING btree ("State", "CreateTime");


--
-- Name: t_sol_pos_reward_claim_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_pos_reward_claim_record_State_LastValidBlockHeight_idx" ON public.t_sol_pos_reward_claim_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_sol_pos_reward_claim_record_TxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_pos_reward_claim_record_TxId_idx" ON public.t_sol_pos_reward_claim_record USING btree ("TxId");


--
-- Name: t_sol_pos_snap_shot_NativeAccount_Day_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_sol_pos_snap_shot_NativeAccount_Day_idx" ON public.t_sol_pos_snap_shot USING btree ("NativeAccount", "Day");


--
-- Name: t_sol_pos_snap_shot_NativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_pos_snap_shot_NativeAccount_idx" ON public.t_sol_pos_snap_shot USING btree ("NativeAccount");


--
-- Name: t_sol_pos_social_media_mission_rec_Provider_UserID_Day_Type_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_sol_pos_social_media_mission_rec_Provider_UserID_Day_Type_idx" ON public.t_sol_pos_social_media_mission_record USING btree ("Provider", "UserID", "Day", "Type");


--
-- Name: t_sol_pos_social_media_mission_record_Provider_UserID_State_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_pos_social_media_mission_record_Provider_UserID_State_idx" ON public.t_sol_pos_social_media_mission_record USING btree ("Provider", "UserID", "State");


--
-- Name: t_sol_pos_star_level_config_NativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_sol_pos_star_level_config_NativeAccount_idx" ON public.t_sol_pos_star_level_config USING btree ("NativeAccount");


--
-- Name: t_sol_stake_reward_GroupId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_stake_reward_GroupId_idx" ON public.t_sol_stake_reward USING btree ("GroupId");


--
-- Name: t_sol_stake_reward_NativeAccount_Day_RewardType_Starred_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_sol_stake_reward_NativeAccount_Day_RewardType_Starred_idx" ON public.t_sol_stake_reward USING btree ("NativeAccount", "Day", "RewardType", "Starred");


--
-- Name: t_sol_stake_reward_NativeAccount_State_Pending_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_stake_reward_NativeAccount_State_Pending_idx" ON public.t_sol_stake_reward USING btree ("NativeAccount", "State", "Pending");


--
-- Name: t_sol_stake_reward_claim_record_State_CreateTime_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_stake_reward_claim_record_State_CreateTime_idx" ON public.t_sol_stake_reward_claim_record USING btree ("State", "CreateTime");


--
-- Name: t_sol_stake_reward_claim_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_stake_reward_claim_record_State_LastValidBlockHeight_idx" ON public.t_sol_stake_reward_claim_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_sol_stake_reward_claim_record_TxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_stake_reward_claim_record_TxId_idx" ON public.t_sol_stake_reward_claim_record USING btree ("TxId");


--
-- Name: t_sol_stake_snap_shot_NativeAccount_Day_StakeType_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_sol_stake_snap_shot_NativeAccount_Day_StakeType_idx" ON public.t_sol_stake_snap_shot USING btree ("NativeAccount", "Day", "StakeType");


--
-- Name: t_sol_stake_snap_shot_NativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_stake_snap_shot_NativeAccount_idx" ON public.t_sol_stake_snap_shot USING btree ("NativeAccount");


--
-- Name: t_sol_stake_star_level_config_NativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_sol_stake_star_level_config_NativeAccount_idx" ON public.t_sol_stake_star_level_config USING btree ("NativeAccount");


--
-- Name: t_sol_swap_token_record_Batch_HouseId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_swap_token_record_Batch_HouseId_idx" ON public.t_sol_swap_token_record USING btree ("Batch", "HouseId");


--
-- Name: t_sol_swap_token_record_HouseId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_swap_token_record_HouseId_idx" ON public.t_sol_swap_token_record USING btree ("HouseId");


--
-- Name: t_sol_swap_token_record_TxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_swap_token_record_TxId_idx" ON public.t_sol_swap_token_record USING btree ("TxId");


--
-- Name: t_sol_swap_token_record_UserNativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_swap_token_record_UserNativeAccount_idx" ON public.t_sol_swap_token_record USING btree ("UserNativeAccount");


--
-- Name: t_sol_transfer_checked_record_FromNativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_checked_record_FromNativeAccount_idx" ON public.t_sol_transfer_checked_record USING btree ("FromNativeAccount");


--
-- Name: t_sol_transfer_checked_record_ReceiptNativeAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_checked_record_ReceiptNativeAccount_idx" ON public.t_sol_transfer_checked_record USING btree ("ReceiptNativeAccount");


--
-- Name: t_sol_transfer_checked_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_checked_record_State_LastValidBlockHeight_idx" ON public.t_sol_transfer_checked_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_sol_transfer_checked_record_TokenMintAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_checked_record_TokenMintAccount_idx" ON public.t_sol_transfer_checked_record USING btree ("TokenMintAccount");


--
-- Name: t_sol_transfer_reward_record_Brand_TokenSymbol_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_reward_record_Brand_TokenSymbol_idx" ON public.t_sol_transfer_reward_record USING btree ("Brand", "TokenSymbol");


--
-- Name: t_sol_transfer_reward_record_Brand_TokenSymbol_idx1; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_reward_record_Brand_TokenSymbol_idx1" ON public.t_sol_transfer_reward_record USING btree ("Brand", "TokenSymbol");


--
-- Name: t_sol_transfer_reward_record_RewardTxId_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_reward_record_RewardTxId_idx" ON public.t_sol_transfer_reward_record USING btree ("RewardTxId");


--
-- Name: t_sol_transfer_reward_record_RewardTxId_idx1; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_reward_record_RewardTxId_idx1" ON public.t_sol_transfer_reward_record USING btree ("RewardTxId");


--
-- Name: t_sol_transfer_reward_record_State_LastValidBlockHeight_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_reward_record_State_LastValidBlockHeight_idx" ON public.t_sol_transfer_reward_record USING btree ("State", "LastValidBlockHeight");


--
-- Name: t_sol_transfer_reward_record_State_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_reward_record_State_idx" ON public.t_sol_transfer_reward_record USING btree ("State");


--
-- Name: t_sol_transfer_reward_record_State_idx1; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_reward_record_State_idx1" ON public.t_sol_transfer_reward_record USING btree ("State");


--
-- Name: t_sol_transfer_reward_record_Version_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_reward_record_Version_idx" ON public.t_sol_transfer_reward_record USING btree ("Version");


--
-- Name: t_sol_transfer_reward_record_Version_idx1; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "t_sol_transfer_reward_record_Version_idx1" ON public.t_sol_transfer_reward_record USING btree ("Version");


--
-- Name: t_swap_scan_info_TokenAccount_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX "t_swap_scan_info_TokenAccount_idx" ON public.t_swap_scan_info USING btree ("TokenAccount");


--
-- Name: t_shit_house_token_quota t_shit_house_token_quota_HouseId_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.t_shit_house_token_quota
    ADD CONSTRAINT "t_shit_house_token_quota_HouseId_fkey" FOREIGN KEY ("HouseId") REFERENCES public.t_shit_house("HouseId");


--
-- Name: TABLE t_activity_exchange_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_activity_exchange_record TO meme_server;


--
-- Name: TABLE t_activity_exchange_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_activity_exchange_rule TO meme_server;


--
-- Name: TABLE t_activity_rank; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_activity_rank TO meme_server;


--
-- Name: TABLE t_aws_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_aws_config TO meme_server;


--
-- Name: TABLE t_campaign_key_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_campaign_key_config TO meme_server;


--
-- Name: TABLE t_campaign_rpc_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_campaign_rpc_config TO meme_server;


--
-- Name: TABLE t_campaign_scan_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_campaign_scan_info TO meme_server;


--
-- Name: TABLE t_campaign_tx; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_campaign_tx TO meme_server;


--
-- Name: TABLE t_chain_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_chain_config TO meme_server;


--
-- Name: TABLE t_daily_claim_stats; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_daily_claim_stats TO meme_server;


--
-- Name: TABLE t_debug_white_list; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_debug_white_list TO meme_server;


--
-- Name: TABLE t_ecommerce_order; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_ecommerce_order TO meme_server;


--
-- Name: TABLE t_facebook_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_facebook_info TO meme_server;


--
-- Name: TABLE t_fee_tolerance; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_fee_tolerance TO meme_server;


--
-- Name: TABLE t_game_rpc_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_game_rpc_config TO meme_server;


--
-- Name: TABLE t_game_scan_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_game_scan_info TO meme_server;


--
-- Name: TABLE t_game_tx; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_game_tx TO meme_server;


--
-- Name: TABLE t_global_daily_exchange_limit; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_global_daily_exchange_limit TO meme_server;


--
-- Name: SEQUENCE t_global_daily_exchange_limit_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,USAGE ON SEQUENCE public.t_global_daily_exchange_limit_id_seq TO meme_server;


--
-- Name: TABLE t_helius_api; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_helius_api TO meme_server;


--
-- Name: TABLE t_helius_scan_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_helius_scan_info TO meme_server;


--
-- Name: TABLE t_micro_service_lb; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_micro_service_lb TO meme_server;


--
-- Name: TABLE t_pos_reward_key_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_pos_reward_key_config TO meme_server;


--
-- Name: TABLE t_pos_rpc_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_pos_rpc_config TO meme_server;


--
-- Name: TABLE t_pos_scan_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_pos_scan_info TO meme_server;


--
-- Name: TABLE t_pos_tx; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_pos_tx TO meme_server;


--
-- Name: TABLE t_rank_activity_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_rank_activity_rule TO meme_server;


--
-- Name: TABLE t_reward_code; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_code TO meme_server;


--
-- Name: TABLE t_reward_code_fee; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_code_fee TO meme_server;


--
-- Name: TABLE t_reward_code_key_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_code_key_config TO meme_server;


--
-- Name: TABLE t_reward_code_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_code_rule TO meme_server;


--
-- Name: TABLE t_reward_discount_rate; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_discount_rate TO meme_server;


--
-- Name: TABLE t_reward_exclude; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_exclude TO meme_server;


--
-- Name: TABLE t_reward_hacker; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_hacker TO meme_server;


--
-- Name: TABLE t_reward_key_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_key_config TO meme_server;


--
-- Name: TABLE t_reward_lottery; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_lottery TO meme_server;


--
-- Name: TABLE t_reward_lottery_claim_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_lottery_claim_record TO meme_server;


--
-- Name: TABLE t_reward_private_key; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_private_key TO meme_server;


--
-- Name: TABLE t_reward_rpc_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_rpc_config TO meme_server;


--
-- Name: TABLE t_reward_scan_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_scan_info TO meme_server;


--
-- Name: TABLE t_reward_test; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_test TO meme_server;


--
-- Name: TABLE t_reward_tx; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_tx TO meme_server;


--
-- Name: TABLE t_reward_tx_scan; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_reward_tx_scan TO meme_server;


--
-- Name: TABLE t_shit_house; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_shit_house TO meme_server;


--
-- Name: TABLE t_shit_house_token_quota; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_shit_house_token_quota TO meme_server;


--
-- Name: TABLE t_sol_airdrop_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_airdrop_record TO meme_server;


--
-- Name: TABLE t_sol_campaign_exchange_score_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_campaign_exchange_score_rule TO meme_server;


--
-- Name: TABLE t_sol_determine_invite_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_determine_invite_record TO meme_server;


--
-- Name: TABLE t_sol_exchange_campaign_score_to_token_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_exchange_campaign_score_to_token_record TO meme_server;


--
-- Name: TABLE t_sol_fee_statistics; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_fee_statistics TO meme_server;


--
-- Name: TABLE t_sol_fund_flow; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_fund_flow TO meme_server;


--
-- Name: TABLE t_sol_fund_flow_old; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_fund_flow_old TO meme_server;


--
-- Name: TABLE t_sol_game_buy_property_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_buy_property_record TO meme_server;


--
-- Name: TABLE t_sol_game_exchange_prize_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_exchange_prize_record TO meme_server;


--
-- Name: TABLE t_sol_game_pack_account; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_pack_account TO meme_server;


--
-- Name: TABLE t_sol_game_prize_account; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_prize_account TO meme_server;


--
-- Name: TABLE t_sol_game_prize_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_prize_info TO meme_server;


--
-- Name: TABLE t_sol_game_property_account; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_property_account TO meme_server;


--
-- Name: TABLE t_sol_game_property_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_property_info TO meme_server;


--
-- Name: TABLE t_sol_game_property_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_property_record TO meme_server;


--
-- Name: TABLE t_sol_game_rank_top; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_rank_top TO meme_server;


--
-- Name: TABLE t_sol_game_receipt_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_receipt_info TO meme_server;


--
-- Name: TABLE t_sol_game_register_claim_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_register_claim_record TO meme_server;


--
-- Name: TABLE t_sol_game_register_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_register_info TO meme_server;


--
-- Name: TABLE t_sol_game_role_account; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_role_account TO meme_server;


--
-- Name: TABLE t_sol_game_role_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_role_info TO meme_server;


--
-- Name: TABLE t_sol_game_unlock_role_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_unlock_role_record TO meme_server;


--
-- Name: TABLE t_sol_game_vote_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_game_vote_record TO meme_server;


--
-- Name: TABLE t_sol_native_account_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_native_account_info TO meme_server;


--
-- Name: TABLE t_sol_official_give_token_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_official_give_token_record TO meme_server;


--
-- Name: TABLE t_sol_official_give_token_reward_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_official_give_token_reward_record TO meme_server;


--
-- Name: TABLE t_sol_official_give_token_reward_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_official_give_token_reward_rule TO meme_server;


--
-- Name: TABLE t_sol_official_transfer_token_reward_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_official_transfer_token_reward_record TO meme_server;


--
-- Name: TABLE t_sol_official_transfer_token_reward_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_official_transfer_token_reward_rule TO meme_server;


--
-- Name: TABLE t_sol_pos_mission_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_mission_config TO meme_server;


--
-- Name: TABLE t_sol_pos_retweet_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_retweet_config TO meme_server;


--
-- Name: TABLE t_sol_pos_reward; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_reward TO meme_server;


--
-- Name: TABLE t_sol_pos_reward_claim_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_reward_claim_record TO meme_server;


--
-- Name: TABLE t_sol_pos_reward_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_reward_rule TO meme_server;


--
-- Name: TABLE t_sol_pos_snap_shot; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_snap_shot TO meme_server;


--
-- Name: TABLE t_sol_pos_social_media_mission_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_social_media_mission_record TO meme_server;


--
-- Name: TABLE t_sol_pos_star_level_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_star_level_config TO meme_server;


--
-- Name: TABLE t_sol_pos_star_level_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_star_level_rule TO meme_server;


--
-- Name: TABLE t_sol_pos_twitter_oauth2_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_pos_twitter_oauth2_config TO meme_server;


--
-- Name: TABLE t_sol_qn_fee; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_qn_fee TO meme_server;


--
-- Name: TABLE t_sol_scan_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_scan_info TO meme_server;


--
-- Name: TABLE t_sol_stake_fix_interest_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_stake_fix_interest_config TO meme_server;


--
-- Name: TABLE t_sol_stake_invite_dist; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_stake_invite_dist TO meme_server;


--
-- Name: TABLE t_sol_stake_invite_rate; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_stake_invite_rate TO meme_server;


--
-- Name: TABLE t_sol_stake_reward; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_stake_reward TO meme_server;


--
-- Name: TABLE t_sol_stake_reward_claim_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_stake_reward_claim_record TO meme_server;


--
-- Name: TABLE t_sol_stake_reward_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_stake_reward_rule TO meme_server;


--
-- Name: TABLE t_sol_stake_snap_shot; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_stake_snap_shot TO meme_server;


--
-- Name: TABLE t_sol_stake_star_level_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_stake_star_level_config TO meme_server;


--
-- Name: TABLE t_sol_stake_star_level_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_stake_star_level_rule TO meme_server;


--
-- Name: TABLE t_sol_swap_new_token_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_swap_new_token_config TO meme_server;


--
-- Name: TABLE t_sol_swap_token_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_swap_token_record TO meme_server;


--
-- Name: TABLE t_sol_swap_token_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_swap_token_rule TO meme_server;


--
-- Name: TABLE t_sol_token_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_token_config TO meme_server;


--
-- Name: TABLE t_sol_transfer_checked_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_transfer_checked_record TO meme_server;


--
-- Name: TABLE t_sol_transfer_reward_claim; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_transfer_reward_claim TO meme_server;


--
-- Name: TABLE t_sol_transfer_reward_distribution; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_transfer_reward_distribution TO meme_server;


--
-- Name: TABLE t_sol_transfer_reward_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_transfer_reward_record TO meme_server;


--
-- Name: TABLE t_sol_transfer_to_dex_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_transfer_to_dex_record TO meme_server;


--
-- Name: TABLE t_sol_transfer_token_reward_rule; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_sol_transfer_token_reward_rule TO meme_server;


--
-- Name: TABLE t_stake_amm_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_stake_amm_config TO meme_server;


--
-- Name: TABLE t_stake_area_leader; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_stake_area_leader TO meme_server;


--
-- Name: TABLE t_stake_buy_token; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_stake_buy_token TO meme_server;


--
-- Name: TABLE t_stake_record; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_stake_record TO meme_server;


--
-- Name: TABLE t_stake_token_pool; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_stake_token_pool TO meme_server;


--
-- Name: TABLE t_stake_total_area_leader; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_stake_total_area_leader TO meme_server;


--
-- Name: TABLE t_swap_key_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_swap_key_config TO meme_server;


--
-- Name: TABLE t_swap_scan_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_swap_scan_info TO meme_server;


--
-- Name: TABLE t_system_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_system_config TO meme_server;


--
-- Name: TABLE t_tweets_info; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_tweets_info TO meme_server;


--
-- Name: TABLE t_user_daily_exchange_quota; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_user_daily_exchange_quota TO meme_server;


--
-- Name: SEQUENCE t_user_daily_exchange_quota_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT SELECT,USAGE ON SEQUENCE public.t_user_daily_exchange_quota_id_seq TO meme_server;


--
-- Name: TABLE t_user_wallet_rpc_config; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.t_user_wallet_rpc_config TO meme_server;


--
-- PostgreSQL database dump complete
--

