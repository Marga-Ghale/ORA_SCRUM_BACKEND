#!/bin/bash
set -euo pipefail

API="http://localhost:8080/api"

ADMIN_EMAIL="dosav79403@mekuron.com"
MEMBER_EMAIL="payok67516@nctime.com"
PASSWORD="password123"

require_token () {
  if [[ -z "$1" || "$1" == "null" ]]; then
    echo "❌ Failed to obtain token"
    exit 1
  fi
}

echo "🔐 Logging in as admin..."
TOKEN1=$(curl -s -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$PASSWORD\"}" \
  | jq -r '.accessToken')

require_token "$TOKEN1"

echo "🔐 Logging in as member..."
TOKEN2=$(curl -s -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$MEMBER_EMAIL\",\"password\":\"$PASSWORD\"}" \
  | jq -r '.accessToken')

require_token "$TOKEN2"

echo "🏗️  Creating workspace..."
WS_ID=$(curl -s -X POST "$API/workspaces" \
  -H "Authorization: Bearer $TOKEN1" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test WS"}' \
  | jq -r '.id')

if [[ -z "$WS_ID" || "$WS_ID" == "null" ]]; then
  echo "❌ Failed to create workspace"
  exit 1
fi

echo "✅ Workspace created: $WS_ID"

echo "📨 Inviting member..."
curl -s -X POST "$API/members/workspace/$WS_ID/invite" \
  -H "Authorization: Bearer $TOKEN1" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$MEMBER_EMAIL\",\"role\":\"member\"}" \
  | jq '.'

echo "🔔 Member Notifications"
echo "======================="
curl -s -X GET "$API/notifications" \
  -H "Authorization: Bearer $TOKEN2" \
  | jq '.'
