package utils

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

//GenValidateCode 生成随机数 如：短信验证码等
func GenValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

//IsContain 判断元素是否存在数组list中
func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

//Min 最小值
func Min(nums ...float64) float64 {
	if len(nums) == 1 {
		return nums[0]
	}
	min := nums[0]
	for i := 1; i < len(nums); i++ {
		min = math.Min(min, nums[i])
	}
	return min
}
