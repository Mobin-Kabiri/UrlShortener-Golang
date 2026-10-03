package internal

// 62 valid charactars (for base62)
const validChar string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func EncodeToBase62String(number uint32) string {
	if number == 0 {
		return string(validChar[0])
	}
	// unint8 = element's range is 0-61
	tmpArr := make([]uint8, 0, 8)
	for number != 0 {
		remainder := number % 62
		tmpArr = append(tmpArr, uint8(remainder))
		number = number / 62
	}

	// it can only contains ASCII so we can use byte
	result := make([]byte, 0, len(tmpArr))
	for i := len(tmpArr)-1; i >= 0; i-- {
		result = append(result, validChar[tmpArr[i]])
	}
	
	return string(result)
}

