package coupler

import (
	"bytes"
	context2 "context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"shbl/fort"
	"time"
)

type Manager struct {
	Down1      chan []byte
	Up2        chan []byte
	IP         string
	Port       string
	SSHConfig  *ssh.ClientConfig
	TermWidth  int
	TermHeight int
}

func (m *Manager) StartSSH(ctx context2.Context) error {
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", m.IP, m.Port), m.SSHConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	s := fort.SSHTerminal{
		Session: session,
	}

	defer func() {
		if s.ExitMsg == "" {
			m.Up2 <- []byte(fmt.Sprintf("the connection was closed on the remote side on %s", time.Now().Format(time.RFC822)))
		} else {
			m.Up2 <- []byte(s.ExitMsg)
		}
	}()

	termType := "xterm-256color"
	err = s.Session.RequestPty(termType, m.TermHeight, m.TermWidth, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	s.Stdin, err = s.Session.StdinPipe()
	if err != nil {
		return err
	}

	s.Stdout, err = s.Session.StdoutPipe()
	if err != nil {
		return err
	}

	s.Stderr, err = s.Session.StderrPipe()
	if err != nil {
		return err
	}

	ctxIn, cancelIn := context2.WithCancel(context2.Background())
	go func() {
		for {
			select {
			case dat := <-m.Down1:
				s.Stdin.Write(dat)
			case <-ctxIn.Done():
				fmt.Println("stdin exit.")
				return
			}
		}
	}()

	ctxOut, cancelOut := context2.WithCancel(context2.Background())
	go func() {
		var outByte = make([]byte, 1)
		var outBuf = bytes.NewBuffer(outByte)
		for {
			select {
			case <-ctxOut.Done():
				fmt.Println("stdout exit.")
				return
			default:
				io.CopyN(outBuf, s.Stdout, 1)
				dat := outBuf.Bytes()
				m.Up2 <- dat
				fmt.Printf(string(dat[:]))
				outBuf.Reset()
			}
		}
	}()

	ctxErr, cancelErr := context2.WithCancel(context2.Background())
	go func() {
		for {
			var errByte = make([]byte, 1)
			var errBuf = bytes.NewBuffer(errByte)
			select {
			case <-ctxErr.Done():
				fmt.Println("stderr exit.")
				return
			default:
				io.CopyN(errBuf, s.Stderr, 1)
				dat := errBuf.Bytes()
				m.Up2 <- dat
				fmt.Println(string(errByte))
				errBuf.Reset()
			}
		}
	}()

	fmt.Println("aaaaaaaaaaa")

	errCh := make(chan error)
	go func() {
		err = s.Session.Shell()
		if err != nil {
			errCh <- err
		}
		err = s.Session.Wait()
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Println("ssh done")
	case err = <-errCh:
		fmt.Println("ssh err: ", err)
	}

	cancelIn()
	cancelOut()
	cancelErr()
	fmt.Println("----------coupler end----------")
	return nil
}
