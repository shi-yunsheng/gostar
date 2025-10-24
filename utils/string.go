package utils

import (
	"hash/fnv"
	"math/rand"
	"strings"
	"time"
	"unicode"
)

// @en generate random string with specified length
//
// @zh 生成指定长度的随机字符串
func GetRandomString(length int) string {
	// @en define character set for random string
	// @zh 定义随机字符串的字符集
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// @en create new random source with current time
	// @zh 使用当前时间创建新的随机源
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// @en generate random string
	// @zh 生成随机字符串
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[r.Intn(len(charset))]
	}

	return string(result)
}

// @en convert camel to snake
//
// @zh 将驼峰转成蛇形
func CamelToSnake(camel string) string {
	runes := []rune(camel)
	var result []rune

	for i, r := range runes {
		// @en if the first character, convert to lowercase
		// @zh 如果是第一个字符，直接转为小写
		if i == 0 {
			result = append(result, unicode.ToLower(r))
			continue
		}

		// @en if the current character is an uppercase letter
		// @zh 当前字符是大写字母时
		if unicode.IsUpper(r) {
			prev := runes[i-1]

			// @en condition 1: the previous character is a lowercase letter
			// condition 2: the previous character is an uppercase letter, and the next character is a lowercase letter
			// condition 3: the previous character should not be a dot or underscore
			// @zh 条件1：前一个字符是小写字母
			// 条件2：前一个字符是大写字母，且下一个字符是小写字母
			// 条件3：前一个字符不应是点或下划线
			if (unicode.IsLower(prev) ||
				(i+1 < len(runes) && unicode.IsLower(runes[i+1]))) &&
				(unicode.IsLetter(prev) || unicode.IsDigit(prev)) {
				result = append(result, '_')
			}
		}

		result = append(result, unicode.ToLower(r))
	}

	return string(result)
}

// @en convert snake to camel
//
// @zh 将蛇形转成驼峰
func SnakeToCamel(s string, isFirstUpper ...bool) string {
	words := strings.Split(s, "_")

	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	// @en handle first word
	// @zh 处理第一个单词
	if len(isFirstUpper) == 0 || !isFirstUpper[0] {
		if len(words[0]) > 0 {
			words[0] = strings.ToLower(words[0][:1]) + words[0][1:]
		}
	}

	return strings.Join(words, "")
}

// @en generate hash from string
//
// @zh 从字符串生成哈希值
func StringHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
