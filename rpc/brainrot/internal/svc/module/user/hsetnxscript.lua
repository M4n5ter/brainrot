local key = KEYS[1]
local exists = redis.call("EXISTS", key)

if exists == 1 then
    error("Hash table exists")
else
    for i = 1, #ARGV, 2 do
        local field = ARGV[i]
        local value = ARGV[i + 1]
        redis.call("HSET", key, field, value)
    end
    redis.call("EXPIRE", key, 604800)
end