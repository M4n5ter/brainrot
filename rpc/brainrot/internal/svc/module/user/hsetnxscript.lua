local key = KEYS[1]
local expire = tonumber(KEYS[2]) or 604800  -- 默认值为604800秒（7天）
local exists = redis.call("EXISTS", key)

if #ARGV % 2 ~= 0 then
    error("ARGV must contain an even number of elements")
end

if exists == 1 then
    error("Hash table exists")
else
    for i = 1, #ARGV, 2 do
        local field = ARGV[i]
        local value = ARGV[i + 1]
        redis.call("HSET", key, field, value)
    end
    redis.call("EXPIRE", key, expire)
end