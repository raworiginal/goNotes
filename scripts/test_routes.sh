#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:3000"
USERNAME="testuser_$$"
PASSWORD="testpassword123"

GREEN='\033[0;32m'
RED='\033[0;31m'
BOLD='\033[1m'
RESET='\033[0m'

pass() { echo -e "${GREEN}✓ $1${RESET}"; }
fail() {
  echo -e "${RED}✗ $1${RESET}"
  exit 1
}
section() { echo -e "\n${BOLD}── $1 ──${RESET}"; }

expect_status() {
  local label="$1" expected="$2" actual="$3"
  if [ "$actual" -eq "$expected" ]; then
    pass "$label (HTTP $actual)"
  else
    fail "$label — expected HTTP $expected, got HTTP $actual"
  fi
}

# ── Auth ──────────────────────────────────────────────────────────────────────

section "Register"
REGISTER_RES=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
REGISTER_BODY=$(echo "$REGISTER_RES" | head -n -1)
REGISTER_STATUS=$(echo "$REGISTER_RES" | tail -n 1)
expect_status "Register user" 201 "$REGISTER_STATUS"
echo "$REGISTER_BODY" | jq .

section "Login"
LOGIN_RES=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
LOGIN_BODY=$(echo "$LOGIN_RES" | head -n -1)
LOGIN_STATUS=$(echo "$LOGIN_RES" | tail -n 1)
expect_status "Login" 200 "$LOGIN_STATUS"
echo "$LOGIN_BODY" | jq .

ACCESS_TOKEN=$(echo "$LOGIN_BODY" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_BODY" | jq -r '.refresh_token')

sleep 5
section "Refresh token"
REFRESH_RES=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")
REFRESH_BODY=$(echo "$REFRESH_RES" | head -n -1)
REFRESH_STATUS=$(echo "$REFRESH_RES" | tail -n 1)
expect_status "Refresh token" 200 "$REFRESH_STATUS"
echo "$REFRESH_BODY" | jq .

ACCESS_TOKEN=$(echo "$REFRESH_BODY" | jq -r '.access_token')
AUTH_HEADER="Authorization: Bearer $ACCESS_TOKEN"

# ── Create notes ───────────────────────────────────────────────────────────────

section "Create text note"
TEXT_NOTE_RES=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/notes/" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"title":"Shopping List","type":"text","body":"milk, eggs, bread"}')
TEXT_NOTE_BODY=$(echo "$TEXT_NOTE_RES" | head -n -1)
TEXT_NOTE_STATUS=$(echo "$TEXT_NOTE_RES" | tail -n 1)
expect_status "Create text note" 201 "$TEXT_NOTE_STATUS"
echo "$TEXT_NOTE_BODY" | jq .
TEXT_NOTE_ID=$(echo "$TEXT_NOTE_BODY" | jq -r '.id')

section "Create checklist note"
CHECKLIST_NOTE_RES=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/notes/" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "title": "Groceries",
    "type": "checklist",
    "items": [
      {"text": "cheese", "completed": false},
      {"text": "coffee", "completed": false}
    ]
  }')
CHECKLIST_NOTE_BODY=$(echo "$CHECKLIST_NOTE_RES" | head -n -1)
CHECKLIST_NOTE_STATUS=$(echo "$CHECKLIST_NOTE_RES" | tail -n 1)
expect_status "Create checklist note" 201 "$CHECKLIST_NOTE_STATUS"
echo "$CHECKLIST_NOTE_BODY" | jq .
CHECKLIST_NOTE_ID=$(echo "$CHECKLIST_NOTE_BODY" | jq -r '.id')

# ── Get notes ──────────────────────────────────────────────────────────────────

section "Get text note by ID"
GET_TEXT_RES=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/notes/$TEXT_NOTE_ID" \
  -H "$AUTH_HEADER")
GET_TEXT_BODY=$(echo "$GET_TEXT_RES" | head -n -1)
GET_TEXT_STATUS=$(echo "$GET_TEXT_RES" | tail -n 1)
expect_status "Get text note by ID" 200 "$GET_TEXT_STATUS"
echo "$GET_TEXT_BODY" | jq .

