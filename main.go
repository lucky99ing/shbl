package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
)

var (
	username = flag.String("username", "root", "username")
	serverIP = flag.String("ip", "127.0.0.1", "IP")
	sshPort  = flag.String("port", "22", "port")
)

func main() {
	flag.Parse()
	privateBytes, err := os.ReadFile("~/.ssh/id_rsa")
	if err != nil {
		panic(err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic(err)
	}
	var hostKey ssh.PublicKey
	cfg := ssh.ClientConfig{
		User:            *username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(private)},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", *serverIP, *sshPort), &cfg)
	if err != nil {
		panic(err)
	}
	_, req, err := client.OpenChannel("ch1", nil)
	if err != nil {
		panic(err)
	}
	d := <-req
	fmt.Println(d.Type)

	// 请求pty：标准输出终端
	//ok, err := ch.SendRequest("pty-req", true, ssh.Marshal(ssh.Request{
	//
	//}))

}
