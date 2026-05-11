package qr 

import (
    "fmt"
	"math/rand"
    "github.com/skip2/go-qrcode"
	"net"
)

func QRGen() (int, error) {
	localIP, _ := getOutboundIP()
	token := rand.Intn(1000)
	dynamicURL := fmt.Sprintf("http://%s:8080/qr/%d", localIP, token)

    // Generate a 256x256 PNG
    // Medium error recovery (15%) is usually the sweet spot
    err := qrcode.WriteFile(dynamicURL, qrcode.Medium, 256, "dynamic_qr.png")
    
    if err != nil {
		return 0, err
    }

	return token, nil
}

func getOutboundIP() (string, error) {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        return "", nil
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String(), nil
}
