package udp

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// HidEvent 表示一个 HID 输入事件（从 0x71 帧解析）
type HidEvent struct {
	Seq      uint8 // 序列号（0-255，用于 reorder 检测）
	Type     uint8 // 设备类型（0x01=keyboard）
	Usage    uint8 // HID usage code
	Pressed  bool  // 按键状态（true=按下，false=释放）
	Modifier uint8 // 修饰键位图
}

// HidEventReceiver 负责解析 0x71 HID 输入事件帧
type HidEventReceiver struct {
	lastSeq uint8
}

// NewHidEventReceiver 创建一个新的 HID 事件接收器
func NewHidEventReceiver() *HidEventReceiver {
	return &HidEventReceiver{
		lastSeq: 0,
	}
}

// ParseHidFrame 解析 0x71 帧
// 输入：已去除 0x57 0xAB 头部的数据（即从 0x71 开始）
// 帧格式：[0x71][SEQ][TYPE][USAGE][PRESSED][MODIFIER][CRC8]
// 返回：解析后的 HidEvent 和错误信息
func (r *HidEventReceiver) ParseHidFrame(data []byte) (*HidEvent, error) {
	// 检查最小长度（7 字节：TYPE + SEQ + TYPE + USAGE + PRESSED + MODIFIER + CRC8）
	if len(data) < 7 {
		return nil, fmt.Errorf("frame too short: %d bytes, expected 7", len(data))
	}

	// 检查帧类型标识
	if data[0] != 0x71 {
		return nil, fmt.Errorf("invalid frame type: 0x%02X, expected 0x71", data[0])
	}

	// 提取字段
	seq := data[1]
	devType := data[2]
	usage := data[3]
	pressed := data[4]
	modifier := data[5]
	crcReceived := data[6]

	// CRC8 校验（计算前 6 字节）
	crcCalculated := crc8(data[0:6])
	if crcReceived != crcCalculated {
		return nil, fmt.Errorf("CRC mismatch: received 0x%02X, calculated 0x%02X", crcReceived, crcCalculated)
	}

	// 检测序列号跳跃（可选：用于 reorder 检测）
	if r.lastSeq != 0 && seq != r.lastSeq+1 {
		if seq == 0 && r.lastSeq == 255 {
			// 正常的 255 -> 0 回绕
		} else {
			fmt.Printf("[HID] Warning: sequence jump detected (last=%d, current=%d)\n", r.lastSeq, seq)
		}
	}
	r.lastSeq = seq

	// 构造事件
	event := &HidEvent{
		Seq:      seq,
		Type:     devType,
		Usage:    usage,
		Pressed:  pressed == 0x01,
		Modifier: modifier,
	}

	return event, nil
}

// crc8 计算 CRC-8 校验和
// 多项式：0x07 (x^8 + x^7 + x^2 + x + 1)
// 初始值：0x00
// 与 C++ 端 crc8() 实现严格一致
func crc8(data []byte) uint8 {
	crc := uint8(0x00)
	for _, b := range data {
		crc ^= b
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x07
			} else {
				crc = crc << 1
			}
		}
	}
	return crc
}

// HandleHidEvent 处理解析后的 HID 事件
func HandleHidEvent(event *HidEvent) {
	action := "released"
	if event.Pressed {
		action = "pressed"
	}

	keyName := GetKeyName(event.Usage)
	modifierText := formatModifiers(event.Modifier)

	fmt.Printf("[HID] seq=%3d type=0x%02X usage=0x%02X (%s) %s modifier=0x%02X (%s)\n",
		event.Seq, event.Type, event.Usage, keyName, action, event.Modifier, modifierText)
}

