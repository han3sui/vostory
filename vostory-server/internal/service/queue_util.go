package service

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// 后台队列：消息格式均为「任务ID:对象ID」
const (
	ttsQueueKey = "vs:tts:queue"
	llmQueueKey = "vs:llm:queue"
)

// removeQueueScript 一次性移除队列中指定前缀的消息，并保持其余消息的顺序。
// 逐条 LREM 是 O(N²)，队列有几千条时会长时间阻塞 Redis。
var removeQueueScript = redis.NewScript(`
local items = redis.call('LRANGE', KEYS[1], 0, -1)
local prefix = ARGV[1]
local kept = {}
for _, v in ipairs(items) do
	if string.sub(v, 1, #prefix) ~= prefix then
		kept[#kept + 1] = v
	end
end
local removed = #items - #kept
if removed > 0 then
	redis.call('DEL', KEYS[1])
	for i = 1, #kept, 1000 do
		redis.call('RPUSH', KEYS[1], unpack(kept, i, math.min(i + 999, #kept)))
	end
end
return removed
`)

// removeTaskQueueMessages 移除队列中属于指定任务的全部消息
func removeTaskQueueMessages(ctx context.Context, rdb *redis.Client, queueKey string, taskID uint64) error {
	return removeQueueScript.Run(ctx, rdb, []string{queueKey}, fmt.Sprintf("%d:", taskID)).Err()
}