section "Get all notes"
ALL_NOTES_RES=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/notes/" \
  -H "$AUTH_HEADER")
ALL_NOTES_BODY=$(echo "$ALL_NOTES_RES" | head -n -1)
ALL_NOTES_STATUS=$(echo "$ALL_NOTES_RES" | tail -n 1)
expect_status "Get all notes" 200 "$ALL_NOTES_STATUS"
echo "$ALL_NOTES_BODY" | jq .

# ── Update (PUT) ───────────────────────────────────────────────────────────────

section "Update text note (PUT)"
PUT_TEXT_RES=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/notes/$TEXT_NOTE_ID" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"title":"Updated Shopping List","type":"text","body":"milk, eggs, bread, butter"}')
PUT_TEXT_BODY=$(echo "$PUT_TEXT_RES" | head -n -1)
PUT_TEXT_STATUS=$(echo "$PUT_TEXT_RES" | tail -n 1)
expect_status "Update text note (PUT)" 200 "$PUT_TEXT_STATUS"
echo "$PUT_TEXT_BODY" | jq .

section "Update checklist note (PUT)"
PUT_CHECKLIST_RES=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/notes/$CHECKLIST_NOTE_ID" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "title": "Updated Groceries",
    "type": "checklist",
    "items": [
      {"text": "cheese", "completed": true},
      {"text": "coffee", "completed": false},
      {"text": "oat milk", "completed": false}
    ]
  }')
PUT_CHECKLIST_BODY=$(echo "$PUT_CHECKLIST_RES" | head -n -1)
PUT_CHECKLIST_STATUS=$(echo "$PUT_CHECKLIST_RES" | tail -n 1)
expect_status "Update checklist note (PUT)" 200 "$PUT_CHECKLIST_STATUS"
echo "$PUT_CHECKLIST_BODY" | jq .

# ── Patch (PATCH) ──────────────────────────────────────────────────────────────

section "Patch text note title (PATCH)"
PATCH_TEXT_RES=$(curl -s -w "\n%{http_code}" -X PATCH "$BASE_URL/notes/$TEXT_NOTE_ID" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"title":"Patched Shopping List"}')
PATCH_TEXT_BODY=$(echo "$PATCH_TEXT_RES" | head -n -1)
PATCH_TEXT_STATUS=$(echo "$PATCH_TEXT_RES" | tail -n 1)
expect_status "Patch text note title" 200 "$PATCH_TEXT_STATUS"
echo "$PATCH_TEXT_BODY" | jq .

section "Patch checklist note title (PATCH)"
PATCH_CHECKLIST_RES=$(curl -s -w "\n%{http_code}" -X PATCH "$BASE_URL/notes/$CHECKLIST_NOTE_ID" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"title":"Patched Groceries"}')
PATCH_CHECKLIST_BODY=$(echo "$PATCH_CHECKLIST_RES" | head -n -1)
PATCH_CHECKLIST_STATUS=$(echo "$PATCH_CHECKLIST_RES" | tail -n 1)
expect_status "Patch checklist note title" 200 "$PATCH_CHECKLIST_STATUS"
echo "$PATCH_CHECKLIST_BODY" | jq .

# ── Delete ─────────────────────────────────────────────────────────────────────

section "Delete text note"
DELETE_RES=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/notes/$TEXT_NOTE_ID" \
  -H "$AUTH_HEADER")
DELETE_STATUS=$(echo "$DELETE_RES" | tail -n 1)
expect_status "Delete text note" 204 "$DELETE_STATUS"

section "Get all notes (after delete)"
FINAL_RES=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/notes/" \
  -H "$AUTH_HEADER")
FINAL_BODY=$(echo "$FINAL_RES" | head -n -1)
FINAL_STATUS=$(echo "$FINAL_RES" | tail -n 1)
expect_status "Get all notes after delete" 200 "$FINAL_STATUS"
echo "$FINAL_BODY" | jq .

echo -e "\n${GREEN}${BOLD}All tests passed.${RESET}"
