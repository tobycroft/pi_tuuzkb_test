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
	// 打印调试日志
	action := "released"
	if event.Pressed {
		action = "pressed"
	}

	fmt.Printf("[HID] seq=%3d type=0x%02X usage=0x%02X %s modifier=0x%02X\n",
		event.Seq, event.Type, event.Usage, action, event.Modifier)
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
// 根据第四个字节（TYPE）区分协议类型
func ParseBinaryFrame(data []byte) {
	if len(data) < 3 {
		fmt.Printf("[UDP] Frame too short: %s\n", hex.EncodeToString(data))
		return
	}

	// 检查帧头
	if data[0] != 0x57 || data[1] != 0xAB {
		fmt.Printf("[UDP] Invalid frame header: 0x%02X%02X (expected 0x57AB)\n", data[0], data[1])
		return
	}

	// 根据第四个字节（TYPE）区分协议类型
	if len(data) < 4 {
		fmt.Printf("[UDP] Frame too short for TYPE: %s\n", hex.EncodeToString(data))
		return
	}

	switch data[3] {
	case 0x01:
		// 0x01 Keyboard event frame
		// 帧格式：[0x57][0xAB][0x08][0x01][usage][pressed][modifiers][checksum]
		if len(data) >= 8 {
			usage := data[4]
			pressed := data[5] == 0x01
			modifiers := data[6]
			action := "RELEASED"
			if pressed {
				action = "PRESSED "
			}
			fmt.Printf("[KEY] usage=0x%02X %s modifiers=0x%02X (%s)\n", 
				usage, action, modifiers, formatModifiers(modifiers))
		} else {
			fmt.Printf("[UDP] Keyboard frame too short: %s\n", hex.EncodeToString(data))
		}
	case 0x04:
		// 0x04 Device mount
		if len(data) >= 6 {
			devAddr := data[4]
			fmt.Printf("[USB] Device mounted: dev_addr=%d\n", devAddr)
		}
	case 0x05:
		// 0x05 Device unmount
		if len(data) >= 6 {
			devAddr := data[4]
			fmt.Printf("[USB] Device unmounted: dev_addr=%d\n", devAddr)
		}
	case 0x06:
		// 0x06 Device info
		// 帧格式：[0x57][0xAB][0x0B][0x06][dev_addr][vid_low][vid_high][pid_low][pid_high][bInterval][itf_num][itf_protocol][instance][checksum]
		if len(data) >= 14 {
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
	case 0x81:
		fmt.Printf("[UDP] 0x81 Device state frame: %s\n", hex.EncodeToString(data))
	case 0x71:
		// 0x71 HID 输入事件帧（旧协议，用于兼容）
		HandleRawHidFrame(data[2:])
	default:
		fmt.Printf("[UDP] Unknown frame type 0x%02X: %s\n", data[3], hex.EncodeToString(data))
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
