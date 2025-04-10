package utils

import (
	"fmt"
	"math/rand"
	"time"
)

var (
	adjectives = []string{
		"快乐的", "聪明的", "勇敢的", "友好的", "活泼的",
		"机智的", "可爱的", "温柔的", "善良的", "帅气的",
	}

	animals = []string{
		"熊猫", "老虎", "狮子", "大象", "长颈鹿",
		"猴子", "兔子", "狐狸", "浣熊", "海豚",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 生成随机名字
func GenerateRandomName() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	animal := animals[rand.Intn(len(animals))]
	return fmt.Sprintf("%s%s", adj, animal)
}
