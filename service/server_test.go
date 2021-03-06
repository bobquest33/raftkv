package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/icexin/raftkv/client"
	"github.com/icexin/raftkv/config"
)

var (
	port = 10000
)

func getAddr() string {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	port++
	return addr
}

func newServerConfig(t *testing.T) (cfg *config.Config, baseDir string, addr string) {
	baseDir, err := ioutil.TempDir(os.TempDir(), "raftkv")
	if err != nil {
		t.Fatal(err)
	}
	addr = getAddr()
	cfg = &config.Config{
		Raft: config.Raft{
			Advertise:         addr,
			DataDir:           filepath.Join(baseDir, "raft"),
			SnapshotInterval:  config.Duration(3 * time.Second),
			SnapshotThreshold: 1000,
			EnableSingleNode:  true,
		},
		Server: config.Server{
			Listen: addr,
		},
		DB: config.DB{
			Dir: filepath.Join(baseDir, "db"),
		},
	}
	return
}

func newServer(t *testing.T) (server *Server, addr string, baseDir string) {
	cfg, baseDir, addr := newServerConfig(t)
	server, err := NewServer(cfg)
	if err != nil {
		os.RemoveAll(baseDir)
		t.Fatal(err)
	}
	return
}

func TestServerClose(t *testing.T) {
	server, _, base := newServer(t)
	defer os.RemoveAll(base)
	go server.Serve()
	time.Sleep(time.Second)
	server.Close()
}

func TestReadWrite(t *testing.T) {
	server, addr, base := newServer(t)
	defer os.RemoveAll(base)
	defer server.Close()
	go server.Serve()
	time.Sleep(2 * time.Second)
	cli := raftkv.NewClient([]string{addr}, nil)
	defer cli.Close()

	key := []byte("key")
	value := []byte("value")
	err := cli.Write(key, value)
	if err != nil {
		t.Fatal(err)
	}

	v, err := cli.Read(key)
	if err != nil {
		t.Fatal(err)
	}

	if string(v) != string(value) {
		t.Errorf("%s %s", v, value)
	}
	err = cli.Delete(key)
	if err != nil {
		t.Fatal(err)
	}
}