// GetKeyName 将 HID usage code 转换为按键名称
// 参考：USB HID Usage Tables - Keyboard/Keypad Page (0x07)
func GetKeyName(usage uint8) string {
	switch usage {
	case 0x00:
		return "NO_KEY"
	case 0x04:
		return "A"
	case 0x05:
		return "B"
	case 0x06:
		return "C"
	case 0x07:
		return "D"
	case 0x08:
		return "E"
	case 0x09:
		return "F"
	case 0x0A:
		return "G"
	case 0x0B:
		return "H"
	case 0x0C:
		return "I"
	case 0x0D:
		return "J"
	case 0x0E:
		return "K"
	case 0x0F:
		return "L"
	case 0x10:
		return "M"
	case 0x11:
		return "N"
	case 0x12:
		return "O"
	case 0x13:
		return "P"
	case 0x14:
		return "Q"
	case 0x15:
		return "R"
	case 0x16:
		return "S"
	case 0x17:
		return "T"
	case 0x18:
		return "U"
	case 0x19:
		return "V"
	case 0x1A:
		return "W"
	case 0x1B:
		return "X"
	case 0x1C:
		return "Y"
	case 0x1D:
		return "Z"
	case 0x1E:
		return "1"
	case 0x1F:
		return "2"
	case 0x20:
		return "3"
	case 0x21:
		return "4"
	case 0x22:
		return "5"
	case 0x23:
		return "6"
	case 0x24:
		return "7"
	case 0x25:
		return "8"
	case 0x26:
		return "9"
	case 0x27:
		return "0"
	case 0x28:
		return "ENTER"
	case 0x29:
		return "ESC"
	case 0x2A:
		return "BACKSPACE"
	case 0x2B:
		return "TAB"
	case 0x2C:
		return "SPACE"
	case 0x2D:
		return "MINUS"
	case 0x2E:
		return "EQUAL"
	case 0x2F:
		return "LBRACKET"
	case 0x30:
		return "RBRACKET"
	case 0x31:
		return "BACKSLASH"
	case 0x33:
		return "SEMICOLON"
	case 0x34:
		return "APOSTROPHE"
	case 0x35:
		return "GRAVE"
	case 0x36:
		return "COMMA"
	case 0x37:
		return "PERIOD"
	case 0x38:
		return "SLASH"
	case 0x39:
		return "CAPS_LOCK"
	case 0x3A:
		return "F1"
	case 0x3B:
		return "F2"
	case 0x3C:
		return "F3"
	case 0x3D:
		return "F4"
	case 0x3E:
		return "F5"
	case 0x3F:
		return "F6"
	case 0x40:
		return "F7"
	case 0x41:
		return "F8"
	case 0x42:
		return "F9"
	case 0x43:
		return "F10"
	case 0x44:
		return "F11"
	case 0x45:
		return "F12"
	case 0x46:
		return "PRINT_SCREEN"
	case 0x47:
		return "SCROLL_LOCK"
	case 0x48:
		return "PAUSE"
	case 0x49:
		return "INSERT"
	case 0x4A:
		return "HOME"
	case 0x4B:
		return "PAGE_UP"
	case 0x4C:
		return "DELETE"
	case 0x4D:
		return "END"
	case 0x4E:
		return "PAGE_DOWN"
	case 0x4F:
		return "RIGHT_ARROW"
	case 0x50:
		return "LEFT_ARROW"
	case 0x51:
		return "DOWN_ARROW"
	case 0x52:
		return "UP_ARROW"
	case 0x53:
		return "NUM_LOCK"
	case 0x54:
		return "KP_SLASH"
	case 0x55:
		return "KP_ASTERISK"
	case 0x56:
		return "KP_MINUS"
	case 0x57:
		return "KP_PLUS"
	case 0x58:
		return "KP_ENTER"
	case 0x59:
		return "KP_1"
	case 0x5A:
		return "KP_2"
	case 0x5B:
		return "KP_3"
	case 0x5C:
		return "KP_4"
	case 0x5D:
		return "KP_5"
	case 0x5E:
		return "KP_6"
	case 0x5F:
		return "KP_7"
	case 0x60:
		return "KP_8"
	case 0x61:
		return "KP_9"
	case 0x62:
		return "KP_0"
	case 0x63:
		return "KP_PERIOD"
	case 0xE0:
		return "L_CTRL_EXT"
	case 0xE1:
		return "L_SHIFT_EXT"
	case 0xE2:
		return "L_ALT_EXT"
	case 0xE3:
		return "L_GUI_EXT"
	case 0xE4:
		return "R_CTRL_EXT"
	case 0xE5:
		return "R_SHIFT_EXT"
	case 0xE6:
		return "R_ALT_EXT"
	case 0xE7:
		return "R_GUI_EXT"
	default:
		return fmt.Sprintf("UNKNOWN(0x%02X)", usage)
	}
}

