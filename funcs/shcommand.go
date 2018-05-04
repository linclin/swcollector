package funcs

import (
	"log"

	"golang.org/x/crypto/ssh"
	"net"
)

func SSHCommand(user, password, ip_port, command string) (result string) {
	PassWd := []ssh.AuthMethod{ssh.Password(password)}
	Conf := ssh.ClientConfig{User: user, Auth: PassWd,HostKeyCallback:
		func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}}
	Client, err := ssh.Dial("tcp", ip_port, &Conf)
	if err != nil {
		log.Println("Connect to", ip_port, "failed,", err.Error())
		return ""
	}

	defer Client.Close()
	session, err := Client.NewSession()
	if err != nil {
		log.Println("Session estabilish", ip_port, "failed,", err.Error())
		return ""
	}

	res, err := session.Output(command)
	if err != nil {
		log.Println("Command on", ip_port, "failed,", err.Error())
		return ""
	}

	return string(res)
}
