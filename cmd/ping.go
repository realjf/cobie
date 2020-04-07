package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net"
	"os"
	"strconv"
	"time"
)


const (
	ECHO_REQUEST_HEAD_LEN = 8 // ICMP报头的长度至少8字节，如果报文包含数据部分则大于8字节
	ECHO_REPLY_HEAD_LEN = 20  // 在接收到echo response消息时，前20字节是ip头。后面的内容才是icmp的内容，应该与echo request的内容一致
)

func init() {
	rootCmd.AddCommand(pingCmd)
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "ping a ip address or a host",
	Long: `ping a ip address or a host`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a ip address argument")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ping(args[0])
	},
}

// 基于ICMP协议
func ping(host string) {
	count := 4 // 发送回显请求数
	size := 32 // 发送缓冲区大小，单位：字节
	var timeout int64 = 1000 // 等待每次回复的超时时间（毫秒）
	nerverstop := false

	// 查找dns主机名字
	cname, _ := net.LookupCNAME(host)
	starttime := time.Now()

	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout * 1000 * 1000))
	//每个域名可能对应多个ip，但实际连接时，请求只会转发到某一个上，故需要获取实际连接的远程ip，才能知道实际ping的机器是哪台
	ip := conn.RemoteAddr()
	fmt.Println("正在 Ping " + cname + " [" + ip.String() + "] 具有 32 字节的数据:")

	var seq int16 = 1
	id0, id1 := genidentifier(host)

	// 发送次数
	sendN := 0
	// 成功应答次数
	recvN := 0
	// 失败请求数
	lostN := 0
	// 所有请求中应答时间最短的一个
	shortT := -1
	// 所有请求中应答时间最长的一个
	longT := -1
	// 所有请求应答时间和
	sumT := 0

	for count > 0 || nerverstop {
		sendN++
		// icmp报文长度，报头8字节，数据部分32字节
		var msg []byte = make([]byte, size + ECHO_REQUEST_HEAD_LEN)
		// 第一个字节表示报文类型, 8表示回显请求
		msg[0] = 8
		// ping的请求和应答，
		msg[1] = 0
		// 校验码占2个字节
		msg[2] = 0
		msg[3] = 0
		// id 标识占2个字节
		msg[4], msg[5] = id0, id1
		// 序号占2个字节
		msg[6], msg[7] = gensequence(seq)

		length := size + ECHO_REQUEST_HEAD_LEN
		// 计算校验和
		check := checkSum(msg[0:length])
		// 左乘右除，把二进制位向右移动
		msg[2] = byte(check >> 8)
		msg[3] = byte(check & 255)

		conn, err = net.DialTimeout("ip:icmp", host, time.Duration(timeout * 1000 * 1000))

		fmt.Println("remote ip:", host)

		checkError(err)

		starttime = time.Now()

		conn.SetDeadline(starttime.Add(time.Duration(timeout * 1000 * 1000)))
		// 发送icmp请求，同时进行计时和次数计算
		_, err = conn.Write(msg[0:length])

		var receive []byte = make([]byte, ECHO_REPLY_HEAD_LEN + length)
		n, err := conn.Read(receive)
		_ = n

		var endduration int = int(int64(time.Since(starttime)) / (1000 * 1000))

		sumT += endduration
		time.Sleep(1000 * 1000 * 1000)

		if err != nil || receive[ECHO_REPLY_HEAD_LEN + 4] != msg[4] ||
			receive[ECHO_REPLY_HEAD_LEN + 5] != msg[5] || receive[ECHO_REPLY_HEAD_LEN + 6] != msg[6] ||
			receive[ECHO_REPLY_HEAD_LEN+7] != msg[7] || endduration >= int(timeout) || receive[ECHO_REPLY_HEAD_LEN] == 11 {
			lostN++
			fmt.Println("对" + cname + "[" + host + "]" + "的请求超时")
		} else {
			if shortT == -1 {
				shortT = endduration
			}else if shortT > endduration {
				shortT = endduration
			}
			if longT == -1 {
				longT = endduration
			}else if longT < endduration {
				longT = endduration
			}
			recvN++
			ttl := int(receive[8])
			fmt.Println("来自 " + cname + "[" + host + "]" + " 的回复: 字节=32 时间=" + strconv.Itoa(endduration) + "ms TTL=" + strconv.Itoa(ttl))
		}
		seq++
		count--
	}
	stat(host, sendN, lostN, recvN, shortT, longT, sumT)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func checkSum(msg []byte) uint16 {
	sum := 0

	length := len(msg)
	for i := 0; i < length; i+= 2 {
		sum += int(msg[i]) * 256 + int(msg[i+1])
	}
	if length % 2 == 1 {
		sum += int(msg[length-1]) * 256
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer uint16 = uint16(^sum)
	return answer
}

func gensequence(v int16) (byte, byte) {
	ret1 := byte(v >> 8)
	ret2 := byte(v & 255)
	return ret1, ret2
}

func genidentifier(host string) (byte, byte) {
	return host[0], host[1]
}

func stat(ip string, sendN int, lostN int, recvN int, shortT int, longT int, sumT int) {
	fmt.Println()
	fmt.Println(ip, " 的 Ping 统计信息:")
	fmt.Printf("    数据包: 已发送 = %d，已接收 = %d，丢失 = %d (%d%% 丢失)，\n", sendN, recvN, lostN, int(lostN*100/sendN))
	fmt.Println("往返行程的估计时间(以毫秒为单位):")
	if recvN != 0 {
		fmt.Printf("    最短 = %dms，最长 = %dms，平均 = %dms\n", shortT, longT, sumT/sendN)
	}
}
