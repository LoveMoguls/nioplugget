#!/bin/bash
set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "[1/4] Starting database..."
if ! docker ps --format '{{.Names}}' | grep -q "^nioplugget-db$"; then
    docker start nioplugget-db
    sleep 3
else
    echo "       already running"
fi

echo "[2/4] Starting backend..."
cd "$PROJECT_DIR/backend"
set -a && source .env && set +a
go run ./cmd/server/main.go > /tmp/nioplugget-backend.log 2>&1 &
BACKEND_PID=$!
echo "       PID $BACKEND_PID — logs: /tmp/nioplugget-backend.log"

echo "[3/4] Starting frontend..."
cd "$PROJECT_DIR/frontend"
npm run dev -- --host 0.0.0.0 > /tmp/nioplugget-frontend.log 2>&1 &
FRONTEND_PID=$!
echo "       PID $FRONTEND_PID — logs: /tmp/nioplugget-frontend.log"

echo "[4/4] Starting Cloudflare tunnel..."
sleep 2
cloudflared tunnel --url http://localhost:5173 > /tmp/nioplugget-cloudflare.log 2>&1 &
CF_PID=$!
echo "       PID $CF_PID — logs: /tmp/nioplugget-cloudflare.log"

echo ""
echo "Waiting for Cloudflare URL..."
for i in $(seq 1 20); do
    CF_URL=$(grep -o 'https://[a-zA-Z0-9-]*\.trycloudflare\.com' /tmp/nioplugget-cloudflare.log 2>/dev/null | head -1)
    if [ -n "$CF_URL" ]; then
        break
    fi
    sleep 1
done

echo ""
echo "================================"
echo " All services started!"
echo "================================"
echo " Frontend (local):  http://localhost:5173"
echo " Frontend (network): http://$(hostname -I | awk '{print $1}'):5173"
if [ -n "$CF_URL" ]; then
    echo " Cloudflare URL:    $CF_URL"
else
    echo " Cloudflare URL:    (check /tmp/nioplugget-cloudflare.log)"
fi
echo " Backend:           http://localhost:8080"
echo "================================"
echo ""
echo "Press Ctrl+C to stop all services."

trap "echo ''; echo 'Stopping...'; kill $BACKEND_PID $FRONTEND_PID $CF_PID 2>/dev/null; docker stop nioplugget-db; echo 'Done.'" EXIT INT TERM

wait