// HandleRawHidFrame 处理原始 0x71 帧（从 MessageRouter 调用）
// 输入：已去除 0x57 0xAB 头部的完整数据
func HandleRawHidFrame(data []byte) {
	receiver := NewHidEventReceiver()
	event, err := receiver.ParseHidFrame(data)
	if err != nil {
		fmt.Printf("[HID] Parse error: %v, raw: %s\n", err, hex.EncodeToString(data))
		return
	}
	HandleHidEvent(event)
}

// ParseBinaryFrame 解析二进制帧（包含 0x57 0xAB 头部）
// 支持从一个 UDP 包中提取多个帧
func ParseBinaryFrame(data []byte) {
	for len(data) >= 3 {
		if data[0] != 0x57 || data[1] != 0xAB {
			fmt.Printf("[UDP] Invalid frame header: 0x%02X%02X (expected 0x57AB)\n", data[0], data[1])
			break
		}

		switch data[2] {
	case 0x71:
		// 0x71 设备事件帧（Pico USB Host → UART 桥接）
		// 帧格式：[0x57][0xAB][0x71][dev_addr][mounted][vid_h][vid_l][pid_h][pid_l]
		//        [bcd_usb_h][bcd_usb_l][b_device_class][b_device_subclass][b_device_protocol][b_max_packet_size0]
		//        [bcd_device_h][bcd_device_l]
		//        [b_num_interfaces][b_configuration_value][bm_attributes][b_max_power]
		//        [itf_num][b_interface_class][b_interface_subclass][itf_protocol][b_interval][instance]
		//        [mfg_len][mfg_data(16)][prod_len][prod_data(16)][serial_len][serial_data(16)]
		//        [checksum]（79字节）
		if len(data) >= 79 {
			devAddr := data[3]
			mounted := data[4] == 0x01
			vid := uint16(data[5])<<8 | uint16(data[6])
			pid := uint16(data[7])<<8 | uint16(data[8])
			bcdUSB := uint16(data[9])<<8 | uint16(data[10])
			bDeviceClass := data[11]
			bDeviceSubclass := data[12]
			bDeviceProtocol := data[13]
			bMaxPacketSize0 := data[14]
			bcdDevice := uint16(data[15])<<8 | uint16(data[16])
			bNumInterfaces := data[17]
			bConfigurationValue := data[18]
			bmAttributes := data[19]
			bMaxPower := data[20]
			itfNum := data[21]
			bInterfaceClass := data[22]
			bInterfaceSubclass := data[23]
			itfProtocol := data[24]
			bInterval := data[25]
			instance := data[26]
			
			// 字符串描述符
			mfgLen := data[27]
			var manufacturer string
			if mfgLen > 0 {
				endIdx := 28 + int(mfgLen)
				if endIdx > 44 {
					endIdx = 44
				}
				manufacturer = decodeUTF16LE(data[28:endIdx])
			}
			
			prodLen := data[44]
			var product string
			if prodLen > 0 {
				endIdx := 45 + int(prodLen)
				if endIdx > 61 {
					endIdx = 61
				}
				product = decodeUTF16LE(data[45:endIdx])
			}
			
			serialLen := data[61]
			var serial string
			if serialLen > 0 {
				endIdx := 62 + int(serialLen)
				if endIdx > 78 {
					endIdx = 78
				}
				serial = decodeUTF16LE(data[62:endIdx])
			}
			
			// XOR 校验
			xorSum := byte(0)
			for i := 0; i < 78; i++ {
				xorSum ^= data[i]
			}
			if xorSum != data[78] {
				fmt.Printf("[UDP] 0x71 checksum failed: calc=0x%02X recv=0x%02X raw=%s\n",
					xorSum, data[78], hex.EncodeToString(data[:79]))
				data = data[79:]
				continue
			}
			
			status := "UNMOUNTED"
			if mounted {
				status = "MOUNTED  "
			}
			
			deviceType := "UNKNOWN"
			if itfProtocol == 1 {
				deviceType = "KEYBOARD"
			} else if itfProtocol == 2 {
				deviceType = "MOUSE"
			}
			
			fmt.Printf("[DEV] %s addr=%d VID=0x%04X PID=0x%04X type=%s\n",
				status, devAddr, vid, pid, deviceType)
			fmt.Printf("      USB=%04X class=%02X/%02X/%02X maxpkt0=%d version=%04X\n",
				bcdUSB, bDeviceClass, bDeviceSubclass, bDeviceProtocol, bMaxPacketSize0, bcdDevice)
			fmt.Printf("      config: numItf=%d cfgVal=%d attr=%02X power=%dmA\n",
				bNumInterfaces, bConfigurationValue, bmAttributes, bMaxPower*2)
			fmt.Printf("      interface: num=%d class=%02X subclass=%02X protocol=%d interval=%dms instance=%d\n",
				itfNum, bInterfaceClass, bInterfaceSubclass, itfProtocol, bInterval, instance)
			if manufacturer != "" {
				fmt.Printf("      manufacturer: %s\n", manufacturer)
			}
			if product != "" {
				fmt.Printf("      product: %s\n", product)
			}
			if serial != "" {
				fmt.Printf("      serial: %s\n", serial)
			}
			data = data[79:]
		} else {
			fmt.Printf("[UDP] 0x71 frame too short: %s\n", hex.EncodeToString(data))
			break
		}
	case 0x77:
		// 0x77 自定义键盘事件帧（Pico USB Host → UART 桥接）
		// 帧格式：[0x57][0xAB][0x77][usage][pressed][modifiers][checksum]（7字节）
		if len(data) >= 7 {
			usage := data[3]
			pressed := data[4] == 0x01
			modifiers := data[5]
			
			// XOR 校验
			xorSum := byte(0)
			for i := 0; i < 6; i++ {
				xorSum ^= data[i]
			}
			if xorSum != data[6] {
				fmt.Printf("[UDP] 0x77 checksum failed: calc=0x%02X recv=0x%02X raw=%s\n",
					xorSum, data[6], hex.EncodeToString(data[:7]))
				data = data[7:]
				continue
			}
			
			action := "RELEASED"
			if pressed {
				action = "PRESSED "
			}
			keyName := GetKeyName(usage)
			fmt.Printf("[KEY] usage=0x%02X (%s) %s modifiers=0x%02X (%s)\n",
				usage, keyName, action, modifiers, formatModifiers(modifiers))
			data = data[7:]
		} else {
			fmt.Printf("[UDP] 0x77 frame too short: %s\n", hex.EncodeToString(data))
			break
		}
	default:
			if len(data) < 4 {
				fmt.Printf("[UDP] Frame too short for TYPE: %s\n", hex.EncodeToString(data))
				break
			}
			length := int(data[2])
			if len(data) < length {
				fmt.Printf("[UDP] Frame truncated: expected %d bytes, got %d: %s\n", length, len(data), hex.EncodeToString(data))
				break
			}
			parseBinaryEncoderFrame(data[:length])
			data = data[length:]
		}
	}
}

