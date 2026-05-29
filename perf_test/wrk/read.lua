-- wrk: чтение объявления по ID (GET /api/admin/ads/{id})
-- Переменные: READ_MIN_AD_ID, READ_MAX_AD_ID

local min_id = tonumber(os.getenv("READ_MIN_AD_ID") or "1")
local max_id = tonumber(os.getenv("READ_MAX_AD_ID") or "100000")

if max_id < min_id then
  error("READ_MAX_AD_ID must be >= READ_MIN_AD_ID")
end

request = function()
  local id = math.random(min_id, max_id)
  return wrk.format("GET", "/api/admin/ads/" .. id)
end
