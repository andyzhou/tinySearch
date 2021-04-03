package tc


import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"sync"
	"time"
)

/*
 * SSH scp service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//internal macro define
const (
	CertPassword = 1
	CertPubKeyFile = 2
	ConnectTimeOut = 10
	ConnectCheckRate = 30
	ScpFileChanSize = 64
	ScpFilePermission = "0666"
)

//ssh config
type SSHConfig struct {
	Mode int //1,2
	Address string
	User string
	CertPath string
	RemoteRootPath string
}

//ssh service info
type SSHService struct {
	switcher bool
	conf *SSHConfig
	mode int `auth mode`
	address string `remote host:ip`
	user string `ssh account`
	cert string `ssh password or key file path`
	remoteRoot string `remote server root path`
	session *ssh.Session
	client *ssh.Client
	scpFileChan chan string `scp file chan`
	closeChan chan bool
	sync.Mutex
}


//construct
func NewSSHService(conf *SSHConfig) *SSHService {
	//init self
	this := &SSHService{
		switcher:false,
		conf:conf,
		session:nil,
		client:nil,
		scpFileChan:make(chan string, ScpFileChanSize),
		closeChan:make(chan bool),
	}

	//set config value
	this.mode = conf.Mode
	this.address = conf.Address
	this.user = conf.User
	this.cert = conf.CertPath
	this.remoteRoot = conf.RemoteRootPath

	//try init connect
	go this.connectServer(this.mode)

	//spawn main process
	go this.runMainProcess()

	return this
}

/////////
//API
////////

//quit
func (s *SSHService) Quit() {
	s.closeChan <- true
	time.Sleep(time.Second/10)
}

//copy file
func (s *SSHService) CopyFile(filePath string) bool {
	if filePath == "" || !s.switcher {
		return false
	}

	//try catch panic of send data to closed chan
	defer func() {
		if err := recover(); err != nil {
			log.Println("SSHService::CopyFile, panic happened, err:", err)
		}
	}()

	//cast to file chan
	s.scpFileChan <- filePath

	return true
}

//connect server with key
func (s *SSHService) ConnectWithKey() bool {
	return s.connectServer(CertPubKeyFile)
}

//connect server with password
func (s *SSHService) ConnectWithPassword() bool {
	return s.connectServer(CertPassword)
}

//set basic auth
func (s *SSHService) SetBasicAuth(host string, port int, user, cert string) bool {
	s.address = fmt.Sprintf("%s:%d", host, port)
	s.user = user
	s.cert = cert
	return true
}

//set remote root path
func (s *SSHService) SetRemotePath(path string) bool {
	if path == "" {
		return false
	}
	s.remoteRoot = path
	return true
}

////////////////
//private func
///////////////

//run main process
func (s *SSHService) runMainProcess() {
	var (
		filePath string
		needQuit bool
		ticker = time.Tick(time.Second * ConnectCheckRate)
	)
	for {
		if needQuit && len(s.scpFileChan) <= 0 {
			break
		}
		select {
		case <- ticker:
			//check connection
			s.checkConnect()
		case filePath = <- s.scpFileChan:
			//read and copy file
			s.readAndCopyFile(filePath)
		case <- s.closeChan:
			needQuit = true
		}
	}
	//do some clean up
	s.cleanUp()
	log.Println("SSHService::runMainProcess, need quit...")
}

//clean up
func (s *SSHService) cleanUp() {
	if s.client != nil {
		s.client.Close()
	}
	if s.session != nil {
		s.session.Close()
	}
}

//read and copy file
func (s *SSHService) readAndCopyFile(filePath string) bool {
	//check client is ok
	if s.client == nil {
		//s.scpFileChan <- filePath
		log.Println("SSHService::readAndCopyFile, client still not inited.")
		return false
	}

	//try open source file
	f, err := os.Open(filePath)
	if err != nil {
		log.Println("SSHService::readAndCopyFile, open file:", filePath, " failed, err:", err.Error())
		return false
	}
	defer f.Close()

	//get file status
	stat, _ := f.Stat()

	//init remote path
	fileName := path.Base(filePath)
	remoteFullPath := fmt.Sprintf("%s/%s", s.remoteRoot, fileName)
	log.Println("SSHService::readAndCopyFile, fileName:", fileName, ", remoteFullPath:", remoteFullPath)

	//begin copy
	err = s.copyFile(f, remoteFullPath, ScpFilePermission, stat.Size())
	if err != nil {
		log.Println("SSHService::readAndCopyFile, copy file failed, err:", err.Error())
		return false
	}

	//need remove original file
	err = os.Remove(filePath)
	if err != nil {
		log.Println("SSHService::readAndCopyFile, remove file:", filePath, " failed, err:", err.Error())
	}

	return true
}

//scp file to remote server
func (s *SSHService) copyFile(r io.Reader, remotePath string, permissions string, size int64) error {
	filename := path.Base(remotePath)
	directory := path.Dir(remotePath)
	log.Println("SSHService::copyFile, filename:", filename, ", sizez:", size, ", directory:", directory)

	//init ssh client session
	s.Lock()
	defer s.Unlock()
	if s.client == nil {
		return errors.New("ssh client is not init yet")
	}

	//init new session
	session, err := s.client.NewSession()
	if err != nil {
		log.Println("Init ssh client session failed, err:", err.Error())
		return err
	}

	defer session.Close()
	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	//init and start command
	cmd := "/usr/bin/scp -t " + directory
	if err := session.Start(cmd); err != nil {
		w.Close()
		return err
	}
	errors := make(chan error)
	go func() {
		errors <- session.Wait()
	}()

	//begin copy original data
	fmt.Fprintln(w, "C"+permissions, size, filename)
	io.Copy(w, r)
	fmt.Fprint(w, "\x00")
	w.Close()

	err = <- errors
	log.Println("SSHService::copyFile finished, err:", err)

	return err
}

//check connect
func (s *SSHService) checkConnect() bool {
	if s.client != nil {
		return false
	}
	go s.connectServer(s.mode)
	return true
}

//connect server
func (s *SSHService) connectServer(mode int) bool {
	var sshConfig *ssh.ClientConfig
	var auth  []ssh.AuthMethod

	if !s.switcher {
		//log.Println("SSHService::connectServer, mode:", mode, ", switcher is closed.")
		return false
	}

	switch mode {
	case CertPassword:
		auth = []ssh.AuthMethod{ssh.Password(s.cert)}
	case CertPubKeyFile:
		auth = []ssh.AuthMethod{s.readPublicKeyFile(s.cert)}
	default:
		log.Println("Un supported mode", mode)
		return false
	}

	//init host key call back
	cb := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}

	//begin init ssh client config
	sshConfig = &ssh.ClientConfig{
		User: s.user,
		Auth: auth,
		//if not need auth at server side, cb should be nil.
		HostKeyCallback:cb,
		Timeout:time.Second * ConnectTimeOut,
	}

	//try connect server
	client, err := ssh.Dial("tcp", s.address, sshConfig)
	if err != nil {
		log.Println("Connect ssh server ", s.address, " failed, err:", err.Error())
		return false
	}

	//connect success, sync client with locker
	s.Lock()
	s.client = client
	s.Unlock()

	log.Println("Connect ssh sever ", s.address, " success")

	return true
}

//read pub key file
func (s *SSHService) readPublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("SSHService::readPublicKeyFile, read key file:", file, " failed, err:", err.Error())
		return nil
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		log.Println("SSHService::readPublicKeyFile, parse private key failed.")
		return nil
	}
	return ssh.PublicKeys(key)
}
