local min_id = tonumber(os.getenv("READ_MIN_AD_ID"))
local max_id = tonumber(os.getenv("READ_MAX_AD_ID"))

request = function()
  local id = math.random(min_id, max_id)
  return wrk.format("GET", "/api/admin/ads/" .. id)
end
