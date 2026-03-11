#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

PORT=3000
BASE_URL="http://localhost:$PORT"
USERNAME="testuser_$(date +%s)"
PASSWORD="secret123"

echo -e "${BLUE}=== Notes API Test Suite ===${NC}\n"

# Test 1: Health check
echo -e "${BLUE}Test 1: Health Check${NC}"
curl -s "$BASE_URL/health" | jq .
echo ""

# Test 2: Register user
echo -e "${BLUE}Test 2: Register User${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
echo "$REGISTER_RESPONSE" | jq .
echo ""

# Test 3: Login
echo -e "${BLUE}Test 3: Login${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")
echo "$LOGIN_RESPONSE" | jq .

# Extract token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo -e "${RED}Failed to get access token${NC}"
  exit 1
fi
echo -e "${GREEN}Token acquired: ${TOKEN:0:20}...${NC}\n"

# Test 4: Create a text note
echo -e "${BLUE}Test 4: Create Text Note${NC}"
CREATE_TEXT=$(curl -s -X POST "$BASE_URL/notes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Shopping","type":"text","body":"Milk, eggs, bread"}')
echo "$CREATE_TEXT" | jq .
NOTE_ID=$(echo "$CREATE_TEXT" | jq -r '.id')
echo ""

# Test 5: Create a checklist note
echo -e "${BLUE}Test 5: Create Checklist Note${NC}"
CREATE_CHECKLIST=$(curl -s -X POST "$BASE_URL/notes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"TODO","type":"checklist","items":[{"text":"Write tests","completed":false},{"text":"Deploy","completed":false}]}')
echo "$CREATE_CHECKLIST" | jq .
echo ""

# Test 6: List all notes
echo -e "${BLUE}Test 6: List All Notes${NC}"
curl -s -X GET "$BASE_URL/notes" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# Test 7: Get a specific note
echo -e "${BLUE}Test 7: Get Specific Note (ID: $NOTE_ID)${NC}"
curl -s -X GET "$BASE_URL/notes/$NOTE_ID" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# Test 8: Update a note (full replacement)
echo -e "${BLUE}Test 8: Update Note (PUT)${NC}"
curl -s -X PUT "$BASE_URL/notes/$NOTE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Shopping List","type":"text","body":"Milk, eggs, bread, butter"}' | jq .
echo ""

# Test 9: Patch a note (partial update)
echo -e "${BLUE}Test 9: Patch Note (PATCH)${NC}"
curl -s -X PATCH "$BASE_URL/notes/$NOTE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Patched Title"}' | jq .
echo ""

# Test 10: Try accessing without token (should fail)
echo -e "${BLUE}Test 10: Access Without Token (Should Fail)${NC}"
curl -s -X GET "$BASE_URL/notes" | jq .
echo ""

# Test 11: Delete a note
echo -e "${BLUE}Test 11: Delete Note${NC}"
DELETE_RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/notes/$NOTE_ID" \
  -H "Authorization: Bearer $TOKEN")
HTTP_CODE=$(echo "$DELETE_RESPONSE" | tail -n1)
BODY=$(echo "$DELETE_RESPONSE" | head -n-1)
echo "HTTP Status: $HTTP_CODE"
if [ ! -z "$BODY" ]; then
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
fi
echo ""

echo -e "${GREEN}=== All Tests Complete ===${NC}"
