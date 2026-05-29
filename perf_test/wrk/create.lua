-- wrk: нагрузочное создание объявлений (POST multipart)
-- Переменные окружения (задаёт run_create.sh из .env):
--   LOADTEST_CAMPAIGN, LOADTEST_GROUP, LOADTEST_PREFIX

local campaign = os.getenv("LOADTEST_CAMPAIGN")
local group = os.getenv("LOADTEST_GROUP")
local prefix = os.getenv("LOADTEST_PREFIX") or "LOADTEST_"

if not campaign or not group then
  error("set LOADTEST_CAMPAIGN and LOADTEST_GROUP (run ./perf_test/setup.sh)")
end

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
  local title = string.format("%s%d", prefix, counter)
  local path = string.format(
    "/api/ad_campaigns/%s/ad_groups/%s/ads",
    campaign,
    group
  )
  local headers = {}
  headers["Content-Type"] = "multipart/form-data; boundary=" .. boundary
  return wrk.format("POST", path, headers, body_for(title))
end