// parseBinaryEncoderFrame 解析二进制编码器帧（LEN+TYPE 格式）
func parseBinaryEncoderFrame(data []byte) {
	length := data[2]
	frameType := data[3]

	switch frameType {
	case 0x01:
		// 0x01 Keyboard event frame
		// 帧格式：[0x57][0xAB][0x08][0x01][usage][pressed][modifiers][checksum]
		if len(data) >= int(length) {
			usage := data[4]
			pressed := data[5] == 0x01
			modifiers := data[6]
			action := "RELEASED"
			if pressed {
				action = "PRESSED "
			}
			keyName := GetKeyName(usage)
			fmt.Printf("[KEY] usage=0x%02X (%s) %s modifiers=0x%02X (%s)\n", 
				usage, keyName, action, modifiers, formatModifiers(modifiers))
		} else {
			fmt.Printf("[UDP] Keyboard frame too short: %s\n", hex.EncodeToString(data))
		}
	case 0x02:
		// 0x02 PING frame
		// 帧格式：[0x57][0xAB][0x06][0x02][payload][checksum]
		if len(data) >= int(length) {
			payload := data[4]
			fmt.Printf("[PING] PING frame, payload=0x%02X\n", payload)
		}
	case 0x03:
		// 0x03 PONG frame
		// 帧格式：[0x57][0xAB][0x06][0x03][payload][checksum]
		if len(data) >= int(length) {
			payload := data[4]
			fmt.Printf("[PING] PONG frame, payload=0x%02X\n", payload)
		}
	case 0x04:
		// 0x04 Device mount
		if len(data) >= int(length) {
			devAddr := data[4]
			fmt.Printf("[USB] Device mounted: dev_addr=%d\n", devAddr)
		}
	case 0x05:
		// 0x05 Device unmount
		if len(data) >= int(length) {
			devAddr := data[4]
			fmt.Printf("[USB] Device unmounted: dev_addr=%d\n", devAddr)
		}
	case 0x06:
		// 0x06 Device info
		// 帧格式：[0x57][0xAB][0x0B][0x06][dev_addr][vid_low][vid_high][pid_low][pid_high][bInterval][itf_num][itf_protocol][instance][checksum]
		if len(data) >= int(length) {
			devAddr := data[4]
			vid := uint16(data[5]) | uint16(data[6])<<8
			pid := uint16(data[7]) | uint16(data[8])<<8
			bInterval := data[9]
			itfNum := data[10]
			itfProtocol := data[11]
			instance := data[12]
			
			protocolName := "Unknown"
			switch itfProtocol {
			case 0x00:
				protocolName = "None"
			case 0x01:
				protocolName = "Keyboard"
			case 0x02:
				protocolName = "Mouse"
			}
			
			fmt.Printf("[USB] Device info: dev_addr=%d VID=0x%04X PID=0x%04X bInterval=%dms itf=%d protocol=%s instance=%d\n",
				devAddr, vid, pid, bInterval, itfNum, protocolName, instance)
		} else {
			fmt.Printf("[UDP] Device info frame too short: %s\n", hex.EncodeToString(data))
		}
	default:
		fmt.Printf("[UDP] Unknown binary frame type 0x%02X: %s\n", frameType, hex.EncodeToString(data))
	}
}

