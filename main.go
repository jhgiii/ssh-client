package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	terminal "golang.org/x/term"
	"log"
	"os"
	"os/user"
	"strconv"
)

func main() {
	//Flags and Defaults
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to get local User Name: %v", err)
	}
	serverPtr := flag.String("server", "localhost", "Server Hostname or IP Address")
	userPtr := flag.String("user", currentUser.Username, "Username, defaults to OS Username")
	portPtr := flag.Int("p", 22, "Remote port. Default is 22")
	flag.Parse()

	server := *serverPtr + ":" + strconv.Itoa(*portPtr)
	username := *userPtr

	key, err := os.ReadFile("/home/jim/.ssh/id_rsa")
	if err != nil {
		log.Fatalf("unable to read private key: %v ", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v ", err)
	}
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	fmt.Println(server)
	conn, err := ssh.Dial("tcp", server, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	defer conn.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := conn.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	// Set IO
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	//in, _ := session.StdinPipe()

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1, // enable echoing
		ssh.ECHOCTL:       1,
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	fileDescriptor := int(os.Stdin.Fd())
	if terminal.IsTerminal(fileDescriptor) {
		originalState, err := terminal.MakeRaw(fileDescriptor)
		if err != nil {
			log.Fatalf("Error at MakeRaw %v", err)
		}
		defer terminal.Restore(fileDescriptor, originalState)

		termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
		if err != nil {
			log.Fatalf("Error at Setting Tterminal.GetSize: %v", err)
		}
		err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if err != nil {
			log.Fatalf("Error when requesting PTY: %v", err)
		}
	}
	err = session.Shell()
	if err != nil {
		log.Fatalf("Error when building shell: %v", err)
	}
	session.Wait()
}
