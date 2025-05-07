
math.randomseed(os.time())

request = function()
    local delta = math.random(1, 1000000)
    local value = 0.35 + math.random(1, 10000)

    local gaugeId = "GaugeCount" .. "_" .. tostring(value)
    local counterId = "PollCount" .. "_" .. tostring(delta)

    --local body = string.format('{"id": "%s", "type": "counter", "delta": %d}', counterId, delta)

    local body = string.format(
        '[%s]',
        table.concat({
            string.format('{"id": "%s", "type": "gauge", "value": %.2f}', gaugeId, value),
            string.format('{"id": "%s", "type": "counter", "delta": %d}', counterId, delta)
        }, ",")
    )

    return wrk.format("POST", "/updates/", {["Content-Type"] = "application/json"}, body)
end