// formatModifiers 将修饰键位图格式化为可读字符串
func formatModifiers(modifiers uint8) string {
	var result []string
	if modifiers&0x01 != 0 {
		result = append(result, "L_CTRL")
	}
	if modifiers&0x02 != 0 {
		result = append(result, "L_SHIFT")
	}
	if modifiers&0x04 != 0 {
		result = append(result, "L_ALT")
	}
	if modifiers&0x08 != 0 {
		result = append(result, "L_GUI")
	}
	if modifiers&0x10 != 0 {
		result = append(result, "R_CTRL")
	}
	if modifiers&0x20 != 0 {
		result = append(result, "R_SHIFT")
	}
	if modifiers&0x40 != 0 {
		result = append(result, "R_ALT")
	}
	if modifiers&0x80 != 0 {
		result = append(result, "R_GUI")
	}
	if len(result) == 0 {
		return "none"
	}
	return strings.Join(result, "|")
}

// decodeUTF16LE 将 UTF-16LE 编码的字节数组解码为 Go 字符串
func decodeUTF16LE(data []byte) string {
	if len(data) == 0 || len(data)%2 != 0 {
		return ""
	}
	runes := make([]rune, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		codePoint := uint16(data[i]) | uint16(data[i+1])<<8
		runes = append(runes, rune(codePoint))
	}
	return string(runes)
}
