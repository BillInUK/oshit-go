#!/bin/bash
# =============================================================
# 通用服务启动脚本（含 watchdog 自动重启）
#
# 用法：
#   ./server.sh start [background|foreground|debug]
#   ./server.sh stop
#   ./server.sh restart [background|foreground|debug]
#   ./server.sh status
#
# start 模式：
#   background  （默认）后台启动，进程崩溃后 watchdog 自动重启
#   foreground  前台启动，日志直接输出到终端，Ctrl+C 停止（无自动重启）
#   debug       前台启动，Delve 远程调试端口 2345（无自动重启）
# =============================================================

# ===== 部署配置（按实际路径修改）=====
DEPLOY_DIR=$(cd "$(dirname "$0")/.." && pwd)
BIN="${DEPLOY_DIR}/bin/micro-server"
ETC_DIR="${DEPLOY_DIR}/etc"
LOG_DIR="${DEPLOY_DIR}/log"
RESTART_DELAY=3     # 重启等待秒数
DEBUG_PORT=2345
# ======================================

export APP_CONF="${ETC_DIR}/application.yaml"
LOG="${LOG_DIR}/server.log"
WATCHDOG_PID_FILE="${LOG_DIR}/server.pid"      # watchdog 进程 PID
APP_PID_FILE="${LOG_DIR}/server.app.pid"        # micro-server 进程 PID
STOP_FLAG="${LOG_DIR}/server.stop"              # 存在时 watchdog 不重启

# =============================================================
# 内部: watchdog 循环
# 通过 "server.sh _watchdog" 调用自身来实现，避免 bash -c 的转义问题
# =============================================================
if [ "$1" = "_watchdog" ]; then
    echo "$(date '+%F %T') [watchdog] started, monitoring ${BIN}" >> "$LOG"

    while true; do
        # 检查停止标志
        if [ -f "$STOP_FLAG" ]; then
            echo "$(date '+%F %T') [watchdog] stop flag detected, exiting" >> "$LOG"
            exit 0
        fi

        # 启动 micro-server
        "$BIN" &
        app_pid=$!
        echo $app_pid > "$APP_PID_FILE"
        echo "$(date '+%F %T') [watchdog] launched micro-server PID=${app_pid}" >> "$LOG"

        # 等待进程退出
        wait $app_pid
        exit_code=$?
        rm -f "$APP_PID_FILE"

        # 再次检查停止标志（stop 命令在 kill 前会写入标志）
        if [ -f "$STOP_FLAG" ]; then
            echo "$(date '+%F %T') [watchdog] process exited (code=${exit_code}), stop flag detected, exiting" >> "$LOG"
            exit 0
        fi

        # 进程意外退出，等待后重启
        echo "$(date '+%F %T') [watchdog] process exited (code=${exit_code}), restarting in ${RESTART_DELAY}s..." >> "$LOG"
        sleep $RESTART_DELAY
    done
fi

# =============================================================
# 以下为正常 start / stop / status 逻辑
# =============================================================

_check_bin() {
    if [ ! -x "$BIN" ]; then
        echo "[server] error: binary not found or not executable: ${BIN}"
        exit 1
    fi
}

_watchdog_running() {
    [ -f "$WATCHDOG_PID_FILE" ] && kill -0 "$(cat "$WATCHDOG_PID_FILE")" 2>/dev/null
}

# ---------- start ----------

start_background() {
    if _watchdog_running; then
        echo "[server] already running, watchdog PID=$(cat "$WATCHDOG_PID_FILE")"
        exit 0
    fi

    _check_bin
    mkdir -p "$LOG_DIR"
    rm -f "$STOP_FLAG"   # 清除上次 stop 留下的标志

    # 启动 watchdog（脚本调用自身）
    nohup "$0" _watchdog >> "$LOG" 2>&1 &
    echo $! > "$WATCHDOG_PID_FILE"

    # 稍等让 watchdog 拉起 micro-server
    sleep 0.8

    echo "[server] started in background (watchdog enabled)"
    echo "  watchdog PID : $(cat "$WATCHDOG_PID_FILE")"
    if [ -f "$APP_PID_FILE" ]; then
        echo "  server   PID : $(cat "$APP_PID_FILE")"
    fi
    echo "  log          : ${LOG}"
}

start_foreground() {
    if _watchdog_running; then
        echo "[server] already running, watchdog PID=$(cat "$WATCHDOG_PID_FILE")"
        exit 0
    fi
    _check_bin
    mkdir -p "$LOG_DIR"
    echo "[server] starting in foreground (Ctrl+C to stop, no auto-restart)..."
    exec "$BIN"
}

start_debug() {
    _check_bin
    if ! command -v dlv &>/dev/null; then
        echo "[server] error: dlv not found"
        echo "  install: go install github.com/go-delve/delve/cmd/dlv@latest"
        exit 1
    fi
    mkdir -p "$LOG_DIR"
    echo "[server] starting in debug mode (no auto-restart)"
    echo "  Delve listen : :${DEBUG_PORT}"
    echo "  IDE connect  : $(hostname -I | awk '{print $1}'):${DEBUG_PORT}"
    exec dlv exec --headless --listen=":${DEBUG_PORT}" --api-version=2 -- "$BIN"
}

start() {
    local mode="${1:-background}"
    case "$mode" in
        background) start_background ;;
        foreground) start_foreground ;;
        debug)      start_debug      ;;
        *)
            echo "[server] unknown mode: ${mode}"
            echo "  available: background | foreground | debug"
            exit 1
            ;;
    esac
}

# ---------- stop ----------

stop() {
    if ! _watchdog_running; then
        echo "[server] not running"
        rm -f "$WATCHDOG_PID_FILE" "$APP_PID_FILE"
        return
    fi

    local wpid
    wpid=$(cat "$WATCHDOG_PID_FILE")

    # 1. 先写停止标志，防止 watchdog 重启
    touch "$STOP_FLAG"

    # 2. 杀掉 micro-server
    if [ -f "$APP_PID_FILE" ]; then
        local apid
        apid=$(cat "$APP_PID_FILE")
        if kill -0 "$apid" 2>/dev/null; then
            kill "$apid"
        fi
        rm -f "$APP_PID_FILE"
    fi

    # 3. 杀掉 watchdog
    kill "$wpid" 2>/dev/null
    rm -f "$WATCHDOG_PID_FILE"

    echo "[server] stopped (watchdog PID=${wpid})"
}

# ---------- status ----------

status() {
    if _watchdog_running; then
        echo "[server] running"
        echo "  watchdog PID : $(cat "$WATCHDOG_PID_FILE")"
        if [ -f "$APP_PID_FILE" ] && kill -0 "$(cat "$APP_PID_FILE")" 2>/dev/null; then
            echo "  server   PID : $(cat "$APP_PID_FILE")"
        else
            echo "  server   PID : (starting or restarting...)"
        fi
    else
        echo "[server] stopped"
    fi
}

# ---------- 入口 ----------

case "$1" in
    start)   start "$2"             ;;
    stop)    stop                   ;;
    restart) stop; sleep 1; start "$2" ;;
    status)  status                 ;;
    *)
        echo "Usage: $0 {start|stop|restart|status} [background|foreground|debug]"
        exit 1
        ;;
esac