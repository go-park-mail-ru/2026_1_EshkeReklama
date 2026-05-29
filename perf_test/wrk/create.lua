local campaign = os.getenv("LOADTEST_CAMPAIGN")
local group = os.getenv("LOADTEST_GROUP")
local prefix = os.getenv("LOADTEST_PREFIX") or "LOADTEST_"
local session_id = os.getenv("WRK_SESSION_ID")
local csrf_token = os.getenv("WRK_CSRF_TOKEN")

local boundary = "----WrkFormBoundary7MA4YWxkTrZu0gW"
local counter = 0

local function body_for(title)
  return table.concat({
    "--", boundary, "\r\n",
    "Content-Disposition: form-data; name=\"title\"\r\n\r\n", title, "\r\n",
    "--", boundary, "\r\n",
    "Content-Disposition: form-data; name=\"short_desc\"\r\n\r\n",
    "load test short description\r\n",
    "--", boundary, "\r\n",
    "Content-Disposition: form-data; name=\"target_url\"\r\n\r\n",
    "https://example.com/loadtest\r\n",
    "--", boundary, "--\r\n",
  })
end

request = function()
  counter = counter + 1
  local title = prefix .. counter
  local path = string.format("/api/ad_campaigns/%s/ad_groups/%s/ads", campaign, group)
  local body = body_for(title)
  return wrk.format("POST", path, {
    ["Content-Type"] = "multipart/form-data; boundary=" .. boundary,
    ["Content-Length"] = tostring(#body),
    ["Cookie"] = "session_id=" .. session_id .. "; csrf_token=" .. csrf_token,
    ["X-CSRF-Token"] = csrf_token,
  }, body)
end
